package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"snipgo/internal/config"
	"snipgo/internal/core"
	"snipgo/internal/history"

	"github.com/spf13/cobra"
)

type cliApp struct {
	manager    *core.Manager
	varHistory *history.VarHistory
}

var app *cliApp

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var logLevel string

func newCLIApp() (*cliApp, error) {
	m, err := core.NewDefaultManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize manager: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	vh, err := history.NewVarHistory(cfg.DataDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize variable history: %w", err)
	}

	return &cliApp{
		manager:    m,
		varHistory: vh,
	}, nil
}

var rootCmd = &cobra.Command{
	Use:     "snipgo",
	Short:   "SnipGo - Local-First Snippet Manager",
	Long:    "SnipGo is a local-first snippet manager that stores snippets as Markdown files.",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, _ := cmd.Flags().GetString("log-level")
		setupLogger(level)

		var err error
		app, err = newCLIApp()
		if err != nil {
			return err
		}
		return nil
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
