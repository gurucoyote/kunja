package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := os.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItemDynamic(listFiles("./")),
)

func GetUserInput() string {
	historyFile := filepath.Join(ConfigDir, "cmd.history")
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
		HistoryFile:  historyFile,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	line, err := rl.Readline()
	if err != nil { // io.EOF
		return ""
	}
	return line
}

// this enters the main loop of asking for user input and executing commands
func Execute() {
	// execute once on startup for commandline params etc.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// enter continuous loop for further commands
	for {
		input := GetUserInput()
		args := strings.Fields(input)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

var ExitCmd = &cobra.Command{
	Use:     "exit",
	Aliases: []string{"q", "Q", "bye"},
	Short:   "Exit the application",
	Annotations: map[string]string{"skip_mcp": "true"},
	Long:    `This command will exit the application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Goodbye!")
		os.Exit(0)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ExitCmd)
}
