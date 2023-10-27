package cmd

import (
	"fmt"
	"kunja/api"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new task",
	Long:  `Create a new task using the provided title and due date.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewApiClient(BaseUrl, "")
		_, err := client.Login(Username, Password, "")
		if err != nil {
			fmt.Println("Error logging in:", err)
			return
		}

		title := strings.Join(args, " ")
		due, _ := cmd.Flags().GetString("due")

		var dueDate time.Time
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
			client.Verbose = true
		}
		createdTask, err := client.CreateTask(projectId, task)
		if err != nil {
			fmt.Println("Error creating task:", err)
			return
		}

		fmt.Println("Task created successfully:", createdTask.ID)
	},
}

func init() {
	newCmd.Flags().StringP("due", "d", "", "Due date for the task")
	rootCmd.AddCommand(newCmd)

	doneCmd := &cobra.Command{
		Use:   "done",
		Short: "Mark a task as done",
		Long:  `Mark a task as done using the provided task ID.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.NewApiClient(BaseUrl, "")
			taskID, _ := strconv.Atoi(args[0])
			task, err := client.GetTask(taskID)
			if err != nil {
				fmt.Println("Error getting task:", err)
				return
			}
			fmt.Println("Task title:", task.Title)
		},
	}
	rootCmd.AddCommand(doneCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a task",
		Long:  `Edit a task using the provided task ID.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := api.NewApiClient(BaseUrl, "")
			taskID, _ := strconv.Atoi(args[0])
			task, err := client.GetTask(taskID)
			if err != nil {
				fmt.Println("Error getting task:", err)
				return
			}
			fmt.Println("Task title:", task.Title)
		},
	}
	rootCmd.AddCommand(editCmd)
}
