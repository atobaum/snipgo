# Hot Reload Implementation Plan

## Project Goal
- When snippet files change outside the GUI (CLI, another editor, `git pull`), the GUI list and the open snippet update automatically without a restart.

## Requirements
1. Watch the data directory and all subdirectories with fsnotify, including subdirectories created later. Skip `.git` and other dot-directories.
2. React only to `.md` files (plus removal/rename of watched subdirectories, which can hide `.md` files).
3. Debounce events (~200ms) so bulk changes trigger one reload.
4. Ignore events caused by the app's own saves, deletes and renames.
5. On an external change, reload the Manager and emit the Wails event `snippets:changed`.
6. Frontend refreshes the list and the open snippet. Unsaved local edits are kept and a "Changed on disk" notice is shown. A snippet deleted on disk clears the selection.
7. Watcher starts in `OnStartup`, stops on shutdown, and is re-targeted by `SetDataDirectory`.

## Design Decisions
- New package `internal/watcher`, independent of Wails. `watcher.New(dir, debounce, onChange)` returns a `*Watcher` with `Close()` and `MarkSelfWrite(paths...)`.
- Self-write detection by state snapshot, not TTL. `MarkSelfWrite` records each path's current state (sha256 of contents, or "absent"). At flush time a changed path whose state still matches its snapshot is dropped. A mismatch means someone else touched it, so it is reported and the snapshot is forgotten. Checking at flush time (after the debounce) means the mark can safely happen right after the write returns.
- `Manager.FilePath(id)` is added so the app can mark the old and new paths around `Save` (rename) and `Delete`.
- Watcher callback captures the Manager it was started with, so swapping managers in `SetDataDirectory` does not race with the callback.
- Frontend: `onSnippetsChanged(cb)` in `bridge.ts` wraps `EventsOn`. A `useHotReload` hook in App handles the event. It looks up the open snippet via `GetAllSnippets` (no error-string parsing) and only replaces it when content fields differ, to avoid resetting the editor needlessly.
- Dirty handling: keep local edits and show a "Changed on disk" badge in the editor header until another snippet is selected or the snippet is refreshed from disk. Auto-save will then write local edits (local wins), which matches the existing auto-save model.

## Implementation Steps
### Step 1: `Manager.FilePath` (test first)
### Step 2: `internal/watcher` (tests first: debounce, subdir add, dot-dir skip, non-.md ignore, self-write ignore, close)
### Step 3: App wiring (`OnStartup`, `OnShutdown`, `SetDataDirectory`, self-write marking), `main.go` `OnShutdown`
### Step 4: Frontend bridge + hook + notice (tests first in `App.test.tsx`)
### Step 5: `wails generate module`, verification, docs

## Critical Files
- `internal/watcher/watcher.go`, `internal/watcher/watcher_test.go`
- `internal/core/manager.go`, `internal/core/manager_test.go`
- `app/app.go`, `main.go`
- `frontend/src/bridge.ts`, `frontend/src/hooks/useHotReload.ts`, `frontend/src/App.tsx`, `frontend/src/App.test.tsx`
- `frontend/src/components/SnippetEditor.tsx`, `frontend/src/components/EditorHeader.tsx`
- `go.mod`, `go.sum`, `docs/FEATURES.md`

## Backward Compatibility
- No change to file format or CLI. The GUI gains a background watcher. If the watcher fails to start, the GUI logs the error and works as before.
