package main

import (
	"fmt"
	"os"

	"snipgo/internal/config"

	"github.com/spf13/cobra"
)

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

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  Data Directory: %s\n", cfg.DataDirectory)

	configPath, err := config.GetConfigPath()
	if err == nil {
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("  Config File: %s\n", configPath)
		} else {
			fmt.Printf("  Config File: (using defaults)\n")
		}
	}

	if os.Getenv("SNIPGO_DATA_DIR") != "" {
		fmt.Printf("  Environment Variable: SNIPGO_DATA_DIR=%s\n", os.Getenv("SNIPGO_DATA_DIR"))
	}

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
