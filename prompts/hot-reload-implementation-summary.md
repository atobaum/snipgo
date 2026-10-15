# Hot Reload Implementation Summary

## What Was Built
- The GUI now picks up snippet file changes made outside the app (CLI, other editors, `git pull`) without a restart. The list refreshes, and the open snippet updates unless it has unsaved edits.

## Backend
- New package `internal/watcher` (no Wails dependency).
  - `New(dir, debounce, onChange)` watches `dir` and every non-hidden subdirectory, including ones created later. A newly created directory is walked so `.md` files that landed before its watch was added are still reported.
  - Only `.md` paths (case-insensitive) are reported. Chmod-only events are ignored. Removing or renaming a watched subdirectory is reported too, since it can hide `.md` files without per-file events.
  - Events are coalesced with a trailing debounce and delivered as one sorted batch.
  - `MarkSelfWrite(paths...)` records each path's state (sha256 of contents, or absent). At flush time, a path still in its recorded state is dropped. A mismatch reports the path and forgets the record. No TTL is involved.
  - `Close()` is idempotent, stops the event goroutine, and guarantees no `onChange` starts after it returns.
- `core.Manager.FilePath(id)` returns the file backing a snippet.
- `app.App`
  - `OnStartup` starts the watcher with a 200ms debounce. `OnShutdown` (wired in `main.go`) closes it. `SetDataDirectory` re-targets it.
  - On a batch, the watcher callback runs `LoadAll()` on the manager it was started with and emits `snippets:changed`.
  - `SaveSnippet`, `CreateSnippet` and `DeleteSnippet` mark the old and new paths as self-writes after the manager call.

## Frontend
- `bridge.ts` exports `onSnippetsChanged(callback)`, a wrapper around Wails `EventsOn` that returns an unsubscribe function.
- `hooks/useHotReload.ts` subscribes once, refreshes the list, then looks up the open snippet in `GetAllSnippets()`.
  - Missing means deleted, so the selection is cleared.
  - Same content fields (timestamps ignored) means nothing to do, which avoids resetting the editor.
  - Different content and not dirty replaces the open snippet.
  - Different content and dirty keeps local edits and sets `changedOnDisk`. The flag clears when another snippet is selected.
- `EditorHeader` shows a "Changed on disk" badge next to the save status.

## Tests
- `internal/watcher/watcher_test.go` covers create, modify, remove, uppercase extension, non-`.md` ignore, `.git` and new dot-directory skip, existing and newly created nested subdirectories, subdirectory removal, debounce of a 20-file burst into one batch, self write/delete/rename ignore, external write after self write, and `Close`.
- `internal/core/manager_test.go` adds `TestManager_FilePath`.
- `frontend/src/App.test.tsx` adds list refresh on event, clean snippet refresh, dirty snippet kept with notice, and deleted snippet clearing the selection.

## Known Limitations
- With auto-save (1.5s), a dirty snippet that changed on disk is soon saved with the local edits, so local wins. The badge stays as a hint until another snippet is selected.
- If the open snippet is deleted externally while it has unsaved edits, the edits are kept with a "Deleted on disk" badge and the next save recreates the file. The snippet is missing from the sidebar list until then.
- `app/` has no automated tests. The Wails wiring was verified by `make build-gui` only, not by running the GUI.
