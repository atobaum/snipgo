package main

import (
	"fmt"
	"strings"

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
		// Print helpful message showing active filters
		msg := "No snippets found"
		if len(tags) > 0 || language != "" || query != "" {
			msg += " matching:\n"
			if len(tags) > 0 {
				msg += fmt.Sprintf("  Tags: %s (AND)\n", strings.Join(tags, ", "))
			}
			if language != "" {
				msg += fmt.Sprintf("  Language: %s\n", language)
			}
			if query != "" {
				msg += fmt.Sprintf("  Query: %q\n", query)
			}
		}
		fmt.Println(msg)
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

	// Output body to stdout
	fmt.Print(selected.Body)
	return nil
}

