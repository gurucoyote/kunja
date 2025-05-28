package cmd

import (
	"encoding/json"
	"fmt"
	"kunja/api"
	"os"
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
	Short: "Create a new task",
	Long:  `Create a new task using the provided title and due date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := strings.Join(args, " ")
		svc := getServices(cmd)
		due, _ := cmd.Flags().GetString("due")

		var dueDate time.Time
		var err error
		if due != "" {
			dueDate, err = time.Parse("2006-01-02", due)
			if err != nil {
				fmt.Println("Error parsing due date:", err)
				return err
			}
			// fmt.Println("due: ", dueDate)
		}

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

		task := api.Task{
			Title:     title,
			DueDate:   dueDate,
			ProjectID: projectId,
		}

		// verbose mode can be handled inside the service adapter later
		createdTask, err := svc.Task.CreateTask(cmd.Context(), projectId, task)
		if err != nil {
			fmt.Println("Error creating task:", err)
			return err
		}

		fmt.Println("Task created successfully:", createdTask.ID)
		return nil
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark a task as done",
	Long:  `Mark a task as done using the provided task ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := strconv.Atoi(args[0])
		svc := getServices(cmd)
		task, err := svc.Task.GetTask(cmd.Context(), taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return err
		}
		task.Done = !task.Done // Toggle the done status
		updatedTask, err := svc.Task.UpdateTask(cmd.Context(), taskID, task)
		if err != nil {
			fmt.Println("Error updating task:", err)
			return err
		}
		if updatedTask.Done {
			fmt.Println("Task marked as done successfully")
		} else {
			fmt.Println("Task marked as not done successfully")
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a task",
	Long:  `Delete a task permanently using the provided task ID.`,
	Args:  cobra.ExactArgs(1),
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
	Short: "Show task details",
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
		projects, err := svc.Project.GetAllProjects(cmd.Context())
		if err != nil {
			fmt.Println("Error retrieving projects:", err)
			return err
		}

		// human-friendly table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTitle\tFav")
		for _, p := range projects {
			fav := ""
			if p.IsFavorite {
				fav = "★"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\n", p.ID, p.Title, fav)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(projectsCmd)
}

var assignedCmd = &cobra.Command{
	Use:   "assigned",
	Short: "List assignees of a task",
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
	rootCmd.AddCommand(editCmd) // Add the edit command to the root command
	rootCmd.AddCommand(assignedCmd)
	rootCmd.AddCommand(usersCmd)
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task",
	Long:  `Edit a task's title, description, or due date.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := strconv.Atoi(args[0])
		svc := getServices(cmd)
		task, err := svc.Task.GetTask(cmd.Context(), taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return err
		}

		// Define the options for editing
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
