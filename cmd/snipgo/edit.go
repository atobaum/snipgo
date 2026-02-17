package main

import (
	"fmt"

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
	snippets := app.manager.GetAll()
	if len(snippets) == 0 {
		return fmt.Errorf("no snippets found")
	}

	// Use fzf to select
	selected, err := selectSnippetWithFzf(snippets)
	if err != nil {
		return err
	}

	// Serialize snippet to markdown
	content, err := serializeSnippetForEdit(selected)
	if err != nil {
		return fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Open in editor
	editedContent, err := editContentInEditor(content, "snipgo-edit-")
	if err != nil {
		return fmt.Errorf("edit cancelled: %w", err)
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
	if err := app.manager.Save(editedSnippet); err != nil {
		return fmt.Errorf("failed to save edited snippet: %w", err)
	}

	fmt.Printf("Snippet '%s' updated successfully\n", editedSnippet.Title)
	return nil
}
