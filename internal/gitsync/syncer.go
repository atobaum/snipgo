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

// Matches what storage.ListFiles loads: *.md files outside hidden directories.
var snippetPathspecs = []string{"*.[mM][dD]", ":(exclude,glob)**/.*/**", ":(exclude,glob).*/**"}

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

func (s *Syncer) precondition(ctx context.Context, kind error, args ...string) error {
	_, stderr, err := s.runner.Run(ctx, s.opts.Dir, args...)
	if err == nil {
		return nil
	}
	return &GitError{Op: "status", Stderr: stderr, Err: kind}
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
	if err := s.precondition(ctx, ErrNotRepository, "rev-parse", "--show-toplevel"); err != nil {
		return res, err
	}
	if _, err := s.git(ctx, "status", "rev-parse", "-q", "--verify", "MERGE_HEAD"); err == nil {
		return res, ErrMergeInProgress
	}
	if err := s.precondition(ctx, ErrUnbornBranch, "rev-parse", "-q", "--verify", "HEAD"); err != nil {
		return res, err
	}
	if err := s.precondition(ctx, ErrDetachedHead, "symbolic-ref", "-q", "HEAD"); err != nil {
		return res, err
	}
	if err := s.precondition(ctx, ErrNoUpstream, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err != nil {
		return res, err
	}

	committed, err := s.commitSnippets(ctx)
	if err != nil {
		return res, err
	}
	res.Committed = committed

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

func (s *Syncer) commitSnippets(ctx context.Context) (bool, error) {
	// git add fails when a wildcard pathspec matches nothing.
	snippets, err := s.git(ctx, "status", append([]string{"ls-files", "--cached", "--others", "--exclude-standard", "--"}, snippetPathspecs...)...)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(snippets) != "" {
		if _, err := s.git(ctx, "add", append([]string{"add", "-A", "--"}, snippetPathspecs...)...); err != nil {
			return false, err
		}
	}
	staged, err := s.git(ctx, "status", append([]string{"diff", "--cached", "--name-only", "--"}, snippetPathspecs...)...)
	if err != nil || strings.TrimSpace(staged) == "" {
		return false, err
	}
	if _, err := s.git(ctx, "commit", append([]string{"commit", "-m", s.commitMessage(), "--"}, snippetPathspecs...)...); err != nil {
		return false, err
	}
	return true, nil
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
	if _, err := s.git(cleanupCtx, "status", "rev-parse", "-q", "--verify", "MERGE_HEAD"); err != nil {
		return pullErr
	}
	var files []string
	if out, err := s.git(cleanupCtx, "status", "-c", "core.quotePath=false", "diff", "--name-only", "--diff-filter=U"); err == nil {
		for _, line := range strings.Split(out, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				files = append(files, line)
			}
		}
	}
	if _, err := s.git(cleanupCtx, "merge", "merge", "--abort"); err != nil {
		return err
	}
	if len(files) == 0 {
		return pullErr
	}
	return &ConflictError{Files: files}
}

func (s *Syncer) commitMessage() string {
	return fmt.Sprintf("snipgo sync: %s (%s)", s.opts.Now().Format("2006-01-02 15:04"), s.opts.Hostname)
}
