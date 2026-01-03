package main

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy [query]",
	Short: "Copy snippet body to clipboard",
	Long:  "Searches for a snippet and copies its body to the clipboard",
	Args:  cobra.ExactArgs(1),
	RunE:  runCopy,
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
