package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	// Get editor from environment variable, default to vi
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Create a new snippet with the title
	snippet := core.NewSnippet(title)

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "snipgo-new-*.md")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Serialize snippet to markdown
	content, err := serializeSnippetForEdit(snippet)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Write to temp file
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write to temporary file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Get file modification time before editing
	beforeStat, err := os.Stat(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat temporary file: %w", err)
	}
	beforeModTime := beforeStat.ModTime()

	// Open editor
	editCmd := exec.Command(editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return nil, fmt.Errorf("editor exited with error: %w", err)
	}

	// Check if file was modified
	afterStat, err := os.Stat(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat temporary file after editing: %w", err)
	}
	afterModTime := afterStat.ModTime()

	// If file wasn't modified, user might have cancelled
	if beforeModTime.Equal(afterModTime) {
		return nil, fmt.Errorf("file was not modified, creation cancelled")
	}

	// Read edited content
	editedContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read edited file: %w", err)
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
