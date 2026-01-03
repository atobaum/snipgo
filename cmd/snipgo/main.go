package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"snipgo/internal/core"

	"github.com/atotto/clipboard"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var manager *core.Manager

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "snipgo",
	Short:   "SnipGo - Local-First Snippet Manager",
	Long:    "SnipGo is a local-first snippet manager that stores snippets as Markdown files.",
	Version: version,
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new snippet",
	Long:  "Interactively create a new snippet by entering description and command",
	Args:  cobra.NoArgs,
	RunE:  runNew,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snippets",
	Long:  "Lists all snippets in a table format",
	RunE:  runList,
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search snippets",
	Long:  "Interactively search and select snippets using fzf. If query is provided, filters results first.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSearch,
}

var copyCmd = &cobra.Command{
	Use:   "copy [query]",
	Short: "Copy snippet body to clipboard",
	Long:  "Searches for a snippet and copies its body to the clipboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopy,
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a snippet interactively",
	Long:  "Interactively select a snippet using fzf and execute its body as a shell command",
	Args:  cobra.NoArgs,
	RunE:  runExec,
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a snippet",
	Long:  "Interactively select a snippet using fzf and edit it with $EDITOR",
	Args:  cobra.NoArgs,
	RunE:  runEdit,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version, commit hash, and build date",
	Args:  cobra.NoArgs,
	RunE:  runVersion,
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate completion script",
	Long: `Generate shell completion scripts for snipgo.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, execute once:
  # Linux:
  snipgo completion zsh > "${fpath[1]}/_snipgo"
  
  # macOS:
  snipgo completion zsh > $(brew --prefix)/share/zsh/site-functions/_snipgo
`,
	Args: cobra.NoArgs,
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	Long: `Generate the autocompletion script for zsh shell.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, add to your ~/.zshrc:
  echo 'source <(snipgo completion zsh)' >> ~/.zshrc

Or install to a system-wide location:
  snipgo completion zsh > ~/.zsh/completions/_snipgo
  echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  echo 'autoload -U compinit && compinit' >> ~/.zshrc
`,
	Args: cobra.NoArgs,
	RunE: runCompletionZsh,
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(versionCmd)
	completionCmd.AddCommand(completionZshCmd)
	rootCmd.AddCommand(completionCmd)
}

func main() {
	var err error
	manager, err = core.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize manager: %v\n", err)
		os.Exit(1)
	}

	if err := manager.LoadAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load snippets: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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

