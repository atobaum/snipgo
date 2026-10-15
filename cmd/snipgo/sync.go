package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"snipgo/internal/config"
	"snipgo/internal/gitsync"

	"github.com/spf13/cobra"
)

const syncTimeout = 2 * time.Minute

var syncCmd = &cobra.Command{
	Use:          "sync",
	Short:        "Sync snippets with the git remote",
	Long:         "Commit snippet changes, pull from the upstream branch and push. The snippet directory must be in a git repository with an upstream branch.",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE:         runSync,
}

type syncer interface {
	Sync(ctx context.Context) (gitsync.Result, error)
}

func runSync(cmd *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, syncTimeout)
	defer cancel()
	s := gitsync.New(gitsync.ExecRunner{}, gitsync.Options{Dir: cfg.DataDirectory, Conflict: gitsync.Mode(cfg.Sync.Conflict)})
	return syncAndReport(ctx, s, cfg.DataDirectory, cmd.OutOrStdout())
}

func syncAndReport(ctx context.Context, s syncer, dir string, out io.Writer) error {
	res, err := s.Sync(ctx)
	if err != nil {
		var ce *gitsync.ConflictError
		if errors.As(err, &ce) {
			return fmt.Errorf("sync conflict in %d file(s): %s. The merge was aborted and nothing was pushed. To resolve, run git -C %q pull, fix the conflicted files, commit, then run snipgo sync again", len(ce.Files), strings.Join(ce.Files, ", "), dir)
		}
		_, msg := gitsync.Describe(err)
		return errors.New(msg)
	}
	fmt.Fprintln(out, gitsync.Summary(res))
	return nil
}
