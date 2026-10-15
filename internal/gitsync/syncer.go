package gitsync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const cleanupTimeout = 10 * time.Second

type Mode string

const (
	ModeManual Mode = "manual"
	ModeLocal  Mode = "local"
	ModeRemote Mode = "remote"
)

type Options struct {
	Dir      string
	Conflict Mode
	Hostname string
	Now      func() time.Time
	LookPath func(string) (string, error)
}

type Result struct {
	Committed bool
	Pulled    bool
	Pushed    bool
}

type Syncer struct {
	runner Runner
	opts   Options
}

func New(r Runner, opts Options) *Syncer {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.Hostname == "" {
		opts.Hostname, _ = os.Hostname()
	}
	if opts.Conflict == "" {
		opts.Conflict = ModeManual
	}
	return &Syncer{runner: r, opts: opts}
}

func (s *Syncer) git(ctx context.Context, op string, args ...string) (string, error) {
	stdout, stderr, err := s.runner.Run(ctx, s.opts.Dir, args...)
	if err == nil {
		return stdout, nil
	}
	if strings.Contains(stderr, "index.lock") {
		return stdout, ErrLocked
	}
	return stdout, &GitError{Op: op, Stderr: stderr, Err: err}
}

// Sync runs one commit, pull, push cycle. A canceled or expired context is returned as is.
func (s *Syncer) Sync(ctx context.Context) (Result, error) {
	res, err := s.sync(ctx)
	if err != nil && ctx.Err() != nil {
		return res, ctx.Err()
	}
	return res, err
}

func (s *Syncer) sync(ctx context.Context) (Result, error) {
	var res Result
	if _, err := s.opts.LookPath("git"); err != nil {
		return res, ErrGitNotFound
	}
	if _, err := s.git(ctx, "status", "rev-parse", "--show-toplevel"); err != nil {
		return res, ErrNotRepository
	}
	if _, err := s.git(ctx, "status", "rev-parse", "-q", "--verify", "MERGE_HEAD"); err == nil {
		return res, ErrMergeInProgress
	}
	if _, err := s.git(ctx, "status", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err != nil {
		return res, ErrNoUpstream
	}

	if _, err := s.git(ctx, "add", "add", "-A", "--", "."); err != nil {
		return res, err
	}
	staged, err := s.git(ctx, "status", "diff", "--cached", "--name-only", "--", ".")
	if err != nil {
		return res, err
	}
	if strings.TrimSpace(staged) != "" {
		if _, err := s.git(ctx, "commit", "commit", "-m", s.commitMessage(), "--", "."); err != nil {
			return res, err
		}
		res.Committed = true
	}

	before, err := s.git(ctx, "status", "rev-parse", "HEAD")
	if err != nil {
		return res, err
	}
	if err := s.pull(ctx); err != nil {
		return res, err
	}
	after, err := s.git(ctx, "status", "rev-parse", "HEAD")
	if err != nil {
		return res, err
	}
	res.Pulled = strings.TrimSpace(before) != strings.TrimSpace(after)

	ahead, err := s.git(ctx, "status", "rev-list", "--count", "@{u}..HEAD")
	if err != nil {
		return res, err
	}
	if n, _ := strconv.Atoi(strings.TrimSpace(ahead)); n > 0 {
		if _, err := s.git(ctx, "push", "push"); err != nil {
			return res, err
		}
		res.Pushed = true
	}
	return res, nil
}

// -X cannot resolve modify/delete conflicts, so conflicts are aborted in every mode.
func (s *Syncer) pull(ctx context.Context) error {
	args := []string{"pull", "--no-rebase", "--no-edit"}
	switch s.opts.Conflict {
	case ModeLocal:
		args = append(args, "-X", "ours")
	case ModeRemote:
		args = append(args, "-X", "theirs")
	}
	_, pullErr := s.git(ctx, "pull", args...)
	if pullErr == nil {
		return nil
	}
	// The merge must be aborted even when ctx already ended, or the next run sees a merge in progress.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupTimeout)
	defer cancel()
	out, err := s.git(cleanupCtx, "status", "-c", "core.quotePath=false", "diff", "--name-only", "--diff-filter=U", "--", ".")
	if err != nil {
		return pullErr
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	if len(files) == 0 {
		return pullErr
	}
	if _, err := s.git(cleanupCtx, "merge", "merge", "--abort"); err != nil {
		return err
	}
	return &ConflictError{Files: files}
}

func (s *Syncer) commitMessage() string {
	return fmt.Sprintf("snipgo sync: %s (%s)", s.opts.Now().Format("2006-01-02 15:04"), s.opts.Hostname)
}
