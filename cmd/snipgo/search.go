package main

import (
	"fmt"
	"os"

	"snipgo/internal/core"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search snippets",
	Long: `Interactively search and select snippets using fzf.

Filters are applied first to narrow down the snippet set, then the query
performs fuzzy matching on the filtered results.

Examples:
  # Search all snippets
  snipgo search -q "docker deploy"

  # Filter by tag
  snipgo search --tag golang

  # Multiple tags (AND logic)
  snipgo search --tag golang --tag web

  # Filter by language
  snipgo search --lang bash

  # Combined filters and query
  snipgo search --tag devops --lang bash -q "deploy"

  # Filters only (no query) - lists all matching snippets
  snipgo search --tag golang`,
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringP("query", "q", "", "Search query (fuzzy match)")
	searchCmd.Flags().StringSliceP("tag", "t", []string{}, "Filter by tags (repeatable, AND logic)")
	searchCmd.Flags().StringP("language", "L", "", "Filter by language")
	searchCmd.Flags().StringP("lang", "", "", "Alias for --language")
	searchCmd.Flags().StringArrayP("var", "v", []string{}, "Variable value in KEY=VALUE format (repeatable)")
	searchCmd.Flags().Bool("raw", false, "Output raw body without variable expansion")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Parse flags
	query, _ := cmd.Flags().GetString("query")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	language, _ := cmd.Flags().GetString("language")
	lang, _ := cmd.Flags().GetString("lang")

	// Handle language alias
	if language == "" && lang != "" {
		language = lang
	}

	// Build search options
	opts := core.SearchOptions{
		Query:    query,
		Tags:     tags,
		Language: language,
	}

	// Perform search with filters
	results := manager.SearchWithFilters(opts)

	if len(results) == 0 {
		return nil
	}

	// Convert SearchResult to Snippet
	snippets := make([]*core.Snippet, len(results))
	for i, result := range results {
		snippets[i] = result.Snippet
	}

	// Use fzf to select
	selected, err := selectSnippetWithFzf(snippets)
	if err != nil {
		return err
	}

	// Parse variable flags
	varFlags, _ := cmd.Flags().GetStringArray("var")
	raw, _ := cmd.Flags().GetBool("raw")

	providedVars, err := parseVarFlags(varFlags)
	if err != nil {
		return err
	}

	// Check if running in terminal for interactive prompts
	isTerminal := true
	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		isTerminal = false
	}

	// Expand body with variables
	expandedBody, err := expandSnippetBody(selected, providedVars, raw, isTerminal, varHistory)
	if err != nil {
		return err
	}

	// Output body to stdout
	fmt.Print(expandedBody)
	return nil
}
