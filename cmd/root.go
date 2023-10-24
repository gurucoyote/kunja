package cmd

import (
	"fmt"
	"kunja/api"
	"github.com/spf13/cobra"
	"encoding/json"
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
		fmt.Println("Logged in with token:", token)

		tasks, err := client.GetAllTasks(api.GetAllTasksParams{})
		if err != nil {
			fmt.Println("Error getting tasks:", err)
			return
		}

		formattedTasks, _ := json.MarshalIndent(tasks, "", "  ")
		fmt.Println(string(formattedTasks))
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
