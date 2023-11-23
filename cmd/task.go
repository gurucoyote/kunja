package cmd

import (
	"encoding/json"
	"fmt"
	"kunja/api"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2" // Import the survey library
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new task",
	Long:  `Create a new task using the provided title and due date.`,
	Run: func(cmd *cobra.Command, args []string) {
		title := strings.Join(args, " ")
		due, _ := cmd.Flags().GetString("due")

		var dueDate time.Time
		var err error
		if due != "" {
			dueDate, err = time.Parse("2006-01-02", due)
			if err != nil {
				fmt.Println("Error parsing due date:", err)
				return
			}
			// fmt.Println("due: ", dueDate)
		}

		projectId := 1
		task := api.Task{
			Title:     title,
			DueDate:   dueDate,
			ProjectID: projectId,
		}

		if Verbose {
			ApiClient.Verbose = true
		}
		createdTask, err := ApiClient.CreateTask(projectId, task)
		if err != nil {
			fmt.Println("Error creating task:", err)
			return
		}

		fmt.Println("Task created successfully:", createdTask.ID)
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "Mark a task as done",
	Long:  `Mark a task as done using the provided task ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, _ := strconv.Atoi(args[0])
		task, err := ApiClient.GetTask(taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return
		}
		task.Done = true
		_, err = ApiClient.UpdateTask(taskID, task)
		if err != nil {
			fmt.Println("Error updating task:", err)
			return
		}
		fmt.Println("Task marked as done successfully")
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show task details",
	Long:  `Show the details of a task in raw indented JSON format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, _ := strconv.Atoi(args[0])
		task, err := ApiClient.GetTask(taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return
		}
		jsonTask, err := json.MarshalIndent(&task, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling task to JSON:", err)
			return
		}
		fmt.Println(string(jsonTask))
	},
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Long:  `List all the projects from the API.`,
	Run: func(cmd *cobra.Command, args []string) {
		projects, err := ApiClient.GetAllProjects()
		if err != nil {
			fmt.Println("Error retrieving projects:", err)
			return
		}
		jsonProjects, err := json.MarshalIndent(&projects, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling projects to JSON:", err)
			return
		}
		fmt.Println(string(jsonProjects))
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
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error converting task ID:", err)
			return
		}
		assignees, err := ApiClient.GetTaskAssignees(taskID)
		if err != nil {
			fmt.Println("Error getting assignees for task:", err)
			return
		}
		for _, assignee := range assignees {
			fmt.Printf("ID: %d, Username: %s, Name: %s\n", assignee.ID, assignee.Username, assignee.Name)
		}
	},
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users",
	Long:  `List all the users from the API.`,
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			ApiClient.Verbose = true
		}
		users, err := ApiClient.GetAllUsers()
		if err != nil {
			fmt.Println("Error retrieving users:", err)
			return
		}
		for _, user := range users {
			fmt.Printf("ID: %d, Username: %s, Name: %s\n", user.ID, user.Username, user.Name)
		}
	},
}

func init() {
	newCmd.Flags().StringP("due", "d", "", "Due date for the task")
	rootCmd.AddCommand(newCmd)

	rootCmd.AddCommand(doneCmd)
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
	Run: func(cmd *cobra.Command, args []string) {
		taskID, _ := strconv.Atoi(args[0])
		task, err := ApiClient.GetTask(taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return
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
				prompt := &survey.Input{Message: "Enter new title:"}
				survey.AskOne(prompt, &task.Title)
			case "Description":
				prompt := &survey.Input{Message: "Enter new description:"}
				survey.AskOne(prompt, &task.Description)
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
		_, err = ApiClient.UpdateTask(taskID, task)
		if err != nil {
			fmt.Println("Error updating task:", err)
			return
		}
		fmt.Println("Task updated successfully")
	},
}
