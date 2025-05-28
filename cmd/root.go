package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"kunja/adapter/vikunja"
	"kunja/api" // Added for api package
	"kunja/internal/service"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Services struct {
	Auth    service.AuthService
	Task    service.TaskService
	Project service.ProjectService
	User    service.UserService
}

type ctxKey int

const servicesKey ctxKey = iota

func getServices(cmd *cobra.Command) Services {
	svc, _ := cmd.Context().Value(servicesKey).(Services)
	return svc
}

var (
	Verbose  bool
	Username string
	Password string
	BaseUrl  string
	ShowAll  bool
)

var rootCmd = &cobra.Command{
	Use:   "kunja",
	Short: "A CLI client for the Vikunja task management API",
	Long:  `A CLI client for the Vikunja task management API. It allows you to interact with the Vikunja API from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip authentication check when running the `login` command
		if cmd.Name() == "login" {
			return
		}
		token := viper.GetString("token")
		if token == "" {
			fmt.Println("No token found – please run `kunja login` first.")
			os.Exit(1)
		}

		client := api.NewApiClient(viper.GetString("baseUrl"), token)
		adapter := vikunja.New(client)

		services := Services{
			Auth:    adapter,
			Task:    adapter,
			Project: adapter,
			User:    adapter,
		}

		ctx := context.WithValue(cmd.Context(), servicesKey, services)
		cmd.SetContext(ctx)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		svc := getServices(cmd)
		allTasks, err := svc.Task.GetAllTasks(cmd.Context(), api.GetAllTasksParams{})
		if err != nil {
			return err
		}
		if Verbose {
			formattedTasks, _ := json.MarshalIndent(allTasks, "", "  ")
			fmt.Println(string(formattedTasks))
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

		// Sort tasks by urgency in descending order, then by ID in descending order
		out, err := buildTaskList(cmd.Context(), svc, Verbose, ShowAll)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

func buildTaskList(ctx context.Context, svc Services, verbose, showAll bool) (string, error) {
	allTasks, err := svc.Task.GetAllTasks(ctx, api.GetAllTasksParams{})
	if err != nil {
		return "", err
	}

	// pretty-print and return raw JSON when verbose
	if verbose {
		if pretty, err := json.MarshalIndent(allTasks, "", "  "); err == nil {
			return string(pretty) + "\n", nil
		}
	}

	var tasks []api.Task
	if !showAll {
		for _, t := range allTasks {
			if !t.Done {
				tasks = append(tasks, t)
			}
		}
	} else {
		tasks = allTasks
	}

	// sort by urgency desc, then ID desc
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Urgency == tasks[j].Urgency {
			return tasks[i].ID > tasks[j].ID
		}
		return tasks[i].Urgency > tasks[j].Urgency
	})

	var b strings.Builder
	for _, task := range tasks {
		fmt.Fprintf(&b, "%d:  %s (Urgency: %.3f)\n", task.ID, task.Title, task.Urgency)
		if task.Description != "" {
			fmt.Fprintf(&b, "Description: %s\n", task.Description)
		}
		if !task.DueDate.IsZero() {
			fmt.Fprintf(&b, "Due Date: %s\n", task.DueDate.Format("2006-01-02"))
		}
		if !task.DoneAt.IsZero() {
			fmt.Fprintf(&b, "Done At: %s\n", task.DoneAt.Format("2006-01-02 15:04:05"))
		}
	}
	return b.String(), nil
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

// projectUsersCmd represents the project-users command
var projectUsersCmd = &cobra.Command{
	Use:   "project-users [PROJECT_ID]",
	Short: "List users a project is shared with (arg PROJECT_ID)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error: Project ID must be a number")
			return fmt.Errorf("project ID must be a number")
		}

		svc := getServices(cmd)
		project, err := svc.Project.GetProject(cmd.Context(), projectID)
		if err != nil {
			fmt.Printf("Error retrieving project: %s\n", err)
			return err
		}
		fmt.Printf("Owner: ID: %d, Username: %s\n", project.Owner.ID, project.Owner.Username)

		users, err := svc.Project.GetProjectUsers(cmd.Context(), projectID)
		if err != nil {
			fmt.Printf("Error retrieving project users: %s\n", err)
			return err
		}

		for _, user := range users {
			fmt.Printf("User: ID: %d, Username: %s, Right: %d\n", user.ID, user.Username, user.Right)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectUsersCmd)
}
func EditStringInEditor(initialContent string) (string, error) {
	// Create a temporary file
	file, err := os.CreateTemp("", "example")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())

	// Write the initial content to the file
	_, err = file.WriteString(initialContent)
	if err != nil {
		return "", err
	}

	// Open the file in the default text editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, file.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Read the contents of the file
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return "", err
	}

	// Remove comments from the content
	lines := strings.Split(string(content), "\n")
	var filteredLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			filteredLines = append(filteredLines, line)
		}
	}
	filteredContent := strings.Join(filteredLines, "\n")

	return filteredContent, nil
}
