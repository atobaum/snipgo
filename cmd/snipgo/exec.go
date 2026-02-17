package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"snipgo/internal/tmpl"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a snippet interactively",
	Long:  "Interactively select a snippet using fzf and execute its body as a shell command",
	Args:  cobra.NoArgs,
	RunE:  runExec,
}

func init() {
	execCmd.Flags().StringArrayP("var", "v", []string{}, "Variable value in KEY=VALUE format (repeatable)")
	execCmd.Flags().Bool("raw", false, "Output raw body without variable expansion")
	execCmd.Flags().Bool("preview", false, "Show preview and confirm before execution (auto-enabled with variables)")
	execCmd.Flags().Bool("no-preview", false, "Disable preview even when variables present")
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

	// Parse flags
	varFlags, _ := cmd.Flags().GetStringArray("var")
	raw, _ := cmd.Flags().GetBool("raw")
	preview, _ := cmd.Flags().GetBool("preview")
	noPreview, _ := cmd.Flags().GetBool("no-preview")

	// Parse variable flags
	providedVars, err := parseVarFlags(varFlags)
	if err != nil {
		return err
	}

	// Check if snippet has variables
	bodyVarNames := tmpl.ExtractVariables(selected.Body)
	hasVariables := len(bodyVarNames) > 0

	// Auto-enable preview if variables present (unless --no-preview set)
	if hasVariables && !noPreview {
		preview = true
	}

	// Expand body with variables (interactive prompts if needed)
	expandedBody, err := expandSnippetBody(selected, providedVars, raw, true)
	if err != nil {
		return err
	}

	// Show preview and confirm if variables or --preview flag
	if preview {
		fmt.Fprintf(os.Stderr, "\n=== PREVIEW ===\n%s\n===============\n\n", expandedBody)

		if hasVariables {
			fmt.Fprintf(os.Stderr, "WARNING: Variable values are interpolated directly into the shell command. Review carefully.\n\n")
		}

		fmt.Fprintf(os.Stderr, "Execute this command? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return fmt.Errorf("execution cancelled")
		}
	}

	// Execute body as shell command
	execCmd := exec.Command("sh", "-c", expandedBody)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
