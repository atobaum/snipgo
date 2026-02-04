package main

import (
	"fmt"
	"strings"

	"snipgo/internal/config"
	"snipgo/internal/git"

	"github.com/spf13/cobra"
)

var gitManager *git.GitManager

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git integration commands",
	Long: `Git integration commands for version control of snippets.

Use 'snipgo git init' to initialize a new repository, or pass any git
command directly: 'snipgo git status', 'snipgo git log', etc.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load config to get data directory and git settings
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Convert config.GitConfig to git.GitConfig
		var gitCfg *git.GitConfig
		if cfg.Git != nil {
			gitCfg = &git.GitConfig{
				Enabled:               cfg.Git.Enabled,
				AutoCommit:            cfg.Git.AutoCommit,
				AutoPush:              cfg.Git.AutoPush,
				CommitMessageTemplate: cfg.Git.CommitMessageTemplate,
				Remote:                cfg.Git.Remote,
				Branch:                cfg.Git.Branch,
			}
		}

		// Create git manager
		gitManager = git.NewGitManager(cfg.DataDirectory, gitCfg)

		// Check if git is installed
		if !gitManager.IsGitInstalled() {
			return git.ErrGitNotInstalled
		}

		return nil
	},
	RunE: runGitPassthrough,
}

var gitInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a git repository for snippets",
	Long: `Initialize a new git repository in the snippets directory.

This enables version control for your snippets, allowing you to track
changes, sync across devices, and restore previous versions.`,
	Args: cobra.NoArgs,
	RunE: runGitInit,
}

var gitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show git status of snippets",
	Long:  "Show the current git status of your snippets directory.",
	Args:  cobra.NoArgs,
	RunE:  runGitStatus,
}

var gitCloneCmd = &cobra.Command{
	Use:   "clone <url>",
	Short: "Clone a remote snippets repository",
	Long: `Clone a remote git repository to your snippets directory.

This will replace your current snippets directory with the cloned repository.
Make sure to backup any existing snippets before running this command.`,
	Args: cobra.ExactArgs(1),
	RunE: runGitClone,
}

func init() {
	gitCmd.AddCommand(gitInitCmd)
	gitCmd.AddCommand(gitStatusCmd)
	gitCmd.AddCommand(gitCloneCmd)
	rootCmd.AddCommand(gitCmd)
}

func runGitInit(cmd *cobra.Command, args []string) error {
	if err := gitManager.InitRepo(); err != nil {
		return err
	}

	fmt.Printf("Initialized git repository at %s\n", gitManager.GetRepoPath())
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add a remote: snipgo git remote add origin <url>")
	fmt.Println("  2. Create your first commit: snipgo commit -m \"Initial commit\"")
	fmt.Println("  3. Push to remote: snipgo push")
	return nil
}

func runGitStatus(cmd *cobra.Command, args []string) error {
	status, err := gitManager.Status()
	if err != nil {
		return err
	}

	if !status.IsRepo {
		fmt.Println("Not a git repository.")
		fmt.Println("Run 'snipgo git init' to initialize.")
		return nil
	}

	fmt.Printf("On branch %s\n", status.Branch)

	if status.HasRemote {
		fmt.Printf("Remote: %s\n", status.RemoteURL)
		if status.UnpushedCommits > 0 {
			fmt.Printf("Your branch is ahead by %d commit(s).\n", status.UnpushedCommits)
		}
	} else {
		fmt.Println("No remote configured.")
	}

	if !status.HasChanges {
		fmt.Println("\nNothing to commit, working tree clean")
		return nil
	}

	fmt.Println()

	if len(status.StagedFiles) > 0 {
		fmt.Println("Changes to be committed:")
		for _, f := range status.StagedFiles {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	if len(status.ModifiedFiles) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, f := range status.ModifiedFiles {
			fmt.Printf("  modified: %s\n", f)
		}
		fmt.Println()
	}

	if len(status.UntrackedFiles) > 0 {
		fmt.Println("Untracked files:")
		for _, f := range status.UntrackedFiles {
			fmt.Printf("  %s\n", f)
		}
	}

	return nil
}

func runGitClone(cmd *cobra.Command, args []string) error {
	url := args[0]

	fmt.Printf("Cloning %s...\n", url)
	if err := gitManager.CloneRepo(url); err != nil {
		return err
	}

	// Enable git in config after successful clone
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Git == nil {
		cfg.Git = config.DefaultGitConfig()
	}
	cfg.Git.Enabled = true

	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Cloned repository to %s\n", gitManager.GetRepoPath())
	fmt.Println("Git integration enabled.")
	return nil
}

func runGitPassthrough(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	// Pass through to git
	output, err := gitManager.Exec(args...)
	if err != nil {
		// Still print output as it may contain useful info
		if output != "" {
			fmt.Print(output)
		}
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}

	fmt.Print(output)
	return nil
}
