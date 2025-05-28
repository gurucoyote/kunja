package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"kunja/adapter/vikunja"
	"kunja/api"
)

/*
prepareServices and listHandler provide a lightweight, MCP-native
implementation of the “list” tool.  They bypass Cobra completely and call
the shared business logic in buildTaskList(), so they do not rely on stdout
redirection and cannot dead-lock.
*/
func prepareServices(ctx context.Context) (context.Context, Services, error) {
	token := viper.GetString("token")
	base := viper.GetString("baseurl")
	if token == "" || base == "" {
		return ctx, Services{}, fmt.Errorf("missing token or baseurl – run `kunja login` first")
	}

	client := api.NewApiClient(base, token)
	adapter := vikunja.New(client)

	svc := Services{
		Auth:    adapter,
		Task:    adapter,
		Project: adapter,
		User:    adapter,
	}
	ctx = context.WithValue(ctx, servicesKey, svc)
	return ctx, svc, nil
}

func listHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argMap, _ := req.Params.Arguments.(map[string]interface{})
	showAll, _ := argMap["all"].(bool)
	verbose, _ := argMap["verbose"].(bool)

	ctx, svc, err := prepareServices(ctx)
	if err != nil {
		return nil, err
	}
	out, err := buildTaskList(ctx, svc, verbose, showAll)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(out), nil
}

var mcpLog string

// buildMCPServer creates an MCP server and registers every eligible Cobra
// command exactly once.  The same builder is reused by the help output and the
// runtime server, so tool metadata is generated in a single place.
func buildMCPServer() *server.MCPServer {
	s := server.NewMCPServer(AppName, Version)

	// Register simple diagnostic tools that are not backed by Cobra.
	registerBuiltinTools(s)

	// ------------------------------------------------------------------
	// Native MCP “list” tool (no Cobra dependency)
	// ------------------------------------------------------------------
	listTool := mcp.NewTool(
		"list",
		mcp.WithDescription("List tasks sorted by urgency; set all=true to include completed tasks."),
		mcp.WithBoolean("all", mcp.Description("include completed tasks")),
		mcp.WithBoolean("verbose", mcp.Description("return raw JSON instead of table")),
	)
	s.AddTool(listTool, listHandler)
	BuiltinTools = append(BuiltinTools, listTool)

	// ------------------------------------------------------------------
	// Native MCP “new” tool (create task)
	// ------------------------------------------------------------------
	newTool := mcp.NewTool(
		"new",
		mcp.WithDescription("Create a new task in a project."),
		mcp.WithString("title", mcp.Required(), mcp.Description("task title")),
		mcp.WithString("due", mcp.Description("due date YYYY-MM-DD")),
		mcp.WithNumber("project", mcp.Required(), mcp.Description("project ID")),
	)
	s.AddTool(newTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argMap, _ := req.Params.Arguments.(map[string]interface{})
		title, _ := argMap["title"].(string)
		due, _ := argMap["due"].(string)
		projFloat, ok := argMap["project"].(float64)
		if !ok {
			return nil, fmt.Errorf("project argument is required")
		}
		projectID := int(projFloat)

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}
		out, err := createTaskSimple(ctx, svc, title, due, projectID)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(out), nil
	})
	BuiltinTools = append(BuiltinTools, newTool)

	// ------------------------------------------------------------------
	// Native MCP “projects” tool (list projects)
	// ------------------------------------------------------------------
	projectsTool := mcp.NewTool(
		"projects",
		mcp.WithDescription("List projects; verbose=true returns raw JSON."),
		mcp.WithBoolean("verbose", mcp.Description("raw JSON output")),
	)
	s.AddTool(projectsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argMap, _ := req.Params.Arguments.(map[string]interface{})
		verbose, _ := argMap["verbose"].(bool)

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}
		out, err := buildProjectList(ctx, svc, verbose)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(out), nil
	})
	BuiltinTools = append(BuiltinTools, projectsTool)

	// ------------------------------------------------------------------
	// Native MCP “done” tool (toggle task done)
	// ------------------------------------------------------------------
	doneTool := mcp.NewTool(
		"done",
		mcp.WithDescription("Toggle the done status of a task."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("task ID")),
	)
	s.AddTool(doneTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argMap, _ := req.Params.Arguments.(map[string]interface{})
		idFloat, ok := argMap["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("id argument is required")
		}
		taskID := int(idFloat)

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}
		out, err := toggleTaskDone(ctx, svc, taskID)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(out), nil
	})
	BuiltinTools = append(BuiltinTools, doneTool)

	return s
}

