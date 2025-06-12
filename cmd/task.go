package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"kunja/api"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/AlecAivazis/survey/v2" // Import the survey library
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new task in a project (--project) with optional --due",
	Long:  `Create a new task using the provided title and due date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		svc := getServices(cmd)
		due, _ := cmd.Flags().GetString("due")


		projectId := viper.GetInt("project")
		if projectId == 0 {
			projects, err := svc.Project.GetAllProjects(cmd.Context())
			if err != nil {
				fmt.Println("Error retrieving projects:", err)
				return err
			}
			var options []string
			for _, p := range projects {
				options = append(options, fmt.Sprintf("%d: %s", p.ID, p.Title))
			}
			var selected string
			prompt := &survey.Select{
				Message: "Select project:",
				Options: options,
			}
			if err := survey.AskOne(prompt, &selected); err != nil {
				fmt.Println("Project selection cancelled")
				return fmt.Errorf("project selection cancelled")
			}
			parts := strings.SplitN(selected, ":", 2)
			projectId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		}

		msg, err := createTaskSimple(cmd.Context(), svc, title, due, projectId)
		if err != nil {
			fmt.Println("Error creating task:", err)
			return err
		}
		fmt.Println(msg)
		return nil
	},
}

var doneCmd = &cobra.Command{
	Use:   "done [TASK_ID...]",
	Short: "Toggle the done status of one or more tasks",
	Long:  `Toggle the done status of the specified task IDs. If no IDs are provided an interactive multi-select is shown.`,
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := getServices(cmd)
		ctx := cmd.Context()

		// ---------------------------------------------------------------
		// Build list of task IDs – either from CLI args or survey prompt
		// ---------------------------------------------------------------
		var ids []int
		if len(args) == 0 {
			// Interactive path: fetch open tasks and present a multi-select
			params := api.GetAllTasksParams{
				PerPage:          100,
				FilterBy:         "done",
				FilterValue:      "false",
				FilterComparator: "equals",
			}
			openTasks, err := fetchTasks(ctx, svc, params, 100)
			if err != nil {
				fmt.Println("Error retrieving tasks:", err)
				return err
			}
			// Show newest tasks first (highest ID first)
			sort.Slice(openTasks, func(i, j int) bool {
				return openTasks[i].ID > openTasks[j].ID
			})
			if len(openTasks) == 0 {
				fmt.Println("No open tasks found.")
				return nil
			}

			var options []string
			for _, t := range openTasks {
				options = append(options, fmt.Sprintf("%d: %s", t.ID, t.Title))
			}
			var selected []string
			prompt := &survey.MultiSelect{
				Message: "Select tasks to toggle done:",
				Options: options,
			}
			if err := survey.AskOne(prompt, &selected); err != nil {
				fmt.Println("Task selection cancelled")
				return fmt.Errorf("task selection cancelled")
			}
			for _, sel := range selected {
				parts := strings.SplitN(sel, ":", 2)
				id, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				ids = append(ids, id)
			}
		} else {
			// Non-interactive path: parse positional arguments
			for _, a := range args {
				id, err := strconv.Atoi(a)
				if err != nil {
					return fmt.Errorf("invalid task ID: %q", a)
				}
				ids = append(ids, id)
			}
		}

		if len(ids) == 0 {
			return fmt.Errorf("no task IDs supplied")
		}

		// ---------------------------------------------------------------
		// Toggle done for each ID and collect results
		// ---------------------------------------------------------------
		var toggled []string
		var failed []string
		for _, id := range ids {
			msg, err := toggleTaskDone(ctx, svc, id)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%d (%v)", id, err))
			} else {
				toggled = append(toggled, fmt.Sprintf("%d (%s)", id, msg))
			}
		}

		if len(toggled) > 0 {
			fmt.Printf("Toggled: %s\n", strings.Join(toggled, ", "))
		}
		if len(failed) > 0 {
			fmt.Printf("Failed:  %s\n", strings.Join(failed, ", "))
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:         "delete",
	Short:       "Delete a task",
	Annotations: map[string]string{"skip_mcp": "true"},
	Long:        `Delete a task permanently using the provided task ID.`,
	Args:        cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error converting task ID:", err)
			return err
		}
		svc := getServices(cmd)
		if _, err := svc.Task.DeleteTask(cmd.Context(), taskID); err != nil {
			fmt.Println("Error deleting task:", err)
			return err
		}
		fmt.Println("Task deleted successfully")
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a task (arg TASK_ID)",
	Long:  `Show the details of a task in raw indented JSON format.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := strconv.Atoi(args[0])
		svc := getServices(cmd)
		task, err := svc.Task.GetTask(cmd.Context(), taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return err
		}
		jsonTask, err := json.MarshalIndent(&task, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling task to JSON:", err)
			return err
		}
		fmt.Println(string(jsonTask))
		return nil
	},
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Long:  `List all the projects from the API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := getServices(cmd)
		out, err := buildProjectList(cmd.Context(), svc, Verbose)
		if err != nil {
			fmt.Println("Error retrieving projects:", err)
			return err
		}
		fmt.Print(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(projectsCmd)
}

var assignedCmd = &cobra.Command{
	Use:   "assigned",
	Short: "List assignees for a task (arg TASK_ID)",
	Long:  `List all the assignees assigned to a task using the provided task ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error converting task ID:", err)
			return err
		}
		svc := getServices(cmd)
		assignees, err := svc.Task.GetTaskAssignees(cmd.Context(), taskID)
		if err != nil {
			fmt.Println("Error getting assignees for task:", err)
			return err
		}
		for _, assignee := range assignees {
			fmt.Printf("ID: %d, Username: %s, Name: %s\n", assignee.ID, assignee.Username, assignee.Name)
		}
		return nil
	},
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users",
	Long:  `List all the users from the API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := getServices(cmd)
		users, err := svc.User.GetAllUsers(cmd.Context())
		if err != nil {
			fmt.Println("Error retrieving users:", err)
			return err
		}
		for _, user := range users {
			fmt.Printf("ID: %d, Username: %s, Name: %s\n", user.ID, user.Username, user.Name)
		}
		return nil
	},
}

func init() {
	newCmd.Flags().StringP("due", "d", "", "Due date for the task")
	newCmd.Flags().IntP("project", "P", 0, "Project ID to create the task in")

	// Make project flag required for MCP (but optional for CLI).
	f := newCmd.Flags().Lookup("project")
	f.Annotations = map[string][]string{"mcp_required": {"true"}}
	viper.BindPFlag("project", f)
	rootCmd.AddCommand(newCmd)

	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(deleteCmd)
	// Flags for non-interactive updates in edit command
	editCmd.Flags().StringP("title", "t", "", "New title for the task")
	editCmd.Flags().String("description", "", "New description for the task")
	editCmd.Flags().String("due", "", "New due date (YYYY-MM-DD)")
	editCmd.Flags().IntP("project", "P", 0, "New project ID")

	rootCmd.AddCommand(editCmd) // Add the edit command to the root command
	rootCmd.AddCommand(assignedCmd)
	rootCmd.AddCommand(usersCmd)
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task (interactive or via flags)",
	Long:  `Edit a task's title, description, or due date. Provide --title, --description, or --due for non-interactive updates; otherwise an interactive editor is opened.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := strconv.Atoi(args[0])
		svc := getServices(cmd)

		// --- flag based (non-interactive) path ---------------------------
		newTitle, _ := cmd.Flags().GetString("title")
		newDesc, _ := cmd.Flags().GetString("description")
		newDue, _ := cmd.Flags().GetString("due")
		newProject, _ := cmd.Flags().GetInt("project")
		scriptable := newTitle != "" || newDesc != "" || newDue != "" || newProject != 0

		// fetch current task (needed for both paths)
		task, err := svc.Task.GetTask(cmd.Context(), taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return err
		}

		if scriptable {
			msg, err := editTaskSimple(cmd.Context(), svc, taskID, newTitle, newDesc, newDue, newProject)
			if err != nil {
				fmt.Println("Error updating task:", err)
				return err
			}
			fmt.Println(msg)
			return nil
		}
		// ----------------------------------------------------------------

		// Define the options for interactive editing
		editOptions := []string{"Title", "Description", "Due Date", "Save"}
		var fieldToEdit string

		// Repeat the selection until the user chooses 'Save'
		for fieldToEdit != "Save" {
			prompt := &survey.Select{
				Message: "Choose a field to edit:",
				Options: editOptions,
			}
			survey.AskOne(prompt, &fieldToEdit)

			switch fieldToEdit {
			case "Title":
				editedTitle, err := EditStringInEditor(task.Title)
				if err != nil {
					fmt.Println("Error editing title:", err)
					continue
				}
				task.Title = editedTitle
			case "Description":
				editedDescription, err := EditStringInEditor(task.Description)
				if err != nil {
					fmt.Println("Error editing description:", err)
					continue
				}
				task.Description = editedDescription
			case "Due Date":
				prompt := &survey.Input{Message: "Enter new due date (YYYY-MM-DD):"}
				var newDueDate string
				survey.AskOne(prompt, &newDueDate)
				if newDueDate != "" {
					parsedDate, err := time.Parse("2006-01-02", newDueDate)
					if err != nil {
						fmt.Println("Error parsing due date:", err)
						continue
					}
					task.DueDate = parsedDate
				}
			}
		}

		// Save the updated task to the API
		_, err = svc.Task.UpdateTask(cmd.Context(), taskID, task)
		if err != nil {
			fmt.Println("Error updating task:", err)
			return err
		}
		fmt.Println("Task updated successfully")
		return nil
	},
}

