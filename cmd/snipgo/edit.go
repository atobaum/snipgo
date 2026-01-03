package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a snippet",
	Long:  "Interactively select a snippet using fzf and edit it with $EDITOR",
	Args:  cobra.NoArgs,
	RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	// Get all snippets
	snippets := manager.GetAll()
	if len(snippets) == 0 {
		return fmt.Errorf("no snippets found")
	}

	// Use fzf to select
	selected, err := selectSnippetWithFzf(snippets)
	if err != nil {
		return err
	}

	// Get editor from environment variable, default to vi
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "snipgo-edit-*.md")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // Clean up temp file

	// Serialize snippet to markdown
	content, err := serializeSnippetForEdit(selected)
	if err != nil {
		return fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Write to temp file
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Get file modification time before editing
	beforeStat, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to stat temporary file: %w", err)
	}
	beforeModTime := beforeStat.ModTime()

	// Open editor
	editCmd := exec.Command(editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Check if file was modified
	afterStat, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to stat temporary file after editing: %w", err)
	}
	afterModTime := afterStat.ModTime()

	// If file wasn't modified, user might have cancelled
	if beforeModTime.Equal(afterModTime) {
		return fmt.Errorf("file was not modified, edit cancelled")
	}

	// Read edited content
	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Parse edited content
	editedSnippet, err := parseSnippetFromEdit(editedContent)
	if err != nil {
		return fmt.Errorf("failed to parse edited content: %w", err)
	}

	// Validate edited snippet
	if err := editedSnippet.Validate(); err != nil {
		return fmt.Errorf("invalid snippet after editing: %w", err)
	}

	// Ensure ID matches (don't allow changing ID)
	editedSnippet.ID = selected.ID
	// Preserve created_at timestamp
	editedSnippet.CreatedAt = selected.CreatedAt

	// Save the edited snippet
	if err := manager.Save(editedSnippet); err != nil {
		return fmt.Errorf("failed to save edited snippet: %w", err)
	}

	fmt.Printf("Snippet '%s' updated successfully\n", editedSnippet.Title)
	return nil
}


