package gitsync

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecRunner_ReturnsPromptlyWhenChildHoldsPipes(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nsleep 30 &\nsleep 30\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	old := waitDelay
	waitDelay = 100 * time.Millisecond
	defer func() { waitDelay = old }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, _, err := ExecRunner{}.Run(ctx, dir, "status")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("Run took %v after the deadline", elapsed)
	}
}
