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
		wantErrHas string
	}{
		{name: "up to date", s: fakeSyncer{}, wantOut: "Already up to date\n"},
		{name: "pushed", s: fakeSyncer{res: gitsync.Result{Committed: true, Pushed: true}}, wantOut: "Synced: committed local changes, pushed\n"},
		{name: "conflict", s: fakeSyncer{err: &gitsync.ConflictError{Files: []string{"a.md"}}}, wantErrHas: "a.md"},
		{name: "no upstream", s: fakeSyncer{err: gitsync.ErrNoUpstream}, wantErrHas: "git push -u"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := syncAndReport(context.Background(), tt.s, &out)
			if tt.wantErrHas != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErrHas)
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
