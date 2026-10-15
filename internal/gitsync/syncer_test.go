package gitsync

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeResp struct {
	stdout, stderr string
	err            error
}

// fakeRunner answers by joined args; a slice is consumed in order and its last entry repeats.
type fakeRunner struct {
	responses map[string][]fakeResp
	calls     []string
	onCall    map[string]func()
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{responses: map[string][]fakeResp{
		"rev-parse -q --verify MERGE_HEAD": {{err: errors.New("exit status 1")}},
		"rev-parse HEAD":                   {{stdout: "aaa\n"}},
		"rev-list --count @{u}..HEAD":      {{stdout: "0\n"}},
		lsFilesKey:                         {{stdout: "a.md\n"}},
	}}
}

func (f *fakeRunner) set(key string, resps ...fakeResp) { f.responses[key] = resps }

func (f *fakeRunner) Run(ctx context.Context, _ string, args ...string) (string, string, error) {
	key := strings.Join(args, " ")
	f.calls = append(f.calls, key)
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	if hook := f.onCall[key]; hook != nil {
		hook()
	}
	resps := f.responses[key]
	if len(resps) == 0 {
		return "", "", nil
	}
	r := resps[0]
	if len(resps) > 1 {
		f.responses[key] = resps[1:]
	}
	return r.stdout, r.stderr, r.err
}

func (f *fakeRunner) called(key string) bool {
	for _, c := range f.calls {
		if c == key {
			return true
		}
	}
	return false
}

func newTestSyncer(r Runner, mode Mode) *Syncer {
	return New(r, Options{
		Dir:      "/snippets",
		Conflict: mode,
		Hostname: "host",
		Now:      func() time.Time { return time.Date(2026, 10, 15, 21, 50, 0, 0, time.Local) },
		LookPath: func(string) (string, error) { return "/usr/bin/git", nil },
	})
}

const conflictKey = "-c core.quotePath=false diff --name-only --diff-filter=U"

const mergeHeadKey = "rev-parse -q --verify MERGE_HEAD"

var (
	noMerge         = fakeResp{err: errors.New("exit status 1")}
	mergeInProgress = fakeResp{stdout: "abc\n"}
)

const specs = "*.[mM][dD] :(exclude,glob)**/.*/** :(exclude,glob).*/**"

const (
	lsFilesKey = "ls-files --cached --others --exclude-standard -- " + specs
	addKey     = "add -A -- " + specs
	stagedKey  = "diff --cached --name-only -- " + specs
	commitKey  = "commit -m snipgo sync: 2026-10-15 21:50 (host) -- " + specs
)

func TestSync_Preconditions(t *testing.T) {
	failing := fakeResp{err: errors.New("exit status 128"), stderr: "fatal"}
	tests := []struct {
		name    string
		setup   func(f *fakeRunner)
		noGit   bool
		wantErr error
	}{
		{name: "git missing", noGit: true, wantErr: ErrGitNotFound},
		{name: "not a repository", setup: func(f *fakeRunner) { f.set("rev-parse --show-toplevel", failing) }, wantErr: ErrNotRepository},
		{name: "merge in progress", setup: func(f *fakeRunner) { f.set("rev-parse -q --verify MERGE_HEAD", fakeResp{stdout: "abc\n"}) }, wantErr: ErrMergeInProgress},
		{name: "unborn branch", setup: func(f *fakeRunner) { f.set("rev-parse -q --verify HEAD", fakeResp{err: errors.New("exit status 1")}) }, wantErr: ErrUnbornBranch},
		{name: "detached head", setup: func(f *fakeRunner) { f.set("symbolic-ref -q HEAD", fakeResp{err: errors.New("exit status 1")}) }, wantErr: ErrDetachedHead},
		{name: "no upstream", setup: func(f *fakeRunner) { f.set("rev-parse --abbrev-ref --symbolic-full-name @{u}", failing) }, wantErr: ErrNoUpstream},
		{name: "index lock on add", setup: func(f *fakeRunner) {
			f.set(addKey, fakeResp{err: errors.New("exit status 128"), stderr: "fatal: Unable to create '/r/.git/index.lock': File exists."})
		}, wantErr: ErrLocked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeRunner()
			if tt.setup != nil {
				tt.setup(f)
			}
			s := newTestSyncer(f, ModeManual)
			if tt.noGit {
				s.opts.LookPath = func(string) (string, error) { return "", errors.New("not found") }
			}
			_, err := s.Sync(context.Background())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Sync error = %v, want %v", err, tt.wantErr)
			}
			if f.called("pull --no-rebase --no-edit") {
				t.Error("pull must not run when a precondition fails")
			}
			if tt.wantErr != ErrLocked && f.called(addKey) {
				t.Error("add must not run when a precondition fails")
			}
		})
	}
}

