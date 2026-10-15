package watcher

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

const (
	testDebounce = 50 * time.Millisecond
	waitForEvent = 2 * time.Second
	quietPeriod  = 300 * time.Millisecond
)

func startWatcher(t *testing.T, dir string) (*Watcher, <-chan []string) {
	t.Helper()
	batches := make(chan []string, 16)
	w, err := New(dir, testDebounce, func(paths []string) { batches <- paths })
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() { w.Close() })
	return w, batches
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
}

func expectBatch(t *testing.T, batches <-chan []string) []string {
	t.Helper()
	select {
	case b := <-batches:
		return b
	case <-time.After(waitForEvent):
		t.Fatal("expected a change notification, got none")
		return nil
	}
}

func expectNoBatch(t *testing.T, batches <-chan []string) {
	t.Helper()
	select {
	case b := <-batches:
		t.Fatalf("expected no change notification, got %v", b)
	case <-time.After(quietPeriod):
	}
}

func TestWatcher_ReportsExternalChanges(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string)
		act        func(t *testing.T, dir string)
		wantChange bool
	}{
		{
			name: "create md file",
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "new.md"), "x")
			},
			wantChange: true,
		},
		{
			name: "modify md file",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.md"), "old")
			},
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.md"), "new")
			},
			wantChange: true,
		},
		{
			name: "remove md file",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.md"), "x")
			},
			act: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, "a.md")); err != nil {
					t.Fatal(err)
				}
			},
			wantChange: true,
		},
		{
			name: "uppercase md extension",
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "README.MD"), "x")
			},
			wantChange: true,
		},
		{
			name: "non-md file is ignored",
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "notes.txt"), "x")
				writeFile(t, filepath.Join(dir, "a.md.swp"), "x")
			},
			wantChange: false,
		},
		{
			name: "md file inside existing .git is ignored",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, ".git", "HEAD"), "ref")
			},
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, ".git", "x.md"), "x")
			},
			wantChange: false,
		},
		{
			name: "md file inside newly created dot-directory is ignored",
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, ".cache", "x.md"), "x")
			},
			wantChange: false,
		},
		{
			name: "md file in existing subdirectory",
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "sub", "deep", "x.md"), "x")
			},
			wantChange: true,
		},
		{
			name: "md file created together with new nested subdirectories",
			act: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a", "b", "x.md"), "x")
			},
			wantChange: true,
		},
		{
			name: "subdirectory with md files removed",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "sub", "x.md"), "x")
			},
			act: func(t *testing.T, dir string) {
				if err := os.RemoveAll(filepath.Join(dir, "sub")); err != nil {
					t.Fatal(err)
				}
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, dir)
			}
			_, batches := startWatcher(t, dir)

			tt.act(t, dir)

			if tt.wantChange {
				expectBatch(t, batches)
			} else {
				expectNoBatch(t, batches)
			}
		})
	}
}

func TestWatcher_WatchesSubdirectoryCreatedLater(t *testing.T) {
	dir := t.TempDir()
	_, batches := startWatcher(t, dir)

	if err := os.Mkdir(filepath.Join(dir, "later"), 0755); err != nil {
		t.Fatal(err)
	}
	expectNoBatch(t, batches)

	target := filepath.Join(dir, "later", "x.md")
	writeFile(t, target, "x")

	got := expectBatch(t, batches)
	if !slices.Contains(got, target) {
		t.Errorf("batch = %v, want it to contain %v", got, target)
	}
}

func TestWatcher_DebouncesBurstIntoOneNotification(t *testing.T) {
	dir := t.TempDir()
	_, batches := startWatcher(t, dir)

	const files = 20
	for i := range files {
		writeFile(t, filepath.Join(dir, "f"+string(rune('a'+i))+".md"), "x")
	}

	got := expectBatch(t, batches)
	if len(got) != files {
		t.Errorf("batch has %d paths, want %d: %v", len(got), files, got)
	}
	expectNoBatch(t, batches)
}

func TestWatcher_IgnoresSelfWrites(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string)
		act        func(t *testing.T, w *Watcher, dir string)
		wantChange bool
	}{
		{
			name: "self write",
			act: func(t *testing.T, w *Watcher, dir string) {
				p := filepath.Join(dir, "a.md")
				writeFile(t, p, "mine")
				w.MarkSelfWrite(p)
			},
			wantChange: false,
		},
		{
			name: "self delete",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.md"), "x")
			},
			act: func(t *testing.T, w *Watcher, dir string) {
				p := filepath.Join(dir, "a.md")
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				w.MarkSelfWrite(p)
			},
			wantChange: false,
		},
		{
			name: "self rename then write",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "old.md"), "x")
			},
			act: func(t *testing.T, w *Watcher, dir string) {
				oldPath, newPath := filepath.Join(dir, "old.md"), filepath.Join(dir, "new.md")
				if err := os.Rename(oldPath, newPath); err != nil {
					t.Fatal(err)
				}
				writeFile(t, newPath, "renamed")
				w.MarkSelfWrite(oldPath, newPath)
			},
			wantChange: false,
		},
		{
			name: "self recreate after external delete",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "a.md"), "x")
			},
			act: func(t *testing.T, w *Watcher, dir string) {
				p := filepath.Join(dir, "a.md")
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
				writeFile(t, p, "mine")
				w.MarkSelfWrite(p)
			},
			wantChange: false,
		},
		{
			name: "external write after self write",
			act: func(t *testing.T, w *Watcher, dir string) {
				p := filepath.Join(dir, "a.md")
				writeFile(t, p, "mine")
				w.MarkSelfWrite(p)
				writeFile(t, p, "theirs")
			},
			wantChange: true,
		},
		{
			name: "self write alongside external change to another file",
			act: func(t *testing.T, w *Watcher, dir string) {
				mine := filepath.Join(dir, "mine.md")
				writeFile(t, mine, "mine")
				w.MarkSelfWrite(mine)
				writeFile(t, filepath.Join(dir, "theirs.md"), "theirs")
			},
			wantChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, dir)
			}
			w, batches := startWatcher(t, dir)

			tt.act(t, w, dir)

			if tt.wantChange {
				expectBatch(t, batches)
			} else {
				expectNoBatch(t, batches)
			}
		})
	}
}

func TestWatcher_ExternalChangeAfterSelfWriteIsReportedLater(t *testing.T) {
	dir := t.TempDir()
	w, batches := startWatcher(t, dir)
	p := filepath.Join(dir, "a.md")

	writeFile(t, p, "mine")
	w.MarkSelfWrite(p)
	expectNoBatch(t, batches)

	writeFile(t, p, "theirs")
	got := expectBatch(t, batches)
	if !slices.Contains(got, p) {
		t.Errorf("batch = %v, want it to contain %v", got, p)
	}
}

func TestWatcher_Close(t *testing.T) {
	dir := t.TempDir()
	w, batches := startWatcher(t, dir)

	writeFile(t, filepath.Join(dir, "a.md"), "x")
	if err := w.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second Close() error: %v", err)
	}

	writeFile(t, filepath.Join(dir, "b.md"), "x")
	expectNoBatch(t, batches)
}

func TestNew_MissingDirectory(t *testing.T) {
	_, err := New(filepath.Join(t.TempDir(), "missing"), testDebounce, func([]string) {})
	if err == nil {
		t.Fatal("New() error = nil, want error for missing directory")
	}
}
