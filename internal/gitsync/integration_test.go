package gitsync

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
}

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func newRemoteWithClones(t *testing.T) (a, b string) {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	gitCmd(t, root, "init", "--bare", "-b", "main", remote)
	a = filepath.Join(root, "a")
	b = filepath.Join(root, "b")
	gitCmd(t, root, "clone", remote, a)
	gitCmd(t, a, "config", "user.name", "A")
	gitCmd(t, a, "config", "user.email", "a@example.com")
	writeFile(t, filepath.Join(a, "x.md"), "base\n")
	gitCmd(t, a, "add", ".")
	gitCmd(t, a, "commit", "-m", "seed")
	gitCmd(t, a, "push", "-u", "origin", "main")
	gitCmd(t, root, "clone", remote, b)
	gitCmd(t, b, "config", "user.name", "B")
	gitCmd(t, b, "config", "user.email", "b@example.com")
	return a, b
}

func syncDir(t *testing.T, dir string, mode Mode) (Result, error) {
	t.Helper()
	return New(ExecRunner{}, Options{Dir: dir, Conflict: mode, Hostname: "test"}).Sync(context.Background())
}

func TestSync_CommitsPushesAndPulls(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	writeFile(t, filepath.Join(a, "new.md"), "hello\n")

	res, err := syncDir(t, a, ModeManual)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Committed || !res.Pushed {
		t.Errorf("A result = %+v, want committed and pushed", res)
	}

	res, err = syncDir(t, b, ModeManual)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Pulled || readFile(t, filepath.Join(b, "new.md")) != "hello\n" {
		t.Errorf("B result = %+v, want new.md pulled", res)
	}

	res, err = syncDir(t, b, ModeManual)
	if err != nil || res != (Result{}) {
		t.Errorf("second sync = %+v, %v, want up to date", res, err)
	}
}

func TestSync_DeletionIsSynced(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	if err := os.Remove(filepath.Join(a, "x.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := syncDir(t, a, ModeManual); err != nil {
		t.Fatal(err)
	}
	if _, err := syncDir(t, b, ModeManual); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(b, "x.md")); !os.IsNotExist(err) {
		t.Errorf("x.md still exists in B: %v", err)
	}
}

func makeContentConflict(t *testing.T, a, b string) {
	t.Helper()
	writeFile(t, filepath.Join(a, "x.md"), "from A\n")
	if _, err := syncDir(t, a, ModeManual); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(b, "x.md"), "from B\n")
}

func assertCleanAfterAbort(t *testing.T, dir string) {
	t.Helper()
	if out, err := exec.Command("git", "-C", dir, "rev-parse", "-q", "--verify", "MERGE_HEAD").CombinedOutput(); err == nil {
		t.Errorf("MERGE_HEAD still present: %s", out)
	}
	content := readFile(t, filepath.Join(dir, "x.md"))
	if strings.Contains(content, "<<<<<<<") {
		t.Errorf("conflict markers left in x.md: %q", content)
	}
}

func TestSync_ManualConflictAbortsCleanly(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	makeContentConflict(t, a, b)

	_, err := syncDir(t, b, ModeManual)
	var ce *ConflictError
	if !errors.As(err, &ce) || len(ce.Files) != 1 || ce.Files[0] != "x.md" {
		t.Fatalf("error = %v, want conflict in x.md", err)
	}
	assertCleanAfterAbort(t, b)
	if got := readFile(t, filepath.Join(b, "x.md")); got != "from B\n" {
		t.Errorf("x.md = %q, want local content kept", got)
	}
}

func TestSync_StrategiesResolveContentConflicts(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeLocal, "from B\n"},
		{ModeRemote, "from A\n"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			requireGit(t)
			a, b := newRemoteWithClones(t)
			makeContentConflict(t, a, b)

			res, err := syncDir(t, b, tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if got := readFile(t, filepath.Join(b, "x.md")); got != tt.want {
				t.Errorf("x.md = %q, want %q", got, tt.want)
			}
			if !res.Pushed {
				t.Error("merge result was not pushed")
			}
		})
	}
}

func TestSync_ModifyDeleteConflictWithStrategy(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	if err := os.Remove(filepath.Join(a, "x.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := syncDir(t, a, ModeManual); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(b, "x.md"), "edited in B\n")

	_, err := syncDir(t, b, ModeRemote)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want *ConflictError", err)
	}
	assertCleanAfterAbort(t, b)
}

func TestSync_ScopesCommitToSnippetDir(t *testing.T) {
	requireGit(t)
	a, _ := newRemoteWithClones(t)
	snippets := filepath.Join(a, "my snippets")
	writeFile(t, filepath.Join(snippets, "cmd.md"), "ls\n")
	writeFile(t, filepath.Join(a, "notes.txt"), "unrelated\n")
	gitCmd(t, a, "add", "notes.txt")

	res, err := syncDir(t, snippets, ModeManual)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Committed {
		t.Fatal("snippet change was not committed")
	}
	files := gitCmd(t, a, "show", "--name-only", "--format=", "HEAD")
	if strings.Contains(files, "notes.txt") || !strings.Contains(files, "my snippets/cmd.md") {
		t.Errorf("last commit files = %q, want only my snippets/cmd.md", files)
	}
	if status := gitCmd(t, a, "status", "--porcelain", "--", "notes.txt"); !strings.HasPrefix(status, "A ") {
		t.Errorf("notes.txt status = %q, want still staged", status)
	}
}

func TestSync_RepositoryErrors(t *testing.T) {
	requireGit(t)

	plain := t.TempDir()
	if _, err := syncDir(t, plain, ModeManual); !errors.Is(err, ErrNotRepository) {
		t.Errorf("plain dir error = %v, want ErrNotRepository", err)
	}

	local := t.TempDir()
	gitCmd(t, local, "init", "-b", "main")
	gitCmd(t, local, "config", "user.name", "L")
	gitCmd(t, local, "config", "user.email", "l@example.com")
	writeFile(t, filepath.Join(local, "x.md"), "x\n")
	gitCmd(t, local, "add", ".")
	gitCmd(t, local, "commit", "-m", "init")
	if _, err := syncDir(t, local, ModeManual); !errors.Is(err, ErrNoUpstream) {
		t.Errorf("no remote error = %v, want ErrNoUpstream", err)
	}
}

func TestSync_NonASCIIConflictFileName(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	const name = "도커 설정.md"
	writeFile(t, filepath.Join(a, name), "base\n")
	if _, err := syncDir(t, a, ModeManual); err != nil {
		t.Fatal(err)
	}
	if _, err := syncDir(t, b, ModeManual); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(a, name), "from A\n")
	if _, err := syncDir(t, a, ModeManual); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(b, name), "from B\n")

	_, err := syncDir(t, b, ModeManual)
	var ce *ConflictError
	if !errors.As(err, &ce) || len(ce.Files) != 1 || ce.Files[0] != name {
		t.Fatalf("error = %v, files = %+v, want conflict in %q", err, ce, name)
	}
	if out, err := exec.Command("git", "-C", b, "rev-parse", "-q", "--verify", "MERGE_HEAD").CombinedOutput(); err == nil {
		t.Errorf("MERGE_HEAD still present: %s", out)
	}
	if got := readFile(t, filepath.Join(b, name)); got != "from B\n" {
		t.Errorf("%s = %q, want local content kept", name, got)
	}
	if status := gitCmd(t, b, "status", "--porcelain", "--untracked-files=no"); strings.Contains(status, "U") {
		t.Errorf("unmerged entries remain: %q", status)
	}
}
