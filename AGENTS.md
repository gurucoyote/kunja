# Project Overview

Kunja is a Go-based command-line client for the Vikunja task management API, built with Cobra. The project is in the process of evolving into a Model Context Protocol (MCP) server so that its capabilities can be exposed as tools for automation. A legacy Python client lives in `python/` strictly as a reference; it is deprecated and must not receive functional updates.

# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` holds Cobra commands; `cmd/mcp.go` builds the Model Context Protocol (MCP) server interface.
- `internal/core` and `internal/service` implement business logic and service abstractions consumed by adapters.
- `adapter/vikunja` adapts the Vikunja REST API client from `api/` to the service interfaces.
- `python/` contains an older, reference-only client; treat it as read-only context when updating behavior.
- Top-level scripts such as `run-inspector.sh` and `refactorTasks.txt` document MCP experiments and refactors.

## Common CLI Commands
- `./kunja` lists open tasks using the authenticated Vikunja account.
- `./kunja login` authenticates the CLI against the Vikunja API.
- `./kunja project-users <PROJECT_ID>` enumerates users assigned to a project.
- `./kunja mcp` starts the MCP server (useful for Codex or the MCP Inspector); see `run-inspector.sh` for an end-to-end example.

## Build, Test, and Development Commands
- `go build -o kunja .` builds the CLI binary; run from the repo root after dependency updates.
- `GOCACHE=$(mktemp -d) go test ./...` executes Go unit tests without requiring write access to the default build cache.
- `./kunja mcp` launches the MCP server over stdio for integration with Codex or the MCP Inspector.

## Coding Style & Naming Conventions
- Format Go sources with `gofmt`; run `gofmt -w <files>` before committing.
- Favor descriptive function names and keep MCP tool identifiers snake_case (e.g., `time_add`) to satisfy Codex constraints.
- For JSON/http structs, match Vikunja field casing and document quirks inline with concise comments.

## Testing Guidelines
- Go testing uses the standard library’s `testing` package; place tests alongside implementation files with `_test.go` suffixes.
- Prefer table-driven tests for service logic and adapters; stub external APIs via the service interfaces defined under `internal`.
- Run the test command above prior to submitting changes; include reproduction steps for any failing cases in PRs.

## Commit & Pull Request Guidelines
- Write imperative commit subjects under 72 characters (e.g., `Fix MCP time tool names`).
- Group related changes per commit, summarizing scope and impact in the body when necessary.
- PRs should describe the problem, solution, and validation (tests, manual runs). Link related issues and attach screenshots or logs for CLI output when meaningful.

## MCP Integration Tips
- When exposing new tools, register them in `cmd/mcp.go`, ensure argument names are CLI-friendly, and document usage in the help output.
- After additions, verify Codex discovery with `codex mcp list` and a smoke test conversation.
