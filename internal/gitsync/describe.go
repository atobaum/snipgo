package gitsync

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Event is one finished sync run.
type Event struct {
	At     time.Time
	Result Result
	Err    error
	Manual bool
	Paused bool
}

// Status is the sync state shown in the GUI.
type Status struct {
	LastRunAt     string   `json:"lastRunAt"`
	OK            bool     `json:"ok"`
	Manual        bool     `json:"manual"`
	Summary       string   `json:"summary"`
	ErrorKind     string   `json:"errorKind"`
	Message       string   `json:"message"`
	ConflictFiles []string `json:"conflictFiles"`
	Paused        bool     `json:"paused"`
}

func NewStatus(ev Event) Status {
	st := Status{LastRunAt: ev.At.Format(time.RFC3339), Manual: ev.Manual, Paused: ev.Paused, ConflictFiles: []string{}}
	if ev.Err == nil {
		st.OK = true
		st.Summary = Summary(ev.Result)
		return st
	}
	st.ErrorKind, st.Message = Describe(ev.Err)
	var ce *ConflictError
	if errors.As(ev.Err, &ce) {
		if ce.Files != nil {
			st.ConflictFiles = ce.Files
		}
	}
	return st
}

// Describe returns a stable error kind for the UI and an actionable message.
func Describe(err error) (kind, message string) {
	var ce *ConflictError
	var ge *GitError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout", "git sync timed out. Check your network connection and git credentials."
	case errors.Is(err, ErrGitNotFound):
		return "git_not_found", "git is not installed or not in PATH."
	case errors.Is(err, ErrNotRepository):
		return "not_repository", withStderr("The snippet directory is not in a git repository. Run git init and add a remote first.", err)
	case errors.Is(err, ErrUnbornBranch):
		return "unborn_branch", "The current branch has no commits yet. Run git commit to make an initial commit first, then push it with git push -u origin HEAD."
	case errors.Is(err, ErrDetachedHead):
		return "detached_head", "HEAD is detached, so there is no branch to sync. Check out a branch with git switch <branch>, then sync again."
	case errors.Is(err, ErrNoUpstream):
		return "no_upstream", withStderr("The current branch has no upstream. Run: git push -u origin HEAD", err)
	case errors.Is(err, ErrMergeInProgress):
		return "merge_in_progress", "A git merge is in progress in the snippet repository. Resolve it with git, then sync again."
	case errors.Is(err, ErrLocked):
		return "locked", "The repository is busy because another git process is using it. Try again in a moment."
	case errors.As(err, &ce):
		return "conflict", fmt.Sprintf("Sync conflict in %d file(s): %s. Automatic sync is paused. Resolve with git (pull, fix, commit), then sync again.", len(ce.Files), strings.Join(ce.Files, ", "))
	case errors.As(err, &ge):
		kind = "git"
		if ge.Op == "pull" || ge.Op == "push" {
			kind = ge.Op
		}
		return kind, ge.Error()
	default:
		return "git", err.Error()
	}
}

func withStderr(message string, err error) string {
	var ge *GitError
	if errors.As(err, &ge) {
		if stderr := strings.TrimSpace(ge.Stderr); stderr != "" {
			return message + " (" + stderr + ")"
		}
	}
	return message
}

func Summary(r Result) string {
	var parts []string
	if r.Committed {
		parts = append(parts, "committed local changes")
	}
	if r.Pulled {
		parts = append(parts, "pulled remote changes")
	}
	if r.Pushed {
		parts = append(parts, "pushed")
	}
	if len(parts) == 0 {
		return "Already up to date"
	}
	return "Synced: " + strings.Join(parts, ", ")
}
