package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"kunja/api"
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
				fmt.Printf("%d:  %s\n", task.ID, task.Title)
				if task.Description != "" {
					fmt.Printf("Description: %s\n", task.Description)
				}
				if !task.DueDate.IsZero() {
					fmt.Printf("Due Date: %s\n", task.DueDate.Format("2006-01-02"))
				}
				fmt.Println()
			}
		}
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			fmt.Println("Verbose mode enabled")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "username for the API")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "password for the API")
	rootCmd.PersistentFlags().StringVarP(&BaseUrl, "baseurl", "b", "", "base URL for the API")
}

func Execute() error {
	return rootCmd.Execute()
}