func runList(cmd *cobra.Command, args []string) error {
	snippets := manager.GetAll()

	if len(snippets) == 0 {
		fmt.Println("No snippets found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tTags\tLanguage\tFavorite")
	fmt.Fprintln(w, "---\t-----\t----\t--------\t--------")

	for _, snippet := range snippets {
		idShort := snippet.ID[:8]
		tags := strings.Join(snippet.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		language := snippet.Language
		if language == "" {
			language = "-"
		}
		favorite := "No"
		if snippet.IsFavorite {
			favorite = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			idShort, snippet.Title, tags, language, favorite)
	}

	return w.Flush()
}

func runSearch(cmd *cobra.Command, args []string) error {
	var snippets []*core.Snippet

	if len(args) > 0 {
		// Search with query
		query := args[0]
		results := manager.Search(query)
		if len(results) == 0 {
			fmt.Printf("No snippets found for query: %s\n", query)
			return nil
		}
		// Convert SearchResult to Snippet
		snippets = make([]*core.Snippet, len(results))
		for i, result := range results {
			snippets[i] = result.Snippet
		}
	} else {
		// No query, use all snippets
		snippets = manager.GetAll()
		if len(snippets) == 0 {
			fmt.Println("No snippets found.")
			return nil
		}
	}

	// Use fzf to select
	selected, err := selectSnippetWithFzf(snippets)
	if err != nil {
		return err
	}

	// Output body to stdout
	fmt.Print(selected.Body)
	return nil
}

func runExec(cmd *cobra.Command, args []string) error {
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

	// Execute body as shell command
	execCmd := exec.Command("sh", "-c", selected.Body)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func runCopy(cmd *cobra.Command, args []string) error {
	query := args[0]
	results := manager.Search(query)

	if len(results) == 0 {
		return fmt.Errorf("no snippets found for query: %s", query)
	}

	// Get the top result
	topResult := results[0]
	body := topResult.Snippet.Body

	if err := clipboard.WriteAll(body); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("Copied body of snippet '%s' to clipboard\n", topResult.Snippet.Title)
	return nil
}

// serializeSnippetForEdit creates a markdown file with frontmatter for editing
func serializeSnippetForEdit(snippet *core.Snippet) ([]byte, error) {
	// Use the core package's SerializeFrontmatter function
	return core.SerializeFrontmatter(snippet)
}

// parseSnippetFromEdit parses a markdown file edited by the user
func parseSnippetFromEdit(content []byte) (*core.Snippet, error) {
	// Use the core package to parse frontmatter
	parsedSnippet, err := core.ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return parsedSnippet, nil
}

// formatSnippetForFzf formats a snippet for fzf display in pet CLI style: [Title] Body #tag1 #tag2
func formatSnippetForFzf(snippet *core.Snippet) string {
	// Get first line of body for display
	bodyFirstLine := ""
	if snippet.Body != "" {
		bodyLines := strings.Split(snippet.Body, "\n")
		bodyFirstLine = bodyLines[0]
		if len(bodyFirstLine) > 100 {
			bodyFirstLine = bodyFirstLine[:100] + "..."
		}
	}

	// Format: [Title] Body
	line := fmt.Sprintf("[%s] %s", snippet.Title, bodyFirstLine)

	// Add tags if present
	if len(snippet.Tags) > 0 {
		tagStr := strings.Join(snippet.Tags, " #")
		line += " #" + tagStr
	}

	return line
}

// selectSnippetWithFzf uses fzf to interactively select a snippet from the given list
func selectSnippetWithFzf(snippets []*core.Snippet) (*core.Snippet, error) {
	if len(snippets) == 0 {
		return nil, fmt.Errorf("no snippets to select from")
	}

	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, fmt.Errorf("fzf is not installed. Please install fzf first: https://github.com/junegunn/fzf")
	}

	// Build fzf input: format each snippet
	var fzfInput strings.Builder
	snippetMap := make(map[string]*core.Snippet) // Map from formatted line to snippet
	for _, snippet := range snippets {
		formatted := formatSnippetForFzf(snippet)
		fzfInput.WriteString(formatted)
		fzfInput.WriteString("\n")
		snippetMap[formatted] = snippet
	}

	// Run fzf
	cmd := exec.Command("fzf", "--ansi", "--height", "40%")
	cmd.Stdin = strings.NewReader(fzfInput.String())
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		// fzf returns exit code 1 when user cancels (ESC)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, fmt.Errorf("fzf error: %w", err)
	}

	// Parse selected line
	selectedLine := strings.TrimSpace(string(output))
	if selectedLine == "" {
		return nil, fmt.Errorf("no snippet selected")
	}

	// Find snippet by matching the selected line
	selectedSnippet, found := snippetMap[selectedLine]
	if !found {
		// If exact match not found, try to extract title from the line
		// Format: [Title] Body #tags
		if strings.HasPrefix(selectedLine, "[") {
			endIdx := strings.Index(selectedLine, "]")
			if endIdx > 0 {
				title := selectedLine[1:endIdx]
				// Find snippet by title
				for _, snippet := range snippets {
					if snippet.Title == title {
						return snippet, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("could not find selected snippet")
	}

	return selectedSnippet, nil
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

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("snipgo version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("date: %s\n", date)
	return nil
}

func runCompletionZsh(cmd *cobra.Command, args []string) error {
	return rootCmd.GenZshCompletion(os.Stdout)
}
