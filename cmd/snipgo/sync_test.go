package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"snipgo/internal/gitsync"
)

type fakeSyncer struct {
	res gitsync.Result
	err error
}

func (f fakeSyncer) Sync(context.Context) (gitsync.Result, error) { return f.res, f.err }

func TestSyncAndReport(t *testing.T) {
	tests := []struct {
		name       string
		s          fakeSyncer
		wantOut    string
		wantErrHas []string
		wantErrNot string
	}{
		{name: "up to date", s: fakeSyncer{}, wantOut: "Already up to date\n"},
		{name: "pushed", s: fakeSyncer{res: gitsync.Result{Committed: true, Pushed: true}}, wantOut: "Synced: committed local changes, pushed\n"},
		{
			name:       "conflict",
			s:          fakeSyncer{err: &gitsync.ConflictError{Files: []string{"a.md", "도커.md"}}},
			wantErrHas: []string{"a.md, 도커.md", `git -C "/my snippets" pull`, "commit", "snipgo sync"},
			wantErrNot: "Automatic sync is paused",
		},
		{name: "no upstream", s: fakeSyncer{err: gitsync.ErrNoUpstream}, wantErrHas: []string{"git push -u"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := syncAndReport(context.Background(), tt.s, "/my snippets", &out)
			if tt.wantErrHas != nil {
				if err == nil {
					t.Fatal("err = nil, want an error")
				}
				for _, want := range tt.wantErrHas {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("err = %q, want containing %q", err, want)
					}
				}
				if tt.wantErrNot != "" && strings.Contains(err.Error(), tt.wantErrNot) {
					t.Errorf("err = %q, must not contain %q", err, tt.wantErrNot)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.wantOut {
				t.Errorf("out = %q, want %q", out.String(), tt.wantOut)
			}
		})
	}
}

func TestSyncCmd_SilencesUsage(t *testing.T) {
	if !syncCmd.SilenceUsage {
		t.Error("SilenceUsage = false, want errors without the usage dump")
	}
}
