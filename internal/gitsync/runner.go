package gitsync

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"
)

// Bounds how long Run waits for pipes held open by git children after the context ends.
var waitDelay = 5 * time.Second

type Runner interface {
	Run(ctx context.Context, dir string, args ...string) (stdout, stderr string, err error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, dir string, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
	// The GUI has no terminal, so a credential prompt would hang forever.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.WaitDelay = waitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return stdout.String(), stderr.String(), ctxErr
	}
	return stdout.String(), stderr.String(), err
}
