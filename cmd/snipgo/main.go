package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"snipgo/internal/core"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var manager *core.Manager

var rootCmd = &cobra.Command{
	Use:   "snipgo",
	Short: "SnipGo - Local-First Snippet Manager",
	Long:  "SnipGo is a local-first snippet manager that stores snippets as Markdown files.",
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new snippet",
	Long:  "Opens $EDITOR to create a new snippet with basic frontmatter",
	RunE:  runAdd,
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
	Long:  "Search snippets by title (fuzzy), tags, or body content",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

var copyCmd = &cobra.Command{
	Use:   "copy [query]",
	Short: "Copy snippet body to clipboard",
	Long:  "Searches for a snippet and copies its body to the clipboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopy,
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(copyCmd)
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

func runAdd(cmd *cobra.Command, args []string) error {
	// Create a new snippet
	snippet := core.NewSnippet("Untitled")

	// Create a temporary file with frontmatter
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("snipgo_%s.md", snippet.ID))

	// Serialize to markdown
	content, err := serializeSnippetForEdit(snippet)
	if err != nil {
		return fmt.Errorf("failed to serialize snippet: %w", err)
	}

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim
	}

	editorCmd := exec.Command(editor, tmpFile)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the edited file
	editedContent, err := os.ReadFile(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Parse the edited content
	editedSnippet, err := parseSnippetFromEdit(editedContent)
	if err != nil {
		return fmt.Errorf("failed to parse edited snippet: %w", err)
	}

	// Save the snippet
	if err := manager.Save(editedSnippet); err != nil {
		return fmt.Errorf("failed to save snippet: %w", err)
	}

	fmt.Printf("Snippet saved: %s\n", editedSnippet.Title)
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
	query := args[0]
	results := manager.Search(query)

	if len(results) == 0 {
		fmt.Printf("No snippets found for query: %s\n", query)
		return nil
	}

	fmt.Printf("Found %d snippet(s):\n\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. [%s] %s\n", i+1, result.Snippet.ID[:8], result.Snippet.Title)
		if len(result.Snippet.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(result.Snippet.Tags, ", "))
		}
		if result.Snippet.Language != "" {
			fmt.Printf("   Language: %s\n", result.Snippet.Language)
		}
		fmt.Println()
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
	var parts []string
	parts = append(parts, "---")
	parts = append(parts, fmt.Sprintf("id: %q", snippet.ID))
	parts = append(parts, fmt.Sprintf("title: %q", snippet.Title))
	parts = append(parts, fmt.Sprintf("tags: []"))
	parts = append(parts, fmt.Sprintf("language: \"\""))
	parts = append(parts, fmt.Sprintf("is_favorite: %v", snippet.IsFavorite))
	parts = append(parts, fmt.Sprintf("created_at: %s", snippet.CreatedAt.Format("2006-01-02T15:04:05Z")))
	parts = append(parts, fmt.Sprintf("updated_at: %s", snippet.UpdatedAt.Format("2006-01-02T15:04:05Z")))
	parts = append(parts, "---")
	parts = append(parts, "")
	parts = append(parts, snippet.Body)

	return []byte(strings.Join(parts, "\n")), nil
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
