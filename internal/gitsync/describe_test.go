package gitsync

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDescribe(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantKind     string
		wantContains string
	}{
		{"git missing", ErrGitNotFound, "git_not_found", "git is not installed"},
		{"not repo", ErrNotRepository, "not_repository", "git init"},
		{"no upstream", ErrNoUpstream, "no_upstream", "git push -u"},
		{"merge in progress", ErrMergeInProgress, "merge_in_progress", "merge is in progress"},
		{"locked", ErrLocked, "locked", "another git process"},
		{"conflict", &ConflictError{Files: []string{"a.md", "b.md"}}, "conflict", "a.md, b.md"},
		{"pull", &GitError{Op: "pull", Stderr: "fatal: unable to access\n"}, "pull", "git pull failed: fatal: unable to access"},
		{"push", &GitError{Op: "push", Stderr: "Permission denied"}, "push", "git push failed: Permission denied"},
		{"commit", &GitError{Op: "commit", Stderr: "Please tell me who you are."}, "git", "git commit failed"},
		{"timeout", &GitError{Op: "pull", Err: context.DeadlineExceeded}, "timeout", "timed out"},
		{"wrapped sentinel", fmt.Errorf("x: %w", ErrNoUpstream), "no_upstream", "git push -u"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, msg := Describe(tt.err)
			if kind != tt.wantKind || !strings.Contains(msg, tt.wantContains) {
				t.Errorf("Describe = (%q, %q), want kind %q containing %q", kind, msg, tt.wantKind, tt.wantContains)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	tests := []struct {
		r    Result
		want string
	}{
		{Result{}, "Already up to date"},
		{Result{Committed: true, Pushed: true}, "Synced: committed local changes, pushed"},
		{Result{Pulled: true}, "Synced: pulled remote changes"},
		{Result{Committed: true, Pulled: true, Pushed: true}, "Synced: committed local changes, pulled remote changes, pushed"},
	}
	for _, tt := range tests {
		if got := Summary(tt.r); got != tt.want {
			t.Errorf("Summary(%+v) = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestNewStatus(t *testing.T) {
	at := time.Date(2026, 10, 15, 21, 50, 0, 0, time.UTC)
	ok := NewStatus(Event{At: at, Result: Result{Pushed: true}, Manual: true})
	if !ok.OK || ok.Summary != "Synced: pushed" || ok.LastRunAt != "2026-10-15T21:50:00Z" || !ok.Manual {
		t.Errorf("success status = %+v", ok)
	}
	bad := NewStatus(Event{At: at, Err: &ConflictError{Files: []string{"a.md"}}, Paused: true})
	if bad.OK || bad.ErrorKind != "conflict" || len(bad.ConflictFiles) != 1 || !bad.Paused || bad.Message == "" {
		t.Errorf("conflict status = %+v", bad)
	}
}
