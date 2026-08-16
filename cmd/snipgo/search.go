package main

import (
	"fmt"
	"log/slog"
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
  # Search all snippets (sorted by usage frequency, most used first)
  snipgo search

  # Search with query
  snipgo search -q "docker deploy"

  # Filter by tag
  snipgo search --tag golang

  # Multiple tags (AND logic)
  snipgo search --tag golang --tag web

  # Filter by language
  snipgo search --lang bash

  # Combined filters and query
  snipgo search --tag devops --lang bash -q "deploy"

  # Sort by name (alphabetical)
  snipgo search --sort name

  # Sort by name descending (reverse alphabetical)
  snipgo search --sort name --order desc

  # Sort by most recently used
  snipgo search --sort last_used

  # Sort by least frequently used
  snipgo search --sort frequency --order asc`,
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringP("query", "q", "", "Search query (fuzzy match)")
	searchCmd.Flags().StringSliceP("tag", "t", []string{}, "Filter by tags (repeatable, AND logic)")
	searchCmd.Flags().StringP("language", "L", "", "Filter by language")
	searchCmd.Flags().StringP("lang", "", "", "Alias for --language")
	searchCmd.Flags().StringArrayP("var", "v", []string{}, "Variable value in KEY=VALUE format (repeatable)")
	searchCmd.Flags().Bool("raw", false, "Output raw body without variable expansion")
	searchCmd.Flags().StringP("sort", "s", "frequency", "Sort order: frequency, name, last_used")
	searchCmd.Flags().StringP("order", "o", "", "Sort direction: asc, desc (default depends on sort field)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Parse flags
	query, _ := cmd.Flags().GetString("query")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	language, _ := cmd.Flags().GetString("language")
	lang, _ := cmd.Flags().GetString("lang")
	sortBy, _ := cmd.Flags().GetString("sort")
	sortOrder, _ := cmd.Flags().GetString("order")

	// Handle language alias
	if language == "" && lang != "" {
		language = lang
	}

	// Build usage data map from tracker
	usageData := make(map[string]core.UsageData)
	if usageTracker != nil {
		for id, entry := range usageTracker.GetAll() {
			usageData[id] = core.UsageData{
				Count:    entry.Count,
				LastUsed: entry.LastUsed,
			}
		}
	}

	// Build search options
	opts := core.SearchOptions{
		Query:     query,
		Tags:      tags,
		Language:  language,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		UsageData: usageData,
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

	// Record usage for the selected snippet
	if usageTracker != nil {
		if err := usageTracker.Record(selected.ID); err != nil {
			slog.Warn("failed to record snippet usage", "id", selected.ID, "error", err)
		}
	}

	// Parse variable flags
	varFlags, _ := cmd.Flags().GetStringArray("var")
	raw, _ := cmd.Flags().GetBool("raw")

	providedVars, err := parseVarFlags(varFlags)
	if err != nil {
		return err
	}

	// Check if /dev/tty is available for interactive prompts
	// (works even when stdin is piped or in zle context)
	isTerminal := false
	if tty, err := os.Open("/dev/tty"); err == nil {
		tty.Close()
		isTerminal = true
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
