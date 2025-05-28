#!/usr/bin/env sh
# ---------------------------------------------------------------------------
# Run the MCP compliance tests against the Kunja MCP server.
# Requires: Go (for building) and Node.js + npx (for the test-runner).
# ---------------------------------------------------------------------------

set -eu

# Move to repository root (directory containing this script’s parent).
cd "$(dirname "$0")"/..

echo "=== Building kunja …"
go build -o kunja .

echo "=== Running MCP compliance tests …"
# Arguments: <executable> <sub-command that starts MCP>
npx --yes @mark3labs/mcp-cli test ./kunja mcp