func TestSync_CommitOnlyWhenStaged(t *testing.T) {
	tests := []struct {
		name       string
		snippets   string
		staged     string
		wantAdd    bool
		wantCommit bool
	}{
		{name: "no snippet files", snippets: "", staged: "", wantAdd: false, wantCommit: false},
		{name: "nothing staged", snippets: "a.md\n", staged: "", wantAdd: true, wantCommit: false},
		{name: "staged changes", snippets: "a.md\n", staged: "a.md\n", wantAdd: true, wantCommit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeRunner()
			f.set(lsFilesKey, fakeResp{stdout: tt.snippets})
			f.set(stagedKey, fakeResp{stdout: tt.staged})
			res, err := newTestSyncer(f, ModeManual).Sync(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if f.called(addKey) != tt.wantAdd {
				t.Errorf("add called = %v, want %v (calls %v)", f.called(addKey), tt.wantAdd, f.calls)
			}
			if res.Committed != tt.wantCommit || f.called(commitKey) != tt.wantCommit {
				t.Errorf("Committed = %v, commit called = %v, want %v (calls %v)", res.Committed, f.called(commitKey), tt.wantCommit, f.calls)
			}
		})
	}
}

func TestSync_PullArgsByMode(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeManual, "pull --no-rebase --no-edit"},
		{ModeLocal, "pull --no-rebase --no-edit -X ours"},
		{ModeRemote, "pull --no-rebase --no-edit -X theirs"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			f := newFakeRunner()
			if _, err := newTestSyncer(f, tt.mode).Sync(context.Background()); err != nil {
				t.Fatal(err)
			}
			if !f.called(tt.want) {
				t.Errorf("expected %q in calls %v", tt.want, f.calls)
			}
		})
	}
}

func TestSync_PulledWhenHeadMoves(t *testing.T) {
	f := newFakeRunner()
	f.set("rev-parse HEAD", fakeResp{stdout: "aaa\n"}, fakeResp{stdout: "bbb\n"})
	res, err := newTestSyncer(f, ModeManual).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Pulled {
		t.Error("Pulled = false, want true when HEAD changed")
	}
}

