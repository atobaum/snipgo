package main

import (
	"errors"
	"fmt"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from remote repository",
	Long: `Pull changes from the configured remote repository.

This fetches and merges changes from the remote. If there are merge
conflicts, you'll need to resolve them manually.`,
	Args: cobra.NoArgs,
	RunE: runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("Pulling from %s...\n", gm.GetRemoteURL())

	if err := gm.Pull(); err != nil {
		if errors.Is(err, git.ErrMergeConflict) {
			fmt.Println("\nMerge conflict detected!")
			fmt.Println("\nTo resolve:")
			fmt.Printf("  1. cd %s\n", cfg.DataDirectory)
			fmt.Println("  2. git status  # Check conflicted files")
			fmt.Println("  3. Edit conflicted files to resolve")
			fmt.Println("  4. git add . && git commit -m \"Resolve conflicts\"")
			fmt.Println("  5. snipgo push")
			return err
		}
		return fmt.Errorf("pull failed: %w", err)
	}

	fmt.Println("Done!")
	return nil
}
