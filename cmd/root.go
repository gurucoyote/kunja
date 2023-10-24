package cmd

import (
	"fmt"
	"kunja/api"
	"github.com/spf13/cobra"
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