var mcpCmd = &cobra.Command{
	Use:         "mcp",
	Short:       "Run Kunja as an MCP server over stdio",
	Annotations: map[string]string{"skip_mcp": "true"},
	RunE:        runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVarP(&mcpLog, "log", "l", "kunja-mcp.log", "debug log file")

	// Custom help prints a human-readable catalogue of all MCP tools.
	mcpCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Run Kunja as an MCP server over stdio.")
		fmt.Fprintln(cmd.OutOrStdout(), "Available tools:")

		// Ensure the BuiltinTools slice is populated; this happens when the
		// MCP server is built (registerBuiltinTools is called).  We avoid
		// duplicate entries by only populating when the slice is still empty.
		if len(BuiltinTools) == 0 {
			_ = buildMCPServer()
		}

		// First show built-in diagnostic tools
		for _, t := range BuiltinTools {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  –  %s\n", t.Name, strings.TrimSpace(t.Description))
		}

		// Cobra-based tools are temporarily hidden from the MCP server.
	})
}

// runMCP starts an MCP server that exposes all Cobra commands as tools.
func runMCP(_ *cobra.Command, _ []string) error {
	// optional log file
	f, err := os.OpenFile(mcpLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	// Build the MCP server and register all tools
	s := buildMCPServer()

	// Serve stdin/stdout
	return server.ServeStdio(s)
}

// genericHandler converts MCP parameters to CLI flags and executes the Cobra command.
func genericHandler(c *cobra.Command) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// ---- log incoming JSON request ---------------------------------
		if raw, err := json.Marshal(req); err == nil {
			log.Printf(">> %s\n", raw)
		}

		// Build CLI args from parameters
		var args []string
		var keys []string

		// MCP v0.30.0 stores all arguments in the Arguments map.
		var argMap map[string]interface{}
		if m, ok := req.Params.Arguments.(map[string]interface{}); ok {
			argMap = m
		} else {
			argMap = map[string]interface{}{}
		}

		for k := range argMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := argMap[k]
			switch vv := v.(type) {
			case bool:
				if vv {
					args = append(args, fmt.Sprintf("--%s", k))
				}
			default:
				args = append(args, fmt.Sprintf("--%s=%v", k, vv))
			}
		}

		// Append positional arguments when Arguments is an array (MCP positional args)
		if rawSlice, ok := req.Params.Arguments.([]interface{}); ok {
			for _, a := range rawSlice {
				args = append(args, fmt.Sprint(a))
			}
		}

		// Capture stdout and drain it concurrently to avoid pipe-buffer dead-locks.
		var buf bytes.Buffer
		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Drain the pipe while the command is running.
		done := make(chan struct{})
		go func() {
			io.Copy(&buf, r)
			close(done)
		}()

		// Execute the command
		c.SetArgs(args)
		c.SetContext(ctx)
		execErr := c.Execute()

		// Restore stdout
		w.Close() // closing writer lets the copier finish
		<-done    // wait until everything is copied
		os.Stdout = stdout

		if execErr != nil {
			log.Printf("!! %v\n", execErr)
			return nil, execErr
		}

		// Prepare MCP result
		result := mcp.NewToolResultText(buf.String())

		// ---- log outgoing JSON response --------------------------------
		if raw, err := json.Marshal(result); err == nil {
			log.Printf("<< %s\n", raw)
		}

		return result, nil
	}
}
