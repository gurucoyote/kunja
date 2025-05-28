package pkg

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CobraToMcp converts a Cobra command into an MCP tool specification
// and attaches the supplied handler.
func CobraToMcp(
	cmd *cobra.Command,
	h func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error),
) *mcp.Tool {
	var opts []mcp.ToolOption

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		switch f.Value.Type() {
		case "string":
			opt := mcp.WithString(f.Name, mcp.Description(f.Usage))
			opts = append(opts, opt)
		case "int":
			opts = append(opts, mcp.WithNumber(f.Name, mcp.Description(f.Usage)))
		case "bool":
			opts = append(opts, mcp.WithBoolean(f.Name, mcp.Description(f.Usage)))
		}
	})

	return mcp.NewTool(
		cmd.Use,
		append([]mcp.ToolOption{mcp.WithDescription(cmd.Short)}, opts...)...,
	).WithHandler(h)
}
