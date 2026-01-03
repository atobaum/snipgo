package main

import (
	"fmt"
	"io"
	"strings"

	"snipgo/internal/core"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new snippet",
	Long:  "Interactively create a new snippet by entering description and command",
	Args:  cobra.NoArgs,
	RunE:  runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
	// Prompt for description (title)
	description, err := readline.Line("Description> ")
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("failed to read description: %w", err)
	}
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	// Prompt for command (body)
	command, err := readline.Line("Command> ")
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("failed to read command: %w", err)
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Create snippet
	snippet := core.NewSnippet(description)
	snippet.Body = command

	// Save the snippet
	if err := manager.Save(snippet); err != nil {
		return fmt.Errorf("failed to save snippet: %w", err)
	}

	fmt.Printf("Snippet saved: %s\n", snippet.Title)
	return nil
}


