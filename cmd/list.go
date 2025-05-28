package cmd

// "list" sub-command: mirrors the default (root) listing behaviour so that the
// functionality is reachable both from an explicit command name (CLI + MCP)
// and from the bare invocation of "kunja".
//
// It simply delegates to rootCmd.RunE, avoiding any code duplication.

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List open or all tasks sorted by urgency",
	Long:  `List tasks from the API.  Shows open tasks by default; add --all to show completed ones as well. Output is sorted by urgency (desc).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delegate to the same RunE function that the root command uses.
		return rootCmd.RunE(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
