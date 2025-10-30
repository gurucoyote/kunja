package cmd

import (
	"errors"
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

func defaultConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, AppName)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".config", AppName)
	}
	return filepath.Join(".", "."+AppName)
}

func legacyConfigDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, "."+AppName)
	}
	return ""
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	defaultDir := defaultConfigDir()
	legacyDir := legacyConfigDir()

	ConfigDir = defaultDir
	viper.AddConfigPath(ConfigDir)
	if legacyDir != "" && legacyDir != ConfigDir {
		viper.AddConfigPath(legacyDir)
	}

	if err := os.MkdirAll(ConfigDir, 0o755); err != nil {
		fmt.Println("Failed to create config directory:", err)
		os.Exit(1)
	}

	configPath := filepath.Join(ConfigDir, "config.yaml")
	createdConfig := false
	legacyPath := ""
	if legacyDir != "" && legacyDir != ConfigDir {
		legacyPath = filepath.Join(legacyDir, "config.yaml")
	}
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		if legacyPath != "" {
			if data, readErr := os.ReadFile(legacyPath); readErr == nil {
				if writeErr := os.WriteFile(configPath, data, 0o600); writeErr != nil {
					fmt.Println("Failed to copy legacy config file:", writeErr)
					os.Exit(1)
				}
			} else if !errors.Is(readErr, os.ErrNotExist) {
				fmt.Println("Failed to read legacy config file:", readErr)
				os.Exit(1)
			} else {
				if createErr := os.WriteFile(configPath, []byte{}, 0o600); createErr != nil {
					fmt.Println("Failed to create config file:", createErr)
					os.Exit(1)
				}
				fmt.Println("Created config file at", configPath)
				createdConfig = true
			}
		} else {
			if createErr := os.WriteFile(configPath, []byte{}, 0o600); createErr != nil {
				fmt.Println("Failed to create config file:", createErr)
				os.Exit(1)
			}
			fmt.Println("Created config file at", configPath)
			createdConfig = true
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Println("Failed to inspect config file:", err)
		os.Exit(1)
	}

	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil && !createdConfig {
		fmt.Println("Warning: failed to read config file:", err)
	}
}
