package main

import (
	"fmt"
	"os"

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

func init() {
	copyCmd.Flags().StringArrayP("var", "v", []string{}, "Variable value in KEY=VALUE format (repeatable)")
	copyCmd.Flags().Bool("raw", false, "Copy raw body without variable expansion")
}

func runCopy(cmd *cobra.Command, args []string) error {
	query := args[0]
	results := manager.Search(query)

	if len(results) == 0 {
		return fmt.Errorf("no snippets found for query: %s", query)
	}

	// Get the top result
	topResult := results[0]

	// Parse flags
	varFlags, _ := cmd.Flags().GetStringArray("var")
	raw, _ := cmd.Flags().GetBool("raw")

	// Parse variable flags
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
	expandedBody, err := expandSnippetBody(topResult.Snippet, providedVars, raw, isTerminal, varHistory)
	if err != nil {
		return err
	}

	if err := clipboard.WriteAll(expandedBody); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	fmt.Printf("Copied body of snippet '%s' to clipboard\n", topResult.Snippet.Title)
	return nil
}
