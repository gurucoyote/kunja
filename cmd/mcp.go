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
	"strconv"
	"strings"
	"time"

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
	client.SetCredentials(viper.GetString("username"), viper.GetString("password"))
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
		mcp.WithDescription("List tasks sorted by urgency. By default only open tasks are returned; set all=true if you also want to see completed (done) tasks."),
		mcp.WithBoolean("all", mcp.Description("include done tasks – use only if you also want to see completed tasks")),
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
	// Native MCP “now” tool (current date/time)
	// ------------------------------------------------------------------
	nowTool := mcp.NewTool(
		"now",
		mcp.WithDescription("Return the current date and time in RFC 3339 format. Call this tool any time you need to calculate a relative date or time such as 'tomorrow', 'in three days', etc."),
	)
	s.AddTool(nowTool, func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(time.Now().Format(time.RFC3339)), nil
	})
	BuiltinTools = append(BuiltinTools, nowTool)

	// ------------------------------------------------------------------
	// Native MCP “timecalc” tool (strict RFC-3339 date/time arithmetic)
	// ------------------------------------------------------------------
	timecalcTool := mcp.NewTool(
		"timecalc",
		mcp.WithDescription(
			"Perform date/time calculations. " +
				"op=add|sub requires ts (RFC 3339 with timezone, e.g. 2024-07-08T10:00:00Z) and dur (Go duration, e.g. 2h30m). " +
				"op=diff requires ts and ts2 (both RFC 3339) and returns the difference in seconds. " +
				"op=convert requires ts and toTZ (IANA zone name). " +
				"Call this tool any time you need to calculate a relative date or time such as 'tomorrow', 'in three days', etc.",
		),
		mcp.WithString("op", mcp.Required(), mcp.Description("add, sub, diff or convert")),
		mcp.WithString("ts", mcp.Required(), mcp.Description("base timestamp (RFC 3339 with timezone)")),
		mcp.WithString("dur", mcp.Description("duration for add/sub (e.g. 2h30m)")),
		mcp.WithString("ts2", mcp.Description("second timestamp for diff (RFC 3339 with timezone)")),
		mcp.WithString("toTZ", mcp.Description("IANA time-zone for convert (e.g. Europe/Berlin)")),
	)
	s.AddTool(timecalcTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		op := strings.ToLower(fmt.Sprint(args["op"]))
		tsStr := fmt.Sprint(args["ts"])
		if tsStr == "" {
			return nil, fmt.Errorf("ts is required and must be RFC 3339 with timezone")
		}
		base, err := time.Parse(time.RFC3339, tsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ts (must be RFC 3339 with timezone): %w", err)
		}

		switch op {
		case "add", "sub":
			durStr := fmt.Sprint(args["dur"])
			if durStr == "" {
				return nil, fmt.Errorf("dur is required for %s", op)
			}
			d, err := time.ParseDuration(durStr)
			if err != nil {
				return nil, fmt.Errorf("invalid dur: %w", err)
			}
			if op == "sub" {
				d = -d
			}
			return mcp.NewToolResultText(base.Add(d).Format(time.RFC3339)), nil

		case "diff":
			ts2Str := fmt.Sprint(args["ts2"])
			if ts2Str == "" {
				return nil, fmt.Errorf("ts2 is required for diff")
			}
			ts2, err := time.Parse(time.RFC3339, ts2Str)
			if err != nil {
				return nil, fmt.Errorf("invalid ts2 (must be RFC 3339 with timezone): %w", err)
			}
			sec := int(ts2.Sub(base).Seconds())
			return mcp.NewToolResultText(fmt.Sprintf("%d", sec)), nil

		case "convert":
			tzName := fmt.Sprint(args["toTZ"])
			if tzName == "" {
				return nil, fmt.Errorf("toTZ is required for convert")
			}
			loc, err := time.LoadLocation(tzName)
			if err != nil {
				return nil, fmt.Errorf("unknown time-zone: %s", tzName)
			}
			return mcp.NewToolResultText(base.In(loc).Format(time.RFC3339)), nil

		default:
			return nil, fmt.Errorf("unknown op: %s (must be add, sub, diff or convert)", op)
		}
	})
	BuiltinTools = append(BuiltinTools, timecalcTool)

	// ------------------------------------------------------------------
	// Native MCP “createproject” tool (create a new project)
	// ------------------------------------------------------------------
	createProjectTool := mcp.NewTool(
		"createproject",
		mcp.WithDescription("Create a new project."),
		mcp.WithString("title", mcp.Required(), mcp.Description("project title")),
	)
	s.AddTool(createProjectTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argMap, _ := req.Params.Arguments.(map[string]interface{})
		title, _ := argMap["title"].(string)
		if strings.TrimSpace(title) == "" {
			return nil, fmt.Errorf("title argument is required")
		}

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}
		p, err := svc.Project.CreateProject(ctx, api.Project{Title: title})
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("Project created: %d – %s", p.ID, p.Title)), nil
	})
	BuiltinTools = append(BuiltinTools, createProjectTool)

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

	// ------------------------------------------------------------------
	// Native MCP “delete” tool (delete tasks)
	// ------------------------------------------------------------------
	deleteTool := mcp.NewTool(
		"delete",
		mcp.WithDescription("Delete one or more tasks by ID."),
		mcp.WithString("ids", mcp.Required(), mcp.Description("comma-separated list of task IDs (e.g. \"12,34,56\")")),
	)
	s.AddTool(deleteTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		argMap, _ := req.Params.Arguments.(map[string]interface{})
		idsRaw, _ := argMap["ids"].(string)
		idsRaw = strings.TrimSpace(idsRaw)
		if idsRaw == "" {
			return nil, fmt.Errorf("ids argument is required")
		}

		var ids []int
		for _, part := range strings.Split(idsRaw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid task ID: %q", part)
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("no valid task IDs supplied")
		}

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}

		var deleted []string
		var failed []string
		for _, id := range ids {
			if _, err := svc.Task.DeleteTask(ctx, id); err != nil {
				failed = append(failed, fmt.Sprintf("%d (%v)", id, err))
			} else {
				deleted = append(deleted, strconv.Itoa(id))
			}
		}

		var b strings.Builder
		if len(deleted) > 0 {
			fmt.Fprintf(&b, "Deleted: %s\n", strings.Join(deleted, ", "))
		}
		if len(failed) > 0 {
			fmt.Fprintf(&b, "Failed:  %s\n", strings.Join(failed, ", "))
		}
		return mcp.NewToolResultText(b.String()), nil
	})
	BuiltinTools = append(BuiltinTools, deleteTool)

	// ------------------------------------------------------------------
	// Native MCP “edit” tool (update task fields)
	// ------------------------------------------------------------------
	editTool := mcp.NewTool(
		"edit",
		mcp.WithDescription("Edit a task (title, description, due date or project)."),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("task ID")),
		mcp.WithString("title", mcp.Description("new title")),
		mcp.WithString("description", mcp.Description("new description")),
		mcp.WithString("due", mcp.Description("new due date YYYY-MM-DD")),
		mcp.WithNumber("project", mcp.Description("new project ID")),
	)
	s.AddTool(editTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})

		idFloat, ok := args["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("id argument is required")
		}
		title, _ := args["title"].(string)
		desc, _ := args["description"].(string)
		due, _ := args["due"].(string)
		projFloat, _ := args["project"].(float64)

		if title == "" && desc == "" && due == "" && projFloat == 0 {
			return nil, fmt.Errorf("provide at least one of title, description, due or project")
		}

		ctx, svc, err := prepareServices(ctx)
		if err != nil {
			return nil, err
		}
		out, err := editTaskSimple(ctx, svc, int(idFloat), title, desc, due, int(projFloat))
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(out), nil
	})
	BuiltinTools = append(BuiltinTools, editTool)

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
