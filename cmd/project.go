package cmd

import (
	"fmt"
	"strings"
	"kunja/api"
	"strconv"

	"github.com/spf13/cobra"
)

var projectNewCmd = &cobra.Command{
	Use:   "project-new [NAME]",
	Short: "Create a new project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		svc := getServices(cmd)
		p, err := svc.Project.CreateProject(cmd.Context(), api.Project{Title: name})
		if err != nil {
			return err
		}
		fmt.Printf("Project created: %d – %s\n", p.ID, p.Title)
		return nil
	},
}

var projectDelCmd = &cobra.Command{
	Use:   "project-del [ID]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := strconv.Atoi(args[0])
		svc := getServices(cmd)
		if _, err := svc.Project.DeleteProject(cmd.Context(), id); err != nil {
			return err
		}
		fmt.Println("Project deleted.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectNewCmd)
	rootCmd.AddCommand(projectDelCmd)
}
