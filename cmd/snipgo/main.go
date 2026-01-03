package main

import (
	"log/slog"
	"os"
	"strings"

	"snipgo/internal/core"

	"github.com/spf13/cobra"
)

var manager *core.Manager

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var logLevel string

var rootCmd = &cobra.Command{
	Use:     "snipgo",
	Short:   "SnipGo - Local-First Snippet Manager",
	Long:    "SnipGo is a local-first snippet manager that stores snippets as Markdown files.",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level, _ := cmd.Flags().GetString("log-level")
		setupLogger(level)

		// Initialize manager once
		var err error
		manager, err = core.NewManager()
		if err != nil {
			slog.Error("failed to initialize manager", "error", err)
			os.Exit(1)
		}

		if err := manager.LoadAll(); err != nil {
			slog.Error("failed to load snippets", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Set log level (debug, info, warn, error)")

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

// setupLogger configures the default logger with the specified log level
func setupLogger(levelStr string) {
	if levelStr == "" {
		levelStr = "info" // default level
	}

	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	// Set default log level before PersistentPreRun (for early error logging)
	setupLogger("info")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
