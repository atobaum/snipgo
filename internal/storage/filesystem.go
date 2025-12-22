package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"snipgo/internal/config"
)

// FileSystem handles file system operations for snippets
type FileSystem struct {
	snippetsDir string
}

// NewFileSystem creates a new FileSystem instance
func NewFileSystem() (*FileSystem, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	snippetsDir := cfg.DataDirectory
	if err := os.MkdirAll(snippetsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snippets directory: %w", err)
	}

	return &FileSystem{
		snippetsDir: snippetsDir,
	}, nil
}

// GetSnippetsDir returns the snippets directory path
func (fs *FileSystem) GetSnippetsDir() string {
	return fs.snippetsDir
}

// ListFiles returns all .md files in the snippets directory
func (fs *FileSystem) ListFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(fs.snippetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// ReadFile reads the content of a file
func (fs *FileSystem) ReadFile(filepath string) ([]byte, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}
	return data, nil
}

// WriteFile writes content to a file
func (fs *FileSystem) WriteFile(filepath string, data []byte) error {
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}
	return nil
}

// DeleteFile deletes a file
func (fs *FileSystem) DeleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filepath, err)
	}
	return nil
}

// FileExists checks if a file exists
func (fs *FileSystem) FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

