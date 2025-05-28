package pkg

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CobraToMcp converts a Cobra command into an MCP tool specification
// and attaches the supplied handler.
func CobraToMcp(cmd *cobra.Command) mcp.Tool {
	var opts []mcp.ToolOption

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Check if this flag is marked “required for MCP” via annotation.
		_, mcpReq := f.Annotations["mcp_required"]

		switch f.Value.Type() {
		case "string":
			if mcpReq {
				opts = append(opts, mcp.WithString(f.Name, mcp.Required(), mcp.Description(f.Usage)))
			} else {
				opts = append(opts, mcp.WithString(f.Name, mcp.Description(f.Usage)))
			}
		case "int":
			if mcpReq {
				opts = append(opts, mcp.WithNumber(f.Name, mcp.Required(), mcp.Description(f.Usage)))
			} else {
				opts = append(opts, mcp.WithNumber(f.Name, mcp.Description(f.Usage)))
			}
		case "bool":
			if mcpReq {
				opts = append(opts, mcp.WithBoolean(f.Name, mcp.Required(), mcp.Description(f.Usage)))
			} else {
				opts = append(opts, mcp.WithBoolean(f.Name, mcp.Description(f.Usage)))
			}
		}
	})

	// Prefer the long description for MCP if available, otherwise fall back to short.
	desc := strings.TrimSpace(cmd.Short)
	if cmd.Long != "" {
		desc = strings.TrimSpace(cmd.Long)
	}

	return mcp.NewTool(
		cmd.Use,
		append(
			[]mcp.ToolOption{
				mcp.WithDescription(desc),
			},
			opts...,
		)...,
	)
}
