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

	"kunja/pkg"
)

var mcpLog string

// buildMCPServer creates an MCP server and registers every eligible Cobra
// command exactly once.  The same builder is reused by the help output and the
// runtime server, so tool metadata is generated in a single place.
func buildMCPServer() *server.MCPServer {
	s := server.NewMCPServer(AppName, Version)

	// Register simple diagnostic tools that are not backed by Cobra.
	registerBuiltinTools(s)

	cmds := rootCmd.Commands()
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })

	for _, c := range cmds {
		if c.Hidden || c.Annotations["skip_mcp"] == "true" || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		s.AddTool(pkg.CobraToMcp(c), genericHandler(c))
	}
	return s
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run Kunja as an MCP server over stdio",
	Annotations: map[string]string{"skip_mcp": "true"},
	RunE:  runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVarP(&mcpLog, "log", "l", "kunja-mcp.log", "debug log file")

	// Custom help prints a human-readable catalogue of all MCP tools.
	mcpCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Run Kunja as an MCP server over stdio.")
		fmt.Fprintln(cmd.OutOrStdout(), "Available tools:")

		// First show built-in diagnostic tools
		for _, t := range BuiltinTools {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  –  %s\n", t.Name, strings.TrimSpace(t.Description))
		}

		// Then show Cobra-based tools
		cmds := rootCmd.Commands()
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })
		for _, c := range cmds {
			if c.Hidden || c.Annotations["skip_mcp"] == "true" || c.Name() == "help" || c.Name() == "completion" {
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s  –  %s\n", c.Use, strings.TrimSpace(c.Short))
		}
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
		w.Close()  // closing writer lets the copier finish
		<-done     // wait until everything is copied
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
