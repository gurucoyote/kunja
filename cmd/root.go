package cmd

import (
	"encoding/json"
	"fmt"
	"kunja/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Verbose   bool
	Username  string
	Password  string
	BaseUrl   string
	ApiClient *api.ApiClient
	ShowAll   bool
)

var rootCmd = &cobra.Command{
	Use:   "kunja",
	Short: "A CLI client for the Vikunja task management API",
	Long:  `A CLI client for the Vikunja task management API. It allows you to interact with the Vikunja API from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ApiClient = api.NewApiClient(viper.GetString("baseUrl"), "")
		token, err := ApiClient.Login(viper.GetString("username"), viper.GetString("password"), "")
		if err != nil {
			fmt.Println("Error logging in:", err)
			return
		}
		if Verbose {
			fmt.Println("Logged in with token:", token)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		allTasks, err := ApiClient.GetAllTasks(api.GetAllTasksParams{})
		if err != nil {
			fmt.Println("Error getting tasks:", err)
			return
		}

		var tasks []api.Task
		if !ShowAll {
			for _, task := range allTasks {
				if !task.Done {
					tasks = append(tasks, task)
				}
			}
		} else {
			tasks = allTasks
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
				if task.Done {
					fmt.Printf("Done At: %s\n", task.DoneAt.Format("2006-01-02 15:04:05"))
				}
				// fmt.Println()
			}
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&Username, "username", "u", "", "username for the API (can also be set with KUNJA_USERNAME environment variable)")
	rootCmd.PersistentFlags().StringVarP(&Password, "password", "p", "", "password for the API (can also be set with KUNJA_PASSWORD environment variable)")
	rootCmd.PersistentFlags().StringVarP(&BaseUrl, "baseurl", "b", "", "base URL for the API (can also be set with KUNJA_BASEURL environment variable)")
	rootCmd.PersistentFlags().BoolVarP(&ShowAll, "all", "a", false, "show all tasks")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("baseurl", rootCmd.PersistentFlags().Lookup("baseurl"))
	viper.BindPFlag("all", rootCmd.PersistentFlags().Lookup("all"))

}