func TestSync_PullFailures(t *testing.T) {
	pullFail := fakeResp{err: errors.New("exit status 1"), stderr: "CONFLICT (content): Merge conflict in a.md"}
	tests := []struct {
		name      string
		mode      Mode
		inMerge   bool
		conflicts string
		wantFiles []string
		wantAbort bool
		wantOp    string
	}{
		{name: "manual conflict", mode: ModeManual, inMerge: true, conflicts: "a.md\nb.md\n", wantFiles: []string{"a.md", "b.md"}, wantAbort: true},
		{name: "conflict left by -X strategy", mode: ModeRemote, inMerge: true, conflicts: "a.md\n", wantFiles: []string{"a.md"}, wantAbort: true},
		{name: "non-ASCII file name", mode: ModeManual, inMerge: true, conflicts: "한글.md\n", wantFiles: []string{"한글.md"}, wantAbort: true},
		{name: "conflict outside snippet dir", mode: ModeManual, inMerge: true, conflicts: ".zshrc\n", wantFiles: []string{".zshrc"}, wantAbort: true},
		{name: "merge in progress without unmerged files", mode: ModeManual, inMerge: true, conflicts: "", wantAbort: true, wantOp: "pull"},
		{name: "network failure", mode: ModeManual, inMerge: false, conflicts: "", wantAbort: false, wantOp: "pull"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeRunner()
			pullKey := "pull --no-rebase --no-edit"
			switch tt.mode {
			case ModeLocal:
				pullKey += " -X ours"
			case ModeRemote:
				pullKey += " -X theirs"
			}
			f.set(pullKey, pullFail)
			if tt.inMerge {
				f.set(mergeHeadKey, noMerge, mergeInProgress)
			}
			f.set(conflictKey, fakeResp{stdout: tt.conflicts})
			_, err := newTestSyncer(f, tt.mode).Sync(context.Background())

			if f.called("merge --abort") != tt.wantAbort {
				t.Errorf("merge --abort called = %v, want %v (calls %v)", f.called("merge --abort"), tt.wantAbort, f.calls)
			}
			if tt.wantFiles != nil {
				var ce *ConflictError
				if !errors.As(err, &ce) {
					t.Fatalf("error = %v, want *ConflictError", err)
				}
				if !reflect.DeepEqual(ce.Files, tt.wantFiles) {
					t.Errorf("Files = %v, want %v", ce.Files, tt.wantFiles)
				}
				return
			}
			var ge *GitError
			if !errors.As(err, &ge) || ge.Op != tt.wantOp {
				t.Fatalf("error = %v, want *GitError op %q", err, tt.wantOp)
			}
		})
	}
}

func TestSync_Push(t *testing.T) {
	tests := []struct {
		name       string
		ahead      string
		pushErr    error
		wantPush   bool
		wantPushed bool
		wantOp     string
	}{
		{name: "up to date", ahead: "0\n", wantPush: false},
		{name: "ahead pushes", ahead: "2\n", wantPush: true, wantPushed: true},
		{name: "push fails", ahead: "1\n", pushErr: errors.New("exit status 128"), wantPush: true, wantOp: "push"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeRunner()
			f.set("rev-list --count @{u}..HEAD", fakeResp{stdout: tt.ahead})
			f.set("push", fakeResp{err: tt.pushErr, stderr: "Permission denied (publickey)."})
			res, err := newTestSyncer(f, ModeManual).Sync(context.Background())
			if f.called("push") != tt.wantPush {
				t.Errorf("push called = %v, want %v", f.called("push"), tt.wantPush)
			}
			if tt.wantOp != "" {
				var ge *GitError
				if !errors.As(err, &ge) || ge.Op != tt.wantOp {
					t.Fatalf("error = %v, want *GitError op %q", err, tt.wantOp)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if res.Pushed != tt.wantPushed {
				t.Errorf("Pushed = %v, want %v", res.Pushed, tt.wantPushed)
			}
		})
	}
}

func TestSync_CommitFailureIsGitError(t *testing.T) {
	f := newFakeRunner()
	f.set(stagedKey, fakeResp{stdout: "a.md\n"})
	f.set(commitKey, fakeResp{err: errors.New("exit status 128"), stderr: "Please tell me who you are."})
	_, err := newTestSyncer(f, ModeManual).Sync(context.Background())
	var ge *GitError
	if !errors.As(err, &ge) || ge.Op != "commit" {
		t.Fatalf("error = %v, want *GitError op commit", err)
	}
}

func TestSync_ConflictAbortedWhenContextEndsMidPull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := newFakeRunner()
	f.onCall = map[string]func(){"pull --no-rebase --no-edit": cancel}
	f.set("pull --no-rebase --no-edit", fakeResp{err: errors.New("exit status 1"), stderr: "CONFLICT"})
	f.set(mergeHeadKey, noMerge, mergeInProgress)
	f.set(conflictKey, fakeResp{stdout: "a.md\n"})
	_, err := newTestSyncer(f, ModeManual).Sync(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if !f.called("merge --abort") {
		t.Error("merge --abort must still run after the context ended")
	}
}
