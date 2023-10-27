package cmd

import (
	"encoding/json"
	"fmt"
	"kunja/api"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Verbose  bool
	Username string
	Password string
	BaseUrl  string
)

var rootCmd = &cobra.Command{
	Use:   "kunja",
	Short: "A CLI client for the Vikunja task management API",
	Long:  `A CLI client for the Vikunja task management API. It allows you to interact with the Vikunja API from the command line.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewApiClient(BaseUrl, "")
		token, err := client.Login(Username, Password, "")
		if err != nil {
			fmt.Println("Error logging in:", err)
			return
		}
		if Verbose {
			fmt.Println("Logged in with token:", token)
		}

		allTasks, err := client.GetAllTasks(api.GetAllTasksParams{})
		if err != nil {
			fmt.Println("Error getting tasks:", err)
			return
		}

		var tasks []api.Task
		for _, task := range allTasks {
			if !task.Done {
				tasks = append(tasks, task)
			}
		}

		if Verbose {
			formattedTasks, _ := json.MarshalIndent(tasks, "", "  ")
			fmt.Println(string(formattedTasks))
		} else {
			for _, task := range tasks {
				fmt.Printf("%d:  %s (Urgency: %.3f)\n", task.ID, task.Title, task.Urgency)
				if task.Description != "" {
					fmt.Printf("Description: %s\n", task.Description)
				}
				if !task.DueDate.IsZero() {
					fmt.Printf("Due Date: %s\n", task.DueDate.Format("2006-01-02"))
				}
				// fmt.Println()
			}
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			fmt.Println("Verbose mode enabled")
		}
	},
}
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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "username for the API (can also be set with KUNJA_USERNAME environment variable)")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "password for the API (can also be set with KUNJA_PASSWORD environment variable)")
	rootCmd.PersistentFlags().StringVarP(&BaseUrl, "baseurl", "b", "", "base URL for the API (can also be set with KUNJA_BASEURL environment variable)")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("baseurl", rootCmd.PersistentFlags().Lookup("baseurl"))

	newCmd.Flags().StringP("due", "d", "", "Due date for the task")
	rootCmd.AddCommand(newCmd)
}
