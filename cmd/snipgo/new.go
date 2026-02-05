package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"snipgo/internal/core"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var (
	useMultiLine bool
	useEditor    bool
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new snippet",
	Long:  "Interactively create a new snippet by entering description and command",
	Args:  cobra.NoArgs,
	RunE:  runNew,
}

func init() {
	newCmd.Flags().BoolVarP(&useMultiLine, "multiline", "m", false, "Enable multiline input (two blank lines to finish)")
	newCmd.Flags().BoolVarP(&useEditor, "editor", "e", false, "Open $EDITOR to create snippet")
}

func runNew(cmd *cobra.Command, args []string) error {
	// Check for mutually exclusive flags
	if useMultiLine && useEditor {
		return fmt.Errorf("--multiline and --editor flags are mutually exclusive")
	}

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

	var snippet *core.Snippet

	if useEditor {
		// Editor mode: open $EDITOR with template
		snippet, err = createSnippetWithEditor(description)
		if err != nil {
			return err
		}
	} else if useMultiLine {
		// Multiline mode: terminal input with continuation prompt
		command, err := scanMultiLine("Command> ", ".......> ")
		if err != nil {
			return err
		}
		if command == "" {
			return fmt.Errorf("command cannot be empty")
		}
		snippet = core.NewSnippet(description)
		snippet.Body = command
	} else {
		// Default: single-line input
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
		snippet = core.NewSnippet(description)
		snippet.Body = command
	}

	// Save the snippet
	if err := manager.Save(snippet); err != nil {
		return fmt.Errorf("failed to save snippet: %w", err)
	}

	fmt.Printf("Snippet saved: %s\n", snippet.Title)
	return nil
}

// scanMultiLine reads multiline input from terminal.
// Two consecutive blank lines finish the input.
func scanMultiLine(firstPrompt, continuationPrompt string) (string, error) {
	fmt.Print(firstPrompt)
	return scanMultiLineFromReader(os.Stdin, continuationPrompt)
}

// scanMultiLineFromReader reads multiline input from the given reader.
// Two consecutive blank lines finish the input. This function is separated
// for testability.
func scanMultiLineFromReader(r io.Reader, continuationPrompt string) (string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	emptyLineCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			emptyLineCount++
			if emptyLineCount >= 2 {
				// Two blank lines = done
				break
			}
			lines = append(lines, line)
		} else {
			emptyLineCount = 0
			lines = append(lines, line)
		}

		if continuationPrompt != "" {
			fmt.Print(continuationPrompt)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n"), nil
}

// createSnippetWithEditor opens $EDITOR with a template for creating a new snippet.
func createSnippetWithEditor(title string) (*core.Snippet, error) {
	// Create a new snippet with the title
	snippet := core.NewSnippet(title)

	// Serialize snippet to markdown
	content, err := serializeSnippetForEdit(snippet)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Open in editor
	editedContent, err := editContentInEditor(content, "snipgo-new-")
	if err != nil {
		return nil, fmt.Errorf("creation cancelled: %w", err)
	}

	// Parse edited content
	editedSnippet, err := parseSnippetFromEdit(editedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse edited content: %w", err)
	}

	// Validate edited snippet
	if err := editedSnippet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid snippet: %w", err)
	}

	return editedSnippet, nil
}
