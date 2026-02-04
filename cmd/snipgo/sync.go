package main

import (
	"errors"
	"fmt"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync snippets with remote repository",
	Long: `Sync snippets with the remote repository (pull + push).

This command first pulls changes from the remote, then pushes any
local commits. It's equivalent to running 'snipgo pull' followed
by 'snipgo push'.`,
	Args: cobra.NoArgs,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
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

	remoteURL := gm.GetRemoteURL()

	// Pull first
	fmt.Printf("Pulling from %s...\n", remoteURL)
	if err := gm.Pull(); err != nil {
		if errors.Is(err, git.ErrMergeConflict) {
			fmt.Println("\nMerge conflict detected!")
			fmt.Println("\nTo resolve:")
			fmt.Printf("  1. cd %s\n", cfg.DataDirectory)
			fmt.Println("  2. git status  # Check conflicted files")
			fmt.Println("  3. Edit conflicted files to resolve")
			fmt.Println("  4. git add . && git commit -m \"Resolve conflicts\"")
			fmt.Println("  5. snipgo sync  # Try again")
			return err
		}
		return fmt.Errorf("pull failed: %w", err)
	}

	// Check if there's anything to push
	status, err := gm.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.UnpushedCommits == 0 {
		fmt.Println("Already up to date.")
		return nil
	}

	// Push
	fmt.Printf("Pushing %d commit(s) to %s...\n", status.UnpushedCommits, remoteURL)
	if err := gm.Push(); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	fmt.Println("Sync complete!")
	return nil
}
