// Package watcher reports external changes to snippet files in a directory tree.
package watcher

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// fileState identifies a file's contents, or its absence, at a point in time.
type fileState struct {
	exists bool
	sum    [sha256.Size]byte
}

// Watcher watches a directory tree for .md file changes and reports them in debounced batches.
// Changes the app made itself can be excluded with MarkSelfWrite.
type Watcher struct {
	fsw      *fsnotify.Watcher
	debounce time.Duration
	onChange func(paths []string)

	// watchedDirs is only touched by the event loop goroutine after New returns.
	watchedDirs map[string]struct{}
	done        chan struct{}

	mu        sync.Mutex
	pending   map[string]struct{}
	timer     *time.Timer
	selfState map[string]fileState
	closed    bool

	flushMu   sync.Mutex
	closeOnce sync.Once
	closeErr  error
}

// New starts watching dir and all its non-hidden subdirectories.
// onChange is called from a background goroutine with the changed paths.
func New(dir string, debounce time.Duration, onChange func(paths []string)) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		fsw:         fsw,
		debounce:    debounce,
		onChange:    onChange,
		watchedDirs: make(map[string]struct{}),
		done:        make(chan struct{}),
		pending:     make(map[string]struct{}),
		selfState:   make(map[string]fileState),
	}

	if _, err := w.addTree(filepath.Clean(dir)); err != nil {
		fsw.Close()
		return nil, err
	}

	go w.run()
	return w, nil
}

// MarkSelfWrite records the current state of paths the app just wrote, renamed or deleted.
// Events for those paths are ignored while the file stays in the recorded state.
func (w *Watcher) MarkSelfWrite(paths ...string) {
	for _, p := range paths {
		p = filepath.Clean(p)
		state, err := readState(p)
		if err != nil {
			continue
		}
		w.mu.Lock()
		w.selfState[p] = state
		w.mu.Unlock()
	}
}

// Close stops watching. No onChange call starts after Close returns.
func (w *Watcher) Close() error {
	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		if w.timer != nil {
			w.timer.Stop()
		}
		w.mu.Unlock()

		w.closeErr = w.fsw.Close()
		<-w.done

		// Wait for an in-flight flush to finish.
		w.flushMu.Lock()
		defer w.flushMu.Unlock()
	})
	return w.closeErr
}

func (w *Watcher) run() {
	defer close(w.done)
	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			w.handle(ev)
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			slog.Warn("file watcher error", "error", err)
		}
	}
}

func (w *Watcher) handle(ev fsnotify.Event) {
	if ev.Op == fsnotify.Chmod {
		return
	}

	if ev.Has(fsnotify.Create) {
		if info, err := os.Lstat(ev.Name); err == nil && info.IsDir() {
			w.handleNewDir(ev.Name)
			return
		}
	}

	if ev.Has(fsnotify.Remove) || ev.Has(fsnotify.Rename) {
		if _, ok := w.watchedDirs[ev.Name]; ok {
			w.unwatchTree(ev.Name)
			w.markPending(ev.Name)
			return
		}
	}

	if isMarkdown(ev.Name) {
		w.markPending(ev.Name)
	}
}

func (w *Watcher) handleNewDir(dir string) {
	if isHidden(filepath.Base(dir)) {
		return
	}
	// Files may land in the new directory before its watch is added, so pick them up by walking it.
	files, err := w.addTree(dir)
	if err != nil {
		slog.Warn("failed to watch new directory", "path", dir, "error", err)
	}
	for _, f := range files {
		w.markPending(f)
	}
}

// addTree watches root and its non-hidden subdirectories and returns the .md files found.
func (w *Watcher) addTree(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return err
			}
			return nil
		}
		if !d.IsDir() {
			if isMarkdown(path) {
				files = append(files, path)
			}
			return nil
		}
		if path != root && isHidden(d.Name()) {
			return filepath.SkipDir
		}
		if err := w.fsw.Add(path); err != nil {
			return fmt.Errorf("failed to watch %s: %w", path, err)
		}
		w.watchedDirs[path] = struct{}{}
		return nil
	})
	return files, err
}

func (w *Watcher) unwatchTree(root string) {
	prefix := root + string(filepath.Separator)
	for dir := range w.watchedDirs {
		if dir == root || strings.HasPrefix(dir, prefix) {
			// The OS may already have dropped the watch, so the error is expected.
			_ = w.fsw.Remove(dir)
			delete(w.watchedDirs, dir)
		}
	}
}

func (w *Watcher) markPending(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	w.pending[path] = struct{}{}
	if w.timer == nil {
		w.timer = time.AfterFunc(w.debounce, w.flush)
	} else {
		w.timer.Reset(w.debounce)
	}
}

func (w *Watcher) flush() {
	w.flushMu.Lock()
	defer w.flushMu.Unlock()

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	pending := w.pending
	w.pending = make(map[string]struct{})
	w.timer = nil
	changed := w.dropSelfWrites(pending)
	w.mu.Unlock()

	if len(changed) == 0 {
		return
	}
	slices.Sort(changed)
	w.onChange(changed)
}

// dropSelfWrites must be called with w.mu held.
func (w *Watcher) dropSelfWrites(paths map[string]struct{}) []string {
	changed := make([]string, 0, len(paths))
	for p := range paths {
		if recorded, ok := w.selfState[p]; ok {
			current, err := readState(p)
			if err == nil && current == recorded {
				continue
			}
			delete(w.selfState, p)
		}
		changed = append(changed, p)
	}
	return changed
}

func readState(path string) (fileState, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return fileState{}, nil
	}
	if err != nil {
		return fileState{}, err
	}
	return fileState{exists: true, sum: sha256.Sum256(data)}, nil
}

func isMarkdown(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".md")
}

func isHidden(name string) bool {
	return len(name) > 1 && strings.HasPrefix(name, ".")
}
