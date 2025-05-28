package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"kunja/pkg"
)

var mcpLog string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run Kunja as an MCP server over stdio",
	RunE:  runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVarP(&mcpLog, "log", "l", "kunja-mcp.log", "debug log file")
}

// runMCP starts an MCP server that exposes all Cobra commands as tools.
func runMCP(_ *cobra.Command, _ []string) error {
	// optional log file
	f, err := os.OpenFile(mcpLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	// Create the MCP server
	s := server.NewMCPServer(AppName, Version)

	// Register all non-hidden sub-commands of rootCmd.
	cmds := rootCmd.Commands()
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })

	for _, c := range cmds {
		if c.Hidden {
			continue
		}
		tool := pkg.CobraToMcp(c)
		s.AddTool(tool, genericHandler(c))
	}

	// Serve stdin/stdout
	return server.ServeStdio(s)
}

// genericHandler converts MCP parameters to CLI flags and executes the Cobra command.
func genericHandler(c *cobra.Command) mcp.ToolHandler {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Build CLI args from parameters
		var args []string
		var keys []string
		for k := range req.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := req.Params[k]
			switch vv := v.(type) {
			case bool:
				if vv {
					args = append(args, fmt.Sprintf("--%s", k))
				}
			default:
				args = append(args, fmt.Sprintf("--%s=%v", k, vv))
			}
		}

		// Capture stdout
		var buf bytes.Buffer
		stdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute the command
		c.SetArgs(args)
		c.SetContext(ctx)
		execErr := c.Execute()

		// Restore stdout
		w.Close()
		io.Copy(&buf, r)
		os.Stdout = stdout

		if execErr != nil {
			return nil, execErr
		}
		return mcp.NewToolResultText(buf.String()), nil
	}
}
