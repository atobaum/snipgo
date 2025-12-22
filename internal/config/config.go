package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	DataDirectory string `yaml:"data_directory"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		DataDirectory: filepath.Join(homeDir, ".snipgo", "snippets"),
	}
}

// LoadConfig loads configuration from environment variables and config file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// 1. Check environment variable first
	if envDataDir := os.Getenv("SNIPGO_DATA_DIR"); envDataDir != "" {
		config.DataDirectory = expandPath(envDataDir)
		return config, nil
	}

	// 2. Load from config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use default
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	fileConfig := &Config{}
	if err := yaml.Unmarshal(data, fileConfig); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config (only if set)
	if fileConfig.DataDirectory != "" {
		config.DataDirectory = expandPath(fileConfig.DataDirectory)
	}

	return config, nil
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~
	if path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return filepath.Clean(path)
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "snipgo", "config.yaml"), nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
