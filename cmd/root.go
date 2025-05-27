package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kunja/api" // Added for api package
	"kunja/adapter/vikunja"
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

var (
	Verbose  bool
	Username string
	Password string
	BaseUrl  string
	ShowAll  bool
	Svc      Services
)

var rootCmd = &cobra.Command{
	Use:   "kunja",
	Short: "A CLI client for the Vikunja task management API",
	Long:  `A CLI client for the Vikunja task management API. It allows you to interact with the Vikunja API from the command line.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("token")
		if token == "" {
			fmt.Println("No token found – please run `kunja login` first.")
			os.Exit(1)
		}

		client := api.NewApiClient(viper.GetString("baseUrl"), token)
		adapter := vikunja.New(client)

		Svc = Services{
			Auth:    adapter,
			Task:    adapter,
			Project: adapter,
			User:    adapter,
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		allTasks, err := Svc.Task.GetAllTasks(cmd.Context(), api.GetAllTasksParams{})
		if err != nil {
			fmt.Println("Error getting tasks:", err)
			return
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
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].Urgency == tasks[j].Urgency {
				return tasks[i].ID > tasks[j].ID // Descending ID
			}
			return tasks[i].Urgency > tasks[j].Urgency // Descending urgency
		})

		// Output tasks
		for _, task := range tasks {
			fmt.Printf("%d:  %s (Urgency: %.3f)\n", task.ID, task.Title, task.Urgency)
			if task.Description != "" {
				fmt.Printf("Description: %s\n", task.Description)
			}
			if !task.DueDate.IsZero() {
				fmt.Printf("Due Date: %s\n", task.DueDate.Format("2006-01-02"))
			}
			if !task.DoneAt.IsZero() {
				fmt.Printf("Done At: %s\n", task.DoneAt.Format("2006-01-02 15:04:05"))
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

// projectUsersCmd represents the project-users command
var projectUsersCmd = &cobra.Command{
	Use:   "project-users [PROJECT_ID]",
	Short: "List all users a project is shared with",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Error: Project ID must be a number")
			return
		}

		project, err := Svc.Project.GetProject(cmd.Context(), projectID)
		if err != nil {
			fmt.Printf("Error retrieving project: %s\n", err)
			return
		}
		fmt.Printf("Owner: ID: %d, Username: %s\n", project.Owner.ID, project.Owner.Username)

		users, err := Svc.Project.GetProjectUsers(cmd.Context(), projectID)
		if err != nil {
			fmt.Printf("Error retrieving project users: %s\n", err)
			return
		}

		for _, user := range users {
			fmt.Printf("User: ID: %d, Username: %s, Right: %d\n", user.ID, user.Username, user.Right)
		}
	},
}

func init() {
	rootCmd.AddCommand(projectUsersCmd)
}
func EditStringInEditor(initialContent string) (string, error) {
	// Create a temporary file
	file, err := ioutil.TempFile("", "example")
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
	cmd := exec.Command("sh", "-c", fmt.Sprintf("$EDITOR %s", file.Name()))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Read the contents of the file
	content, err := ioutil.ReadFile(file.Name())
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
