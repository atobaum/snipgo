package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a snippet interactively",
	Long:  "Interactively select a snippet using fzf and execute its body as a shell command",
	Args:  cobra.NoArgs,
	RunE:  runExec,
}

func runExec(cmd *cobra.Command, args []string) error {
	// Get all snippets
	snippets := manager.GetAll()
	if len(snippets) == 0 {
		return fmt.Errorf("no snippets found")
	}

	// Use fzf to select
	selected, err := selectSnippetWithFzf(snippets)
	if err != nil {
		return err
	}

	// Execute body as shell command
	execCmd := exec.Command("sh", "-c", selected.Body)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

