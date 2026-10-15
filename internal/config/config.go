package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ConflictManual = "manual"
	ConflictLocal  = "local"
	ConflictRemote = "remote"

	defaultSyncIntervalMinutes = 10
)

// SyncConfig controls git sync of the data directory.
type SyncConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	IntervalMinutes int    `yaml:"interval_minutes" json:"intervalMinutes"`
	Conflict        string `yaml:"conflict" json:"conflict"`
}

func DefaultSyncConfig() SyncConfig {
	return SyncConfig{IntervalMinutes: defaultSyncIntervalMinutes, Conflict: ConflictManual}
}

func validConflict(c string) bool {
	return c == ConflictManual || c == ConflictLocal || c == ConflictRemote
}

// ValidateSync rejects values a user tries to save.
func ValidateSync(s SyncConfig) error {
	if s.IntervalMinutes < 1 {
		return fmt.Errorf("sync interval must be at least 1 minute, got %d", s.IntervalMinutes)
	}
	if !validConflict(s.Conflict) {
		return fmt.Errorf("sync conflict must be one of manual, local, remote, got %q", s.Conflict)
	}
	return nil
}

// NormalizeSync replaces invalid values from a config file with defaults.
func NormalizeSync(s SyncConfig) SyncConfig {
	if s.IntervalMinutes < 1 {
		slog.Warn("invalid sync.interval_minutes, using default", "value", s.IntervalMinutes, "default", defaultSyncIntervalMinutes)
		s.IntervalMinutes = defaultSyncIntervalMinutes
	}
	if !validConflict(s.Conflict) {
		slog.Warn("invalid sync.conflict, using default", "value", s.Conflict, "default", ConflictManual)
		s.Conflict = ConflictManual
	}
	return s
}

// SetValue sets a config key from its string form, as used by `snipgo config set`.
func SetValue(cfg *Config, key, value string) error {
	switch key {
	case "data_directory":
		cfg.DataDirectory = value
	case "sync.enabled":
		b, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("sync.enabled must be true or false, got %q", value)
		}
		cfg.Sync.Enabled = b
	case "sync.interval_minutes":
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || n < 1 {
			return fmt.Errorf("sync.interval_minutes must be an integer of at least 1, got %q", value)
		}
		cfg.Sync.IntervalMinutes = n
	case "sync.conflict":
		if !validConflict(value) {
			return fmt.Errorf("sync.conflict must be one of manual, local, remote, got %q", value)
		}
		cfg.Sync.Conflict = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}

// Config holds the application configuration
type Config struct {
	DataDirectory string     `yaml:"data_directory"`
	Sync          SyncConfig `yaml:"sync"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DataDirectory: "~/.config/snipgo/snippets",
		Sync:          DefaultSyncConfig(),
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
		if os.Getenv("SNIPGO_CONFIG_PATH") != "" {
			slog.Warn("config file does not exist, using defaults", "path", configPath)
		} else {
			slog.Debug("config file does not exist, using defaults", "path", configPath)
		}
		config.DataDirectory = expandPath(config.DataDirectory)
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	fileConfig := &Config{Sync: DefaultSyncConfig()}
	if err := yaml.Unmarshal(data, fileConfig); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config (only if set)
	if fileConfig.DataDirectory != "" {
		config.DataDirectory = fileConfig.DataDirectory
	}

	config.Sync = NormalizeSync(fileConfig.Sync)

	config.DataDirectory = expandPath(config.DataDirectory)
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
