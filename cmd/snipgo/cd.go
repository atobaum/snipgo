package main

import (
	"fmt"

	"snipgo/internal/config"

	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:   "cd",
	Short: "Print the snippets data directory path",
	Long: `Print the path to the snippets data directory.

Use this command to navigate to your snippets directory:

  cd $(snipgo cd)

Or in fish shell:

  cd (snipgo cd)

This is useful for:
  - Manually editing snippet files
  - Running git commands directly
  - Exploring your snippets with other tools`,
	Args: cobra.NoArgs,
	RunE: runCd,
}

func init() {
	rootCmd.AddCommand(cdCmd)
}

func runCd(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Just print the path - user can use cd $(snipgo cd)
	fmt.Println(cfg.DataDirectory)
	return nil
}
