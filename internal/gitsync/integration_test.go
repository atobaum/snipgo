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
	if status := gitCmd(t, dir, "status", "--porcelain", "--untracked-files=no"); status != "" {
		t.Errorf("working tree not clean after abort: %q", status)
	}
	content := readFile(t, filepath.Join(dir, "x.md"))
	if strings.Contains(content, "<<<<<<<") {
		t.Errorf("conflict markers left in x.md: %q", content)
	}
}

func TestSync_ConflictOutsideSnippetDirIsAborted(t *testing.T) {
	requireGit(t)
	a, b := newRemoteWithClones(t)
	writeFile(t, filepath.Join(a, "snippets", "s.md"), "snippet\n")
	writeFile(t, filepath.Join(a, ".zshrc"), "base\n")
	gitCmd(t, a, "add", ".")
	gitCmd(t, a, "commit", "-m", "dotfiles")
	gitCmd(t, a, "push")
	gitCmd(t, b, "pull")

	writeFile(t, filepath.Join(a, ".zshrc"), "from A\n")
	gitCmd(t, a, "commit", "-am", "A zshrc")
	gitCmd(t, a, "push")
	writeFile(t, filepath.Join(b, ".zshrc"), "from B\n")
	gitCmd(t, b, "commit", "-am", "B zshrc")

	_, err := syncDir(t, filepath.Join(b, "snippets"), ModeManual)
	var ce *ConflictError
	if !errors.As(err, &ce) || len(ce.Files) != 1 || ce.Files[0] != ".zshrc" {
		t.Fatalf("error = %v, want conflict in .zshrc", err)
	}
	assertCleanAfterAbort(t, b)
	if got := readFile(t, filepath.Join(b, ".zshrc")); got != "from B\n" {
		t.Errorf(".zshrc = %q, want local content kept", got)
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

func TestSync_StagesOnlySnippetFiles(t *testing.T) {
	tests := []struct {
		name   string
		subdir string
	}{
		{name: "repo root", subdir: ""},
		{name: "inside hidden parent", subdir: filepath.Join(".config", "snipgo", "snippets")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireGit(t)
			a, _ := newRemoteWithClones(t)
			dir := filepath.Join(a, tt.subdir)
			writeFile(t, filepath.Join(dir, "a.md"), "a\n")
			writeFile(t, filepath.Join(dir, "sub", "b.md"), "b\n")
			writeFile(t, filepath.Join(dir, "C.MD"), "c\n")
			writeFile(t, filepath.Join(dir, ".DS_Store"), "junk")
			writeFile(t, filepath.Join(dir, "a.md.swp"), "swap")
			writeFile(t, filepath.Join(dir, ".obsidian", "workspace.json"), "{}")
			writeFile(t, filepath.Join(dir, ".obsidian", "hidden.md"), "h\n")
			nested := filepath.Join(dir, "nested")
			writeFile(t, filepath.Join(nested, "n.md"), "n\n")
			gitCmd(t, nested, "init", "-q")

			res, err := syncDir(t, dir, ModeManual)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Committed || !res.Pushed {
				t.Fatalf("result = %+v, want committed and pushed", res)
			}
			got := strings.Fields(gitCmd(t, a, "show", "--name-only", "--format=", "HEAD"))
			rel := func(p string) string { return filepath.ToSlash(filepath.Join(tt.subdir, p)) }
			want := []string{rel("C.MD"), rel("a.md"), rel("sub/b.md")}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("committed files = %v, want %v", got, want)
			}

			if err := os.Remove(filepath.Join(dir, "a.md")); err != nil {
				t.Fatal(err)
			}
			res, err = syncDir(t, dir, ModeManual)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Committed || !res.Pushed {
				t.Fatalf("deletion result = %+v, want committed and pushed", res)
			}
			remoteFiles := gitCmd(t, a, "ls-tree", "-r", "--name-only", "origin/main")
			if strings.Contains(remoteFiles, rel("a.md")) || !strings.Contains(remoteFiles, rel("sub/b.md")) {
				t.Errorf("remote files = %q, want a.md deleted and sub/b.md kept", remoteFiles)
			}
		})
	}
}

func TestSync_NoSnippetFiles(t *testing.T) {
	requireGit(t)
	a, _ := newRemoteWithClones(t)
	dir := filepath.Join(a, "empty")
	writeFile(t, filepath.Join(dir, ".DS_Store"), "junk")

	res, err := syncDir(t, dir, ModeManual)
	if err != nil || res != (Result{}) {
		t.Errorf("sync = %+v, %v, want up to date", res, err)
	}
}
