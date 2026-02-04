package main

import (
	"fmt"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push commits to remote repository",
	Long: `Push local commits to the configured remote repository.

This pushes all unpushed commits to the remote. Make sure you have
committed your changes first with 'snipgo commit'.`,
	Args: cobra.NoArgs,
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
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

	if !gm.HasRemote() {
		return git.ErrNoRemote
	}

	// Check for unpushed commits
	status, err := gm.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.UnpushedCommits == 0 {
		fmt.Println("Everything up-to-date")
		return nil
	}

	fmt.Printf("Pushing %d commit(s) to %s...\n", status.UnpushedCommits, status.RemoteURL)

	if err := gm.Push(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	fmt.Println("Done!")
	return nil
}
