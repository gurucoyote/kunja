package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	// Native MCP time tools – simpler, forgiving parameters
	// ------------------------------------------------------------------

	// time_add
	addTool := mcp.NewTool(
		"time_add",
		mcp.WithDescription("Add a duration to a timestamp. Defaults to now."),
		mcp.WithString("ts", mcp.Description("RFC3339, YYYY-MM-DD, 'now', or unix seconds/ms")),
		mcp.WithNumber("seconds"),
		mcp.WithNumber("minutes"),
		mcp.WithNumber("hours"),
		mcp.WithNumber("days"),
		mcp.WithString("dur", mcp.Description("Go duration (e.g. 2h30m), ISO-8601 (e.g. P1DT30M) or human (e.g. '2 hours')")),
	)
	s.AddTool(addTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		base, err := parseTS(pickArg(args, "ts"))
		if err != nil {
			return nil, err
		}
		d, err := parseDur(args)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(base.Add(d).Format(time.RFC3339)), nil
	})
	BuiltinTools = append(BuiltinTools, addTool)

	// time_sub
	subTool := mcp.NewTool(
		"time_sub",
		mcp.WithDescription("Subtract a duration from a timestamp. Defaults to now."),
		mcp.WithString("ts", mcp.Description("RFC3339, YYYY-MM-DD, 'now', or unix seconds/ms")),
		mcp.WithNumber("seconds"),
		mcp.WithNumber("minutes"),
		mcp.WithNumber("hours"),
		mcp.WithNumber("days"),
		mcp.WithString("dur", mcp.Description("Go duration (e.g. 2h30m), ISO-8601 (e.g. P1DT30M) or human (e.g. '2 hours')")),
	)
	s.AddTool(subTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		base, err := parseTS(pickArg(args, "ts"))
		if err != nil {
			return nil, err
		}
		d, err := parseDur(args)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(base.Add(-d).Format(time.RFC3339)), nil
	})
	BuiltinTools = append(BuiltinTools, subTool)

	// time_diff
	diffTool := mcp.NewTool(
		"time_diff",
		mcp.WithDescription("Difference between two timestamps. Returns a number (default seconds)."),
		mcp.WithString("ts", mcp.Required(), mcp.Description("RFC3339, YYYY-MM-DD, 'now', or unix seconds/ms")),
		mcp.WithString("ts2", mcp.Required(), mcp.Description("RFC3339, YYYY-MM-DD, or unix epoch")),
		mcp.WithString("unit", mcp.Description("seconds|minutes|hours|days (default: seconds)")),
	)
	s.AddTool(diffTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		t1, err := parseTS(pickArg(args, "ts"))
		if err != nil {
			return nil, err
		}
		t2, err := parseTS(pickArg(args, "ts2"))
		if err != nil {
			return nil, err
		}
		u := strings.ToLower(pickArg(args, "unit", "units"))
		delta := t2.Sub(t1).Seconds()
		switch u {
		case "minutes", "minute", "min", "mins", "m":
			delta /= 60
		case "hours", "hour", "hr", "hrs", "h":
			delta /= 3600
		case "days", "day", "d":
			delta /= 86400
		}
		return mcp.NewToolResultText(fmt.Sprintf("%.0f", delta)), nil
	})
	BuiltinTools = append(BuiltinTools, diffTool)

	// time_convert
	convertTool := mcp.NewTool(
		"time_convert",
		mcp.WithDescription("Convert a timestamp to another time-zone."),
		mcp.WithString("ts", mcp.Required(), mcp.Description("RFC3339, YYYY-MM-DD, or unix epoch")),
		mcp.WithString("toTZ", mcp.Required(), mcp.Description("IANA time-zone, e.g. Europe/Berlin")),
		mcp.WithString("fromTZ", mcp.Description("interpret naive ts in this zone (if needed)")),
	)
	s.AddTool(convertTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		tsStr := pickArg(args, "ts")
		to := pickArg(args, "toTZ", "to_tz", "tz", "timezone")
		from := pickArg(args, "fromTZ", "from_tz", "from")
		t, err := parseTSWithFrom(tsStr, from)
		if err != nil {
			return nil, err
		}
		loc, err := time.LoadLocation(to)
		if err != nil {
			return nil, fmt.Errorf("unknown time-zone: %s", to)
		}
		return mcp.NewToolResultText(t.In(loc).Format(time.RFC3339)), nil
	})
	BuiltinTools = append(BuiltinTools, convertTool)

	// Compatibility wrapper: timecalc
	timecalcTool := mcp.NewTool(
		"timecalc",
		mcp.WithDescription("Compatibility wrapper for time calculations (prefer time_add, time_sub, time_diff, time_convert)."),
		mcp.WithString("op", mcp.Required(), mcp.Description("add|plus|+ / sub|minus|- / diff|delta / convert|tz")),
		mcp.WithString("ts", mcp.Description("base timestamp; defaults to 'now' for add/sub")),
		mcp.WithString("dur", mcp.Description("duration for add/sub")),
		mcp.WithString("ts2", mcp.Description("second timestamp for diff")),
		mcp.WithString("toTZ", mcp.Description("target time-zone for convert")),
		mcp.WithString("fromTZ", mcp.Description("interpret naive ts in this zone")),
		mcp.WithNumber("seconds"),
		mcp.WithNumber("minutes"),
		mcp.WithNumber("hours"),
		mcp.WithNumber("days"),
	)
	s.AddTool(timecalcTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, _ := req.Params.Arguments.(map[string]interface{})
		op := strings.ToLower(pickArg(args, "op"))
		switch op {
		case "add", "plus", "+":
			base, err := parseTS(pickArg(args, "ts"))
			if err != nil {
				return nil, err
			}
			d, err := parseDur(args)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(base.Add(d).Format(time.RFC3339)), nil

		case "sub", "minus", "-":
			base, err := parseTS(pickArg(args, "ts"))
			if err != nil {
				return nil, err
			}
			d, err := parseDur(args)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(base.Add(-d).Format(time.RFC3339)), nil

		case "diff", "delta":
			t1, err := parseTS(pickArg(args, "ts"))
			if err != nil {
				return nil, err
			}
			t2, err := parseTS(pickArg(args, "ts2"))
			if err != nil {
				return nil, err
			}
			u := strings.ToLower(pickArg(args, "unit", "units"))
			delta := t2.Sub(t1).Seconds()
			switch u {
			case "minutes", "minute", "min", "mins", "m":
				delta /= 60
			case "hours", "hour", "hr", "hrs", "h":
				delta /= 3600
			case "days", "day", "d":
				delta /= 86400
			}
			return mcp.NewToolResultText(fmt.Sprintf("%.0f", delta)), nil

		case "convert", "tz":
			tsStr := pickArg(args, "ts")
			to := pickArg(args, "toTZ", "to_tz", "tz", "timezone")
			from := pickArg(args, "fromTZ", "from_tz", "from")
			t, err := parseTSWithFrom(tsStr, from)
			if err != nil {
				return nil, err
			}
			loc, err := time.LoadLocation(to)
			if err != nil {
				return nil, fmt.Errorf("unknown time-zone: %s", to)
			}
			return mcp.NewToolResultText(t.In(loc).Format(time.RFC3339)), nil

		default:
			return nil, fmt.Errorf("unknown op: %s (use add/sub/diff/convert)", op)
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
	defaultLogPath := filepath.Join(defaultConfigDir(), "kunja-mcp.log")
	mcpCmd.Flags().StringVarP(&mcpLog, "log", "l", defaultLogPath, "debug log file")

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
	if strings.TrimSpace(mcpLog) != "" {
		if err := os.MkdirAll(filepath.Dir(mcpLog), 0o755); err == nil {
			if f, err := os.OpenFile(mcpLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
				defer f.Close()
				log.SetOutput(io.MultiWriter(os.Stderr, f))
			}
		}
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

// ---------------------------------------------------------------------
// Helper functions for time tools – forgiving parsing
// ---------------------------------------------------------------------

func pickArg(m map[string]interface{}, names ...string) string {
	for _, n := range names {
		if v, ok := m[n]; ok {
			return fmt.Sprint(v)
		}
	}
	return ""
}

func strArg(m map[string]interface{}, name string) string {
	return fmt.Sprint(m[name])
}

func numArg(m map[string]interface{}, name string) float64 {
	if v, ok := m[name]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func parseTS(input string) (time.Time, error) { return parseTSWithFrom(input, "") }

func parseTSWithFrom(input, fromTZ string) (time.Time, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	if s == "" || s == "now" {
		return time.Now(), nil
	}

	// Unix epoch (seconds or milliseconds)
	if reNum := regexp.MustCompile(`^\d+$`); reNum.MatchString(s) {
		if len(s) >= 13 {
			ms, _ := strconv.ParseInt(s, 10, 64)
			return time.Unix(0, ms*int64(time.Millisecond)), nil
		}
		sec, _ := strconv.ParseInt(s, 10, 64)
		return time.Unix(sec, 0), nil
	}

	// RFC 3339 with timezone
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	// RFC3339-like without zone -> interpret in fromTZ or local
	if strings.Contains(input, "T") && !strings.ContainsAny(input, "zZ+-") {
		loc := time.Local
		if fromTZ != "" {
			if l, err := time.LoadLocation(fromTZ); err == nil {
				loc = l
			}
		}
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", input, loc); err == nil {
			return t, nil
		}
	}

	// Date-only
	if !strings.Contains(input, "T") {
		loc := time.Local
		if fromTZ != "" {
			if l, err := time.LoadLocation(fromTZ); err == nil {
				loc = l
			}
		}
		if d, err := time.ParseInLocation("2006-01-02", input, loc); err == nil {
			return d, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid ts: %q (RFC3339, YYYY-MM-DD, 'now' or unix epoch)", input)
}

func parseDur(args map[string]interface{}) (time.Duration, error) {
	// Structured fields first
	total := time.Duration(0)
	total += time.Duration(numArg(args, "seconds")) * time.Second
	total += time.Duration(numArg(args, "minutes")) * time.Minute
	total += time.Duration(numArg(args, "hours")) * time.Hour
	total += time.Duration(numArg(args, "days")) * 24 * time.Hour
	if total > 0 {
		return total, nil
	}

	// String-based: Go duration, ISO-8601, or simple human phrases
	s := pickArg(args, "dur", "duration", "delta")
	if strings.TrimSpace(s) == "" {
		return 0, fmt.Errorf("no duration provided")
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if d, err := parseISODur(s); err == nil {
		return d, nil
	}
	if d, err := parseHumanDur(s); err == nil {
		return d, nil
	}
	return 0, fmt.Errorf("invalid duration: %q", s)
}

// Support a small ISO-8601 subset: PnDTnHnMnS (any part optional)
func parseISODur(s string) (time.Duration, error) {
	re := regexp.MustCompile(`(?i)^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)
	m := re.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, fmt.Errorf("not ISO-8601 duration")
	}
	val := func(i int) int64 {
		if i >= len(m) || m[i] == "" {
			return 0
		}
		v, _ := strconv.ParseInt(m[i], 10, 64)
		return v
	}
	d := time.Duration(0)
	d += time.Duration(val(1)) * 24 * time.Hour
	d += time.Duration(val(2)) * time.Hour
	d += time.Duration(val(3)) * time.Minute
	d += time.Duration(val(4)) * time.Second
	return d, nil
}

// Parse human phrases like "2 hours", "1 day 30 min", "90min"
func parseHumanDur(s string) (time.Duration, error) {
	re := regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(days?|d|hours?|hrs?|h|minutes?|mins?|m|seconds?|secs?|s)`)
	var sec float64
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		numStr := m[1]
		unit := strings.ToLower(m[2])
		v, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			continue
		}
		switch unit {
		case "day", "days", "d":
			sec += v * 86400
		case "hour", "hours", "hr", "hrs", "h":
			sec += v * 3600
		case "minute", "minutes", "min", "mins", "m":
			sec += v * 60
		case "second", "seconds", "sec", "secs", "s":
			sec += v
		}
	}
	if sec == 0 {
		return 0, fmt.Errorf("not human duration")
	}
	return time.Duration(sec * float64(time.Second)), nil
}
