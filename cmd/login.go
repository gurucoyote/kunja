package cmd

import (
	"fmt"

	"kunja/adapter/vikunja"
	"kunja/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Vikunja API and store the token in the config",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		username := viper.GetString("username")
		password := viper.GetString("password")
		baseURL := viper.GetString("baseUrl")

		if username == "" || password == "" || baseURL == "" {
			fmt.Println("username, password and baseurl must be set (flags, env or config)")
			return
		}

		client := api.NewApiClient(baseURL, "")
		adapter := vikunja.New(client)

		token, err := adapter.Login(ctx, username, password, "")
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}

		viper.Set("token", token)
		if err := viper.WriteConfig(); err != nil {
			fmt.Println("Failed to write config:", err)
			return
		}

		fmt.Println("Login successful – token saved to config.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
