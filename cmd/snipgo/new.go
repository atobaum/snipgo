package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"snipgo/internal/core"

	"github.com/atotto/clipboard"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var newOpts newOptions

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new snippet",
	Long: `Create a new snippet.

The body is read from the clipboard (--clipboard), from piped stdin, or
interactively. With --title, clipboard and stdin input need no prompts.

Examples:
  # Interactive
  snipgo new

  # Save a command from stdin
  pbpaste | snipgo new -t "List ports" --tags net,mac -L bash

  # Save the clipboard contents
  snipgo new -c -t "List ports"

  # Review stdin in $EDITOR with frontmatter pre-filled
  pbpaste | snipgo new -e -t "List ports" --tags net`,
	Args: cobra.NoArgs,
	RunE: runNew,
}

func init() {
	newCmd.Flags().StringVarP(&newOpts.title, "title", "t", "", "Snippet title (skips the title prompt)")
	newCmd.Flags().StringVar(&newOpts.tags, "tags", "", "Comma-separated tags")
	newCmd.Flags().StringVarP(&newOpts.language, "language", "L", "", "Snippet language")
	newCmd.Flags().StringVar(&newOpts.lang, "lang", "", "Alias for --language")
	newCmd.Flags().BoolVarP(&newOpts.clipboard, "clipboard", "c", false, "Read the body from the clipboard")
	newCmd.Flags().BoolVarP(&newOpts.multiline, "multiline", "m", false, "Enable multiline input (two blank lines to finish)")
	newCmd.Flags().BoolVarP(&newOpts.editor, "editor", "e", false, "Open $EDITOR with frontmatter and body")
}

type newOptions struct {
	title     string
	tags      string
	language  string
	lang      string
	clipboard bool
	multiline bool
	editor    bool
}

type bodySource int

const (
	bodyFromPrompt bodySource = iota
	bodyFromClipboard
	bodyFromStdin
)

type newIO struct {
	stdin         io.Reader
	stdinPiped    bool
	readClipboard func() (string, error)
	openTTY       func() (io.ReadCloser, error)
	promptLine    func(prompt string) (string, error)
	editContent   func(content []byte) ([]byte, error)
}

func runNew(cmd *cobra.Command, args []string) error {
	env := newIO{
		stdin:         os.Stdin,
		stdinPiped:    isStdinPiped(),
		readClipboard: clipboard.ReadAll,
		openTTY:       openTTY,
		promptLine:    readline.Line,
		editContent: func(content []byte) ([]byte, error) {
			return editContentInEditor(content, "snipgo-new-")
		},
	}

	snippet, err := buildNewSnippet(newOpts, env)
	if err != nil {
		return err
	}

	if err := manager.Save(snippet); err != nil {
		return fmt.Errorf("failed to save snippet: %w", err)
	}

	fmt.Printf("Snippet saved: %s\n", snippet.Title)
	return nil
}

func isStdinPiped() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice == 0
}

func chooseBodySource(opts newOptions, stdinPiped bool) (bodySource, error) {
	if opts.multiline && opts.editor {
		return 0, fmt.Errorf("--multiline and --editor flags are mutually exclusive")
	}
	if opts.clipboard && stdinPiped {
		return 0, fmt.Errorf("--clipboard cannot be used with piped stdin")
	}
	if opts.clipboard && opts.multiline {
		return 0, fmt.Errorf("--clipboard and --multiline flags are mutually exclusive")
	}
	if stdinPiped && opts.multiline {
		return 0, fmt.Errorf("--multiline cannot be used with piped stdin")
	}

	switch {
	case opts.clipboard:
		return bodyFromClipboard, nil
	case stdinPiped:
		return bodyFromStdin, nil
	default:
		return bodyFromPrompt, nil
	}
}

func buildNewSnippet(opts newOptions, env newIO) (*core.Snippet, error) {
	source, err := chooseBodySource(opts, env.stdinPiped)
	if err != nil {
		return nil, err
	}

	var body string
	switch source {
	case bodyFromClipboard:
		text, err := env.readClipboard()
		if err != nil {
			return nil, fmt.Errorf("failed to read clipboard: %w", err)
		}
		body = trimSingleTrailingNewline(text)
	case bodyFromStdin:
		data, err := io.ReadAll(env.stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
		body = trimSingleTrailingNewline(string(data))
	}
	if source != bodyFromPrompt && strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	title, err := readTitle(opts, env)
	if err != nil {
		return nil, err
	}

	snippet := core.NewSnippet(title)
	snippet.Tags = parseTagsFlag(opts.tags)
	snippet.Language = opts.language
	if snippet.Language == "" {
		snippet.Language = opts.lang
	}
	snippet.Body = body

	if opts.editor {
		return editNewSnippet(snippet, env)
	}

	if source == bodyFromPrompt {
		command, err := promptCommand(opts, env)
		if err != nil {
			return nil, err
		}
		snippet.Body = command
	}
	return snippet, nil
}

func readTitle(opts newOptions, env newIO) (string, error) {
	if title := strings.TrimSpace(opts.title); title != "" {
		return title, nil
	}

	var title string
	if env.stdinPiped {
		tty, err := env.openTTY()
		if err != nil {
			return "", fmt.Errorf("no terminal available to prompt for a title; pass it with --title (-t)")
		}
		defer tty.Close()
		fmt.Fprint(os.Stderr, "Description> ")
		scanner := bufio.NewScanner(tty)
		if !scanner.Scan() {
			return "", fmt.Errorf("cancelled")
		}
		title = scanner.Text()
	} else {
		line, err := env.promptLine("Description> ")
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("cancelled")
			}
			return "", fmt.Errorf("failed to read description: %w", err)
		}
		title = line
	}

	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("description cannot be empty")
	}
	return title, nil
}

func promptCommand(opts newOptions, env newIO) (string, error) {
	var command string
	if opts.multiline {
		fmt.Println("(Enter two blank lines to finish)")
		lines, err := scanMultiLine("Command> ", ".......> ")
		if err != nil {
			return "", err
		}
		command = lines
	} else {
		line, err := env.promptLine("Command> ")
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("cancelled")
			}
			return "", fmt.Errorf("failed to read command: %w", err)
		}
		command = strings.TrimSpace(line)
	}
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}
	return command, nil
}

func editNewSnippet(draft *core.Snippet, env newIO) (*core.Snippet, error) {
	content, err := serializeSnippetForEdit(draft)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize snippet: %w", err)
	}

	edited, err := env.editContent(content)
	if err != nil {
		return nil, fmt.Errorf("creation cancelled: %w", err)
	}

	snippet, err := parseSnippetFromEdit(edited)
	if err != nil {
		return nil, fmt.Errorf("failed to parse edited content: %w", err)
	}
	snippet.ID = draft.ID
	snippet.CreatedAt = draft.CreatedAt
	snippet.Body = trimSingleTrailingNewline(snippet.Body)

	if err := snippet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid snippet after editing: %w", err)
	}
	if strings.TrimSpace(snippet.Body) == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}
	return snippet, nil
}

func parseTagsFlag(raw string) []string {
	tags := []string{}
	seen := make(map[string]bool)
	for _, part := range strings.Split(raw, ",") {
		tag := strings.TrimSpace(part)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	return tags
}

func trimSingleTrailingNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
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
