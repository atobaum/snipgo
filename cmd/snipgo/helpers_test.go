package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditContentInEditor_Modified(t *testing.T) {
	// Create a mock editor script that modifies the file
	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")

	// Script that appends "modified" to the file
	script := `#!/bin/sh
echo "modified" >> "$1"
`
	if err := os.WriteFile(mockEditor, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock editor: %v", err)
	}

	// Set EDITOR to our mock
	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", mockEditor)
	defer os.Setenv("EDITOR", oldEditor)

	// Test editContentInEditor
	content := []byte("original content\n")
	result, err := editContentInEditor(content, "test-")
	if err != nil {
		t.Fatalf("editContentInEditor() error = %v", err)
	}

	expected := "original content\nmodified\n"
	if string(result) != expected {
		t.Errorf("editContentInEditor() = %q, want %q", string(result), expected)
	}
}

func TestEditContentInEditor_NotModified(t *testing.T) {
	// Create a mock editor script that does NOT modify the file
	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")

	// Script that does nothing (just exits)
	script := `#!/bin/sh
exit 0
`
	if err := os.WriteFile(mockEditor, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock editor: %v", err)
	}

	// Set EDITOR to our mock
	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", mockEditor)
	defer os.Setenv("EDITOR", oldEditor)

	// Test editContentInEditor - should return error since file not modified
	content := []byte("original content\n")
	_, err := editContentInEditor(content, "test-")
	if err == nil {
		t.Fatal("editContentInEditor() expected error for unmodified file, got nil")
	}
	if err.Error() != "file was not modified" {
		t.Errorf("editContentInEditor() error = %q, want %q", err.Error(), "file was not modified")
	}
}

func TestEditContentInEditor_EditorError(t *testing.T) {
	// Create a mock editor script that exits with error
	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor.sh")

	// Script that exits with error
	script := `#!/bin/sh
exit 1
`
	if err := os.WriteFile(mockEditor, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create mock editor: %v", err)
	}

	// Set EDITOR to our mock
	oldEditor := os.Getenv("EDITOR")
	os.Setenv("EDITOR", mockEditor)
	defer os.Setenv("EDITOR", oldEditor)

	// Test editContentInEditor - should return error
	content := []byte("original content\n")
	_, err := editContentInEditor(content, "test-")
	if err == nil {
		t.Fatal("editContentInEditor() expected error for editor failure, got nil")
	}
}

func TestEditContentInEditor_DefaultEditor(t *testing.T) {
	// Skip this test in CI/automated environments because vi hangs waiting for input
	// This test verifies default editor behavior which requires a real terminal
	t.Skip("Skipping default editor test - requires interactive terminal")
}
