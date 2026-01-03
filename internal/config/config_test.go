package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.DataDirectory == "" {
		t.Error("DefaultConfig() DataDirectory is empty")
	}

	// Verify it contains ~/.config/snipgo/snippets (with ~, not expanded)
	expected := "~/.config/snipgo/snippets"
	if cfg.DataDirectory != expected {
		t.Errorf("DefaultConfig() DataDirectory = %v, want %v", cfg.DataDirectory, expected)
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalConfigPathEnv := os.Getenv("SNIPGO_CONFIG_PATH")
	defer func() {
		os.Setenv("SNIPGO_CONFIG_PATH", originalConfigPathEnv)
	}()

	// Create temporary directory for config file
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		setup     func() error
		cleanup   func() error
		wantDir   string
		wantErr   bool
	}{
		{
			name: "no config file",
			setup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				// Remove config file if exists
				homeDir, _ := os.UserHomeDir()
				configPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
				os.Remove(configPath)
				return nil
			},
			cleanup: func() error {
				return nil
			},
			wantDir: "", // Will use default
			wantErr: false,
		},
		{
			name: "config file exists with valid YAML",
			setup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				homeDir, _ := os.UserHomeDir()
				configDir := filepath.Join(homeDir, ".config", "snipgo")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					return err
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := "data_directory: " + tmpDir + "\n"
				return os.WriteFile(configPath, []byte(content), 0644)
			},
			cleanup: func() error {
				homeDir, _ := os.UserHomeDir()
				configPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
				return os.Remove(configPath)
			},
			wantDir: tmpDir,
			wantErr: false,
		},
		{
			name: "config file with invalid YAML",
			setup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				homeDir, _ := os.UserHomeDir()
				configDir := filepath.Join(homeDir, ".config", "snipgo")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					return err
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := "invalid: [unclosed\n"
				return os.WriteFile(configPath, []byte(content), 0644)
			},
			cleanup: func() error {
				homeDir, _ := os.UserHomeDir()
				configPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
				return os.Remove(configPath)
			},
			wantDir: "", // Will use default
			wantErr: true, // LoadConfig returns error on parse error
		},
		{
			name: "config file with empty data_directory",
			setup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				homeDir, _ := os.UserHomeDir()
				configDir := filepath.Join(homeDir, ".config", "snipgo")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					return err
				}
				configPath := filepath.Join(configDir, "config.yaml")
				content := "data_directory: \"\"\n"
				return os.WriteFile(configPath, []byte(content), 0644)
			},
			cleanup: func() error {
				homeDir, _ := os.UserHomeDir()
				configPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
				return os.Remove(configPath)
			},
			wantDir: "", // Empty string means use default
			wantErr: false,
		},
		{
			name: "config file path from environment variable",
			setup: func() error {
				// Create custom config file in temp dir
				customConfigPath := filepath.Join(tmpDir, "custom-config.yaml")
				content := "data_directory: " + tmpDir + "\n"
				if err := os.WriteFile(customConfigPath, []byte(content), 0644); err != nil {
					return err
				}
				os.Setenv("SNIPGO_CONFIG_PATH", customConfigPath)
				return nil
			},
			cleanup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				return nil
			},
			wantDir: tmpDir,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			cfg, err := LoadConfig()

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				if tt.cleanup != nil {
					tt.cleanup()
				}
				return
			}

			if cfg == nil {
				t.Error("LoadConfig() returned nil config")
				if tt.cleanup != nil {
					tt.cleanup()
				}
				return
			}

			if tt.wantDir != "" {
				if cfg.DataDirectory != tt.wantDir {
					t.Errorf("LoadConfig() DataDirectory = %v, want %v", cfg.DataDirectory, tt.wantDir)
				}
			} else {
				// Verify default is used (with ~, not expanded)
				expected := "~/.config/snipgo/snippets"
				if cfg.DataDirectory != expected {
					t.Errorf("LoadConfig() DataDirectory = %v, want default %v", cfg.DataDirectory, expected)
				}
			}

			if tt.cleanup != nil {
				if err := tt.cleanup(); err != nil {
					t.Logf("Cleanup warning: %v", err)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		path     string
		setup    func() error
		cleanup  func() error
		want     string
	}{
		{
			name: "expand ~",
			path: "~/.snipgo",
			want: filepath.Join(homeDir, ".snipgo"),
		},
		{
			name: "expand ~ with subpath",
			path: "~/test/path",
			want: filepath.Join(homeDir, "test", "path"),
		},
		{
			name: "expand environment variable",
			path: "$HOME/.snipgo",
			setup: func() error {
				return nil
			},
			want: filepath.Join(homeDir, ".snipgo"),
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "normal path",
			path: "/tmp/test",
			want: "/tmp/test",
		},
		{
			name: "path with ~ in middle",
			path: "/tmp/~test",
			want: filepath.Clean("/tmp/~test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			got := expandPath(tt.path)

			if got != tt.want {
				t.Errorf("expandPath(%v) = %v, want %v", tt.path, got, tt.want)
			}

			if tt.cleanup != nil {
				if err := tt.cleanup(); err != nil {
					t.Logf("Cleanup warning: %v", err)
				}
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("SNIPGO_CONFIG_PATH")
	defer os.Setenv("SNIPGO_CONFIG_PATH", originalEnv)

	tests := []struct {
		name      string
		setup     func() error
		cleanup   func() error
		wantPath  string
		wantErr   bool
	}{
		{
			name: "environment variable set",
			setup: func() error {
				os.Setenv("SNIPGO_CONFIG_PATH", "/tmp/custom-config.yaml")
				return nil
			},
			cleanup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				return nil
			},
			wantPath: "/tmp/custom-config.yaml",
			wantErr:  false,
		},
		{
			name: "no environment variable, use default",
			setup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				return nil
			},
			cleanup: func() error {
				return nil
			},
			wantPath: "", // Will check against default
			wantErr:  false,
		},
		{
			name: "environment variable with ~ expansion",
			setup: func() error {
				os.Setenv("SNIPGO_CONFIG_PATH", "~/custom-config.yaml")
				return nil
			},
			cleanup: func() error {
				os.Unsetenv("SNIPGO_CONFIG_PATH")
				return nil
			},
			wantPath: "", // Will check against expanded path
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			path, err := GetConfigPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				if tt.cleanup != nil {
					tt.cleanup()
				}
				return
			}

			if tt.wantPath != "" {
				if path != tt.wantPath {
					t.Errorf("GetConfigPath() = %v, want %v", path, tt.wantPath)
				}
			} else {
				// Verify default path or expanded path
				if path == "" {
					t.Error("GetConfigPath() returned empty path")
				}
				if tt.name == "no environment variable, use default" {
					homeDir, _ := os.UserHomeDir()
					expected := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
					if path != expected {
						t.Errorf("GetConfigPath() = %v, want default %v", path, expected)
					}
				} else if tt.name == "environment variable with ~ expansion" {
					homeDir, _ := os.UserHomeDir()
					expected := filepath.Join(homeDir, "custom-config.yaml")
					if path != expected {
						t.Errorf("GetConfigPath() = %v, want expanded %v", path, expected)
					}
				}
			}

			if tt.cleanup != nil {
				if err := tt.cleanup(); err != nil {
					t.Logf("Cleanup warning: %v", err)
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original config path
	homeDir, _ := os.UserHomeDir()
	originalConfigPath := filepath.Join(homeDir, ".config", "snipgo", "config.yaml")
	originalConfigExists := false
	var originalContent []byte
	if _, err := os.Stat(originalConfigPath); err == nil {
		originalConfigExists = true
		originalContent, _ = os.ReadFile(originalConfigPath)
	}
	defer func() {
		if originalConfigExists {
			os.WriteFile(originalConfigPath, originalContent, 0644)
		} else {
			os.Remove(originalConfigPath)
		}
	}()

	// Create config directory
	configDir := filepath.Dir(originalConfigPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "save valid config",
			config: &Config{
				DataDirectory: tmpDir,
			},
			wantErr: false,
		},
		{
			name: "save config with empty data_directory",
			config: &Config{
				DataDirectory: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Remove config file if exists
			os.Remove(originalConfigPath)

			err := SaveConfig(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(originalConfigPath); os.IsNotExist(err) {
					t.Error("SaveConfig() did not create config file")
					return
				}

				// Verify content
				content, err := os.ReadFile(originalConfigPath)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}

				if len(content) == 0 {
					t.Error("SaveConfig() created empty config file")
				}

				// Verify it contains data_directory
				contentStr := string(content)
				if !contains(contentStr, "data_directory") {
					t.Error("SaveConfig() config file does not contain data_directory")
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

