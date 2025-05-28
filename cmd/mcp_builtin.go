package cmd

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// List of built-in diagnostic tools (exposed e.g. in help)
var BuiltinTools []mcp.Tool

// registerBuiltinTools adds simple diagnostic tools that bypass Cobra.
// They are useful to verify that the Kunja MCP server itself works
// even when the Cobra integration fails.
func registerBuiltinTools(s *server.MCPServer) {
	// ---- ping ---------------------------------------------------------
	pingTool := mcp.Tool{
		Name:        "ping",
		Description: "Return «pong» – verifies that the MCP server is alive.",
		Parameters:  map[string]interface{}{}, // encode as {} not null
	}
	s.AddTool(pingTool, func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("pong"), nil
	})
	BuiltinTools = append(BuiltinTools, pingTool)

	// ---- echo ---------------------------------------------------------
	echoTool := mcp.Tool{
		Name:        "echo",
		Description: "Echo back the supplied text argument.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{"type": "string"},
			},
			"required": []string{"text"},
		},
	}
	s.AddTool(echoTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		text := fmt.Sprint(args["text"])
		return mcp.NewToolResultText(text), nil
	})
	BuiltinTools = append(BuiltinTools, echoTool)

	// ---- sum ----------------------------------------------------------
	sumTool := mcp.Tool{
		Name:        "sum",
		Description: "Return the sum of two integers.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "integer"},
				"b": map[string]interface{}{"type": "integer"},
			},
			"required": []string{"a", "b"},
		},
	}
	s.AddTool(sumTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		// JSON numbers arrive as float64
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		sum := int(a) + int(b)
		return mcp.NewToolResultText(fmt.Sprintf("%d", sum)), nil
	})
	BuiltinTools = append(BuiltinTools, sumTool)
}