// createTaskSimple contains the non-interactive business logic for creating a
// task.  It is reused by both the CLI and the MCP “new” tool.
func createTaskSimple(ctx context.Context, svc Services, title, dueStr string, projectID int) (string, error) {
	if projectID == 0 {
		return "", fmt.Errorf("project ID must be provided (flag --project)")
	}

	var dueDate time.Time
	var err error
	if dueStr != "" {
		dueDate, err = time.Parse("2006-01-02", dueStr)
		if err != nil {
			return "", fmt.Errorf("invalid due date: %w", err)
		}
	}

	task := api.Task{
		Title:     title,
		DueDate:   dueDate,
		ProjectID: projectID,
	}

	created, err := svc.Task.CreateTask(ctx, projectID, task)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Task created successfully: %d", created.ID), nil
}

// ---------------------------------------------------------------------
// Shared helpers used by both CLI commands and native MCP tools
// ---------------------------------------------------------------------

// buildProjectList returns a table (or JSON when verbose) of all projects.
func buildProjectList(ctx context.Context, svc Services, verbose bool) (string, error) {
	projects, err := svc.Project.GetAllProjects(ctx)
	if err != nil {
		return "", err
	}

	// pretty-print JSON for verbose output
	if verbose {
		if pretty, err := json.MarshalIndent(projects, "", "  "); err == nil {
			return string(pretty) + "\n", nil
		}
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tFav")
	for _, p := range projects {
		fav := ""
		if p.IsFavorite {
			fav = "★"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.ID, p.Title, fav)
	}
	w.Flush()
	return buf.String(), nil
}

// toggleTaskDone flips the Done flag of a task and saves it.
func toggleTaskDone(ctx context.Context, svc Services, taskID int) (string, error) {
	task, err := svc.Task.GetTask(ctx, taskID)
	if err != nil {
		return "", err
	}
	task.Done = !task.Done
	updated, err := svc.Task.UpdateTask(ctx, taskID, task)
	if err != nil {
		return "", err
	}
	if updated.Done {
		return "Task marked as done successfully", nil
	}
	return "Task marked as not done successfully", nil
}

// ---------------------------------------------------------------------
// editTaskSimple – shared helper for non-interactive task updates
// ---------------------------------------------------------------------
func editTaskSimple(ctx context.Context, svc Services, taskID int,
	title, desc, due string, projectID int) (string, error) {

	if title == "" && desc == "" && due == "" && projectID == 0 {
		return "", fmt.Errorf("at least one of --title/--description/--due/--project is required")
	}

	task, err := svc.Task.GetTask(ctx, taskID)
	if err != nil {
		return "", err
	}

	if title != "" {
		task.Title = title
	}
	if desc != "" {
		task.Description = desc
	}
	if due != "" {
		dt, err := time.Parse("2006-01-02", due)
		if err != nil {
			return "", fmt.Errorf("invalid --due: %w", err)
		}
		task.DueDate = dt
	}

	if projectID != 0 {
		task.ProjectID = projectID
	}

	if _, err := svc.Task.UpdateTask(ctx, taskID, task); err != nil {
		return "", err
	}
	return "Task updated successfully", nil
}
