package cmd

import (
	"fmt"
	"kunja/api"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
		}

		task := api.Task{
			Title:   title,
			DueDate: dueDate,
		}

		projectId := 1
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

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a task",
	Long:  `Edit a task using the provided task ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, _ := strconv.Atoi(args[0])
		task, err := ApiClient.GetTask(taskID)
		if err != nil {
			fmt.Println("Error getting task:", err)
			return
		}
		yamlTask, err := yaml.Marshal(&task)
		if err != nil {
			fmt.Println("Error marshaling task to YAML:", err)
			return
		}
		fmt.Println(string(yamlTask))
	},
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
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(assignedCmd)
	rootCmd.AddCommand(usersCmd)
}
