package main

import (
	"fmt"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <snippet-id> <commit-hash>",
	Short: "Restore a snippet to a previous version",
	Long: `Restore a snippet to a specific version from git history.

Use 'snipgo history <snippet-id>' to see available versions.
The restored version will be staged but not committed.`,
	Args: cobra.ExactArgs(2),
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
	snippetID := args[0]
	commitHash := args[1]

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

	// Find the snippet
	snippet, err := manager.GetByID(snippetID)
	if err != nil {
		return fmt.Errorf("snippet not found: %s", snippetID)
	}

	// Get the file path
	filename, err := manager.GetFilenameByID(snippetID)
	if err != nil {
		return fmt.Errorf("failed to get filename: %w", err)
	}

	// Restore the file
	if err := gm.RestoreFile(filename, commitHash); err != nil {
		return fmt.Errorf("failed to restore: %w", err)
	}

	// Reload snippets to reflect the change
	if err := manager.LoadAll(); err != nil {
		return fmt.Errorf("failed to reload snippets: %w", err)
	}

	fmt.Printf("Restored '%s' to version %s\n", snippet.Title, commitHash)
	fmt.Println("\nThe file has been restored and staged.")
	fmt.Println("To commit this change: snipgo commit -m \"Restore to previous version\"")
	fmt.Println("To undo this restore: snipgo git checkout HEAD -- <filename>")
	return nil
}
