package cmd

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerBuiltinTools adds simple diagnostic tools that bypass Cobra.
// They are useful to verify that the Kunja MCP server itself works
// even when the Cobra integration fails.
func registerBuiltinTools(s *server.MCPServer) {
	// ---- ping ---------------------------------------------------------
	s.AddTool(mcp.Tool{
		Name:        "ping",
		Description: "Return «pong» – verifies that the MCP server is alive.",
	}, func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("pong"), nil
	})

	// ---- echo ---------------------------------------------------------
	s.AddTool(mcp.Tool{
		Name:        "echo",
		Description: "Echo back the supplied text argument.",
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		text := fmt.Sprint(args["text"])
		return mcp.NewToolResultText(text), nil
	})

	// ---- sum ----------------------------------------------------------
	s.AddTool(mcp.Tool{
		Name:        "sum",
		Description: "Return the sum of two integers.",
	}, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		// JSON numbers arrive as float64
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		sum := int(a) + int(b)
		return mcp.NewToolResultText(fmt.Sprintf("%d", sum)), nil
	})
}
