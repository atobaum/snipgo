package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestConfig creates a temporary config file and sets SNIPGO_CONFIG_PATH
// Returns cleanup function and error
func setupTestConfig(tmpDir string) (func(), error) {
	originalEnv := os.Getenv("SNIPGO_CONFIG_PATH")
	
	// Create temporary config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	content := "data_directory: " + tmpDir + "\n"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return nil, err
	}
	
	os.Setenv("SNIPGO_CONFIG_PATH", configPath)
	
	cleanup := func() {
		os.Setenv("SNIPGO_CONFIG_PATH", originalEnv)
		os.Remove(configPath)
	}
	
	return cleanup, nil
}

func TestNewFileSystem(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test config
	cleanup, err := setupTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}
	defer cleanup()

	fs, err := NewFileSystem()
	if err != nil {
		t.Fatalf("NewFileSystem() error = %v, want nil", err)
	}

	if fs == nil {
		t.Fatal("NewFileSystem() returned nil")
	}

	if fs.GetSnippetsDir() != tmpDir {
		t.Errorf("NewFileSystem() snippetsDir = %v, want %v", fs.GetSnippetsDir(), tmpDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("NewFileSystem() did not create snippets directory")
	}
}

func TestFileSystem_GetSnippetsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	if got := fs.GetSnippetsDir(); got != tmpDir {
		t.Errorf("FileSystem.GetSnippetsDir() = %v, want %v", got, tmpDir)
	}
}

func TestFileSystem_ListFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	tests := []struct {
		name      string
		setup     func() error
		wantCount int
		wantErr   bool
	}{
		{
			name: "empty directory",
			setup: func() error {
				return nil
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single .md file",
			setup: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte("content"), 0644)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple .md files",
			setup: func() error {
				files := []string{"test1.md", "test2.md", "test3.md"}
				for _, f := range files {
					if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "filter non-md files",
			setup: func() error {
				if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), []byte("content"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("content"), 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "case insensitive .md extension",
			setup: func() error {
				if err := os.WriteFile(filepath.Join(tmpDir, "test1.MD"), []byte("content"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(tmpDir, "test2.Md"), []byte("content"), 0644); err != nil {
					return err
				}
				return nil
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "files in subdirectory",
			setup: func() error {
				subDir := filepath.Join(tmpDir, "subdir")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(subDir, "test.md"), []byte("content"), 0644)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before setup
			os.RemoveAll(tmpDir)
			os.MkdirAll(tmpDir, 0755)

			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			files, err := fs.ListFiles()

			if (err != nil) != tt.wantErr {
				t.Errorf("FileSystem.ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(files) != tt.wantCount {
					t.Errorf("FileSystem.ListFiles() count = %v, want %v", len(files), tt.wantCount)
				}
			}
		})
	}
}

func TestFileSystem_ReadFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	tests := []struct {
		name    string
		setup   func() (string, []byte)
		want    []byte
		wantErr bool
	}{
		{
			name: "read existing file",
			setup: func() (string, []byte) {
				content := []byte("test content")
				filePath := filepath.Join(tmpDir, "test.md")
				os.WriteFile(filePath, content, 0644)
				return filePath, content
			},
			want:    []byte("test content"),
			wantErr: false,
		},
		{
			name: "read file with multiline content",
			setup: func() (string, []byte) {
				content := []byte("Line 1\nLine 2\nLine 3")
				filePath := filepath.Join(tmpDir, "test.md")
				os.WriteFile(filePath, content, 0644)
				return filePath, content
			},
			want:    []byte("Line 1\nLine 2\nLine 3"),
			wantErr: false,
		},
		{
			name: "read non-existent file",
			setup: func() (string, []byte) {
				return filepath.Join(tmpDir, "nonexistent.md"), nil
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, expectedContent := tt.setup()

			got, err := fs.ReadFile(filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("FileSystem.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if string(got) != string(expectedContent) {
					t.Errorf("FileSystem.ReadFile() = %v, want %v", string(got), string(expectedContent))
				}
			}
		})
	}
}

func TestFileSystem_WriteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	tests := []struct {
		name    string
		filePath string
		content  []byte
		wantErr  bool
	}{
		{
			name:     "write new file",
			filePath: filepath.Join(tmpDir, "test.md"),
			content:  []byte("test content"),
			wantErr:  false,
		},
		{
			name:     "write file with multiline content",
			filePath: filepath.Join(tmpDir, "multiline.md"),
			content:  []byte("Line 1\nLine 2\nLine 3"),
			wantErr:  false,
		},
		{
			name:     "overwrite existing file",
			filePath: filepath.Join(tmpDir, "existing.md"),
			content:  []byte("new content"),
			wantErr:  false,
		},
		{
			name:     "write empty file",
			filePath: filepath.Join(tmpDir, "empty.md"),
			content:  []byte(""),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For overwrite test, create file first
			if tt.name == "overwrite existing file" {
				os.WriteFile(tt.filePath, []byte("old content"), 0644)
			}

			err := fs.WriteFile(tt.filePath, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("FileSystem.WriteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was written
				readContent, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Errorf("FileSystem.WriteFile() failed to read back file: %v", err)
					return
				}

				if string(readContent) != string(tt.content) {
					t.Errorf("FileSystem.WriteFile() content = %v, want %v", string(readContent), string(tt.content))
				}
			}
		})
	}
}

func TestFileSystem_DeleteFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "delete existing file",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "test.md")
				os.WriteFile(filePath, []byte("content"), 0644)
				return filePath
			},
			wantErr: false,
		},
		{
			name: "delete non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.md")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setup()

			err := fs.DeleteFile(filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("FileSystem.DeleteFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was deleted
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Errorf("FileSystem.DeleteFile() file still exists: %v", filePath)
				}
			}
		})
	}
}

func TestFileSystem_FileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs := &FileSystem{
		snippetsDir: tmpDir,
	}

	tests := []struct {
		name     string
		setup    func() string
		want     bool
	}{
		{
			name: "file exists",
			setup: func() string {
				filePath := filepath.Join(tmpDir, "test.md")
				os.WriteFile(filePath, []byte("content"), 0644)
				return filePath
			},
			want: true,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.md")
			},
			want: false,
		},
		{
			name: "directory exists (not a file)",
			setup: func() string {
				dirPath := filepath.Join(tmpDir, "subdir")
				os.MkdirAll(dirPath, 0755)
				return dirPath
			},
			want: true, // FileExists checks if path exists, not if it's a file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			got := fs.FileExists(path)

			if got != tt.want {
				t.Errorf("FileSystem.FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

