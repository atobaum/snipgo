package gitsync

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrGitNotFound     = errors.New("git executable not found")
	ErrNotRepository   = errors.New("snippet directory is not inside a git repository")
	ErrNoUpstream      = errors.New("current branch has no upstream")
	ErrMergeInProgress = errors.New("a merge is in progress")
	ErrLocked          = errors.New("another git process is running in this repository")
)

// ConflictError means the merge has already been aborted.
type ConflictError struct {
	Files []string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("merge conflict in %s", strings.Join(e.Files, ", "))
}

type GitError struct {
	Op     string
	Stderr string
	Err    error
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git %s failed: %s", e.Op, strings.TrimSpace(e.Stderr))
}

func (e *GitError) Unwrap() error { return e.Err }
