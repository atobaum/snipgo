package main

import (
	"fmt"
	"strings"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var showDiff bool

var historyCmd = &cobra.Command{
	Use:   "history <snippet-id>",
	Short: "Show version history of a snippet",
	Long: `Show the git commit history for a specific snippet.

Use --diff to show the changes made in each commit.`,
	Args: cobra.ExactArgs(1),
	RunE: runHistory,
}

func init() {
	historyCmd.Flags().BoolVar(&showDiff, "diff", false, "Show diff for each commit")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	snippetID := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	gitCfg := configToGitConfig(cfg.Git)
	gm := git.NewGitManager(cfg.DataDirectory, gitCfg)

	if !gm.IsGitInstalled() {
		return git.ErrGitNotInstalled
	}

	if !gm.IsGitRepo() {
		return git.ErrNotARepository
	}

	// Find the snippet file
	snippet, err := manager.GetByID(snippetID)
	if err != nil {
		return fmt.Errorf("snippet not found: %s", snippetID)
	}

	// Get the file path (relative to repo)
	filename, err := manager.GetFilenameByID(snippetID)
	if err != nil {
		return fmt.Errorf("failed to get filename: %w", err)
	}

	// Get history
	commits, err := gm.GetFileHistory(filename)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(commits) == 0 {
		fmt.Printf("No history found for snippet: %s\n", snippet.Title)
		fmt.Println("The snippet may not have been committed yet.")
		return nil
	}

	fmt.Printf("History for: %s\n", snippet.Title)
	fmt.Printf("File: %s\n\n", filename)
	fmt.Printf("%-12s %-20s %s\n", "COMMIT", "DATE", "MESSAGE")
	fmt.Println(strings.Repeat("-", 60))

	for _, c := range commits {
		dateStr := c.Date.Format("2006-01-02 15:04")
		fmt.Printf("%-12s %-20s %s\n", c.ShortHash, dateStr, c.Message)

		if showDiff {
			diff, err := gm.GetFileDiff(filename, c.Hash)
			if err == nil && diff != "" {
				fmt.Println()
				// Indent diff output
				for _, line := range strings.Split(diff, "\n") {
					fmt.Printf("    %s\n", line)
				}
				fmt.Println()
			}
		}
	}

	fmt.Printf("\nTo restore a version: snipgo restore %s <commit-hash>\n", snippetID)
	return nil
}
