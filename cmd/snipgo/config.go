package main

import (
	"fmt"
	"os"

	"snipgo/internal/config"

	"github.com/spf13/cobra"
)

// ConfigPathSource indicates where the config path came from
type ConfigPathSource string

const (
	ConfigPathSourceEnv     ConfigPathSource = "env"
	ConfigPathSourceDefault ConfigPathSource = "default"
)

// GetConfigPath returns the config path and its source
// Priority: 1. SNIPGO_CONFIG_PATH env var, 2. default path
func GetConfigPath(cmd *cobra.Command) (string, ConfigPathSource, error) {
	// 1. Check environment variable
	if envConfigPath := os.Getenv("SNIPGO_CONFIG_PATH"); envConfigPath != "" {
		path, err := config.GetConfigPath()
		if err != nil {
			return "", "", err
		}
		return path, ConfigPathSourceEnv, nil
	}

	// 2. Use default
	path, err := config.GetConfigPath()
	if err != nil {
		return "", "", err
	}
	return path, ConfigPathSourceDefault, nil
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View or set configuration options",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  "Set a configuration value. Available keys: data_directory",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap the configuration",
	Long:  "Bootstrap the configuration",
	Args:  cobra.NoArgs,
	RunE:  runConfigBootstrap,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configBootstrapCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	configPath, _, err := GetConfigPath(cmd)
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  Config File: %s\n", configPath)
	fmt.Printf("  Data Directory: %s\n", cfg.DataDirectory)

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "data_directory":
		cfg.DataDirectory = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}

func runConfigBootstrap(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()

	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s. Please delete file and re-run", configPath)
	}

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Configuration bootstrapped: %s\n", configPath)
	runConfigShow(cmd, args)
	return nil
}
