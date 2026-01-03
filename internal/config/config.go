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
	return &Config{
		DataDirectory: "~/.config/snipgo/snippets",
	}
}

// LoadConfig loads configuration from environment variables and config file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	configPath, err := GetConfigPath()

	if err != nil {
		return config, fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: config file does not exist: %s, using defaults\n", configPath)
		// Use default config
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
// Priority: 1. SNIPGO_CONFIG_PATH env var, 2. default path (~/.config/snipgo/config.yaml)
func GetConfigPath() (string, error) {
	// 1. Check environment variable first
	if envConfigPath := os.Getenv("SNIPGO_CONFIG_PATH"); envConfigPath != "" {
		return expandPath(envConfigPath), nil
	}

	// 2. Use default path
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
		return fmt.Errorf("failed to get config path: %w", err)
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
