package main

import (
	"fmt"

	"snipgo/internal/core"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search snippets",
	Long:  "Interactively search and select snippets using fzf. If query is provided, filters results first.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runSearch,
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

