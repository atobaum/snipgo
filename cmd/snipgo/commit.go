package main

import (
	"fmt"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit all snippet changes",
	Long: `Commit all changes in the snippets directory.

If no message is provided, a default message will be generated.
This command stages all changes and creates a commit.`,
	Args: cobra.NoArgs,
	RunE: runCommit,
}

func init() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
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

	// Check if there are changes to commit
	status, err := gm.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if !status.HasChanges {
		fmt.Println("Nothing to commit, working tree clean")
		return nil
	}

	// Use provided message or default
	msg := commitMessage
	if msg == "" {
		msg = "Update snippets"
	}

	// Commit all changes
	if err := gm.CommitAll(msg); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Count changes
	totalChanges := len(status.StagedFiles) + len(status.ModifiedFiles) + len(status.UntrackedFiles)
	fmt.Printf("Committed: %d file(s) changed\n", totalChanges)

	// Auto-push if enabled
	if gitCfg != nil && gitCfg.AutoPush && gm.HasRemote() {
		fmt.Println("Auto-pushing to remote...")
		if err := gm.Push(); err != nil {
			fmt.Printf("Warning: auto-push failed: %v\n", err)
		} else {
			fmt.Println("Pushed successfully")
		}
	}

	return nil
}

// configToGitConfig converts config.GitConfig to git.GitConfig
func configToGitConfig(cfg *config.GitConfig) *git.GitConfig {
	if cfg == nil {
		return nil
	}
	return &git.GitConfig{
		Enabled:               cfg.Enabled,
		AutoCommit:            cfg.AutoCommit,
		AutoPush:              cfg.AutoPush,
		CommitMessageTemplate: cfg.CommitMessageTemplate,
		Remote:                cfg.Remote,
		Branch:                cfg.Branch,
	}
}
