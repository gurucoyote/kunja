package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	Version   = "0.1"
	AppName   = "kunja"
	ConfigDir string
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	ConfigDir = filepath.Join(os.Getenv("HOME"), "."+AppName)
	configPath := filepath.Join(ConfigDir, "config.yaml")
	viper.AddConfigPath(ConfigDir)

	if _, err := os.Stat(ConfigDir); os.IsNotExist(err) {
		err = os.Mkdir(ConfigDir, 0755)
		if err != nil {
			fmt.Println("Failed to create config directory:", err)
			os.Exit(1)
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		_, err := os.Create(configPath)
		if err != nil {
			fmt.Println("Failed to create config file:", err)
			os.Exit(1)
		}
		fmt.Println("Created config file at", configPath)
	}

	viper.ReadInConfig()
}
