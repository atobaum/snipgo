package main

import (
	"fmt"
	"os"

	"snipgo/internal/core"

	"github.com/spf13/cobra"
)

var manager *core.Manager

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "snipgo",
	Short:   "SnipGo - Local-First Snippet Manager",
	Long:    "SnipGo is a local-first snippet manager that stores snippets as Markdown files.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(versionCmd)
	completionCmd.AddCommand(completionZshCmd)
	rootCmd.AddCommand(completionCmd)
}

func main() {
	var err error
	manager, err = core.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize manager: %v\n", err)
		os.Exit(1)
	}

	if err := manager.LoadAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load snippets: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
