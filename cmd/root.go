package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	Verbose  bool
	Username string
)

var rootCmd = &cobra.Command{
	Use:   "vikunja",
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
}

func Execute() error {
	return rootCmd.Execute()
}
