# Git Sync Implementation Plan

## Project Goal
- Keep the snippet directory in sync across machines through a git remote, both periodically (GUI) and on demand (GUI button, `snipgo sync`), without SnipGo storing any credentials.

## Requirements
1. One sync run = stage → commit → pull → push, scoped to the snippet directory.
2. Manual sync from the GUI ("Sync now") and the CLI (`snipgo sync`). Manual sync works whenever the repository is set up, regardless of `sync.enabled`.
3. Automatic sync runs only in the GUI process: once at GUI startup, then every `sync.interval_minutes`. No background daemon.
4. Conflict handling is configurable via `sync.conflict`:
   - `manual` (default): abort the merge (`git merge --abort`), report the conflicted files, pause automatic sync until a manual sync succeeds.
   - `local`: merge with `-X ours`.
   - `remote`: merge with `-X theirs`.
   No conflict markers are ever left in snippet files by SnipGo.
5. The user prepares the repository (`git init`, `git remote add`, upstream branch, authentication). SnipGo only syncs and reports clearly when something is missing.
6. GUI shows toasts on errors (conflict, auth/network failure, missing repository or upstream, missing `git`) and on manual sync success. Successful automatic syncs are silent.
7. Settings live in `config.yaml` under `sync:` and are editable in the GUI Settings page and via `snipgo config set sync.<key> <value>`.
8. Files changed by pull are picked up by the existing hot reload.

## Design Decisions
- **Use the system `git` binary via `os/exec`.** It reuses the user's ssh keys, credential helpers and `.gitconfig`, and supports merge strategies (`-X ours/theirs`, `merge --abort`). `go-git` was rejected because its pull is fast-forward only and auth integration would have to be rebuilt.
- **Repository root may be the snippet directory or any parent.** All commands run with `git -C <snippet dir>`, and staging uses the pathspec `-- <snippet dir>` so only snippet files are committed.
- **Config shape**
  ```yaml
  sync:
    enabled: false
    interval_minutes: 10
    conflict: manual   # manual | local | remote
  ```
  Invalid values fall back to defaults with a warning. `interval_minutes` minimum is 1.
- **Commit message:** `snipgo sync: <YYYY-MM-DD HH:MM> (<hostname>)`.
- **Package layout**
  - `internal/gitsync`: `Runner` interface (`Run(ctx, dir, args...) (stdout, stderr string, err error)`) with an exec implementation; `Syncer` with `Sync(ctx) (Result, error)`; `Scheduler` with `Start`, `Stop`, `TriggerNow`, `Paused`.
  - Typed errors so callers can react without string parsing: `ErrGitNotFound`, `ErrNotRepository`, `ErrNoUpstream`, `ErrMergeInProgress`, `ErrLocked` (index.lock held by another git process), `*ConflictError{Files []string}`, `*RemoteError{Op string, Stderr string}` (pull/push failures such as auth or network).
  - `Result{Committed bool, Pulled bool, Pushed bool, At time.Time}`.
- **Sync steps in detail**
  1. Preconditions: `exec.LookPath("git")`; `git rev-parse --show-toplevel`; `git rev-parse --abbrev-ref --symbolic-full-name @{u}`; `MERGE_HEAD` absent (`git rev-parse -q --verify MERGE_HEAD`).
  2. `git add -A -- <dir>`; `git diff --cached --quiet -- <dir>` decides whether to commit.
  3. `git commit -m <message> -- <dir>` when there are staged changes.
  4. `git pull --no-rebase --no-edit [-X ours|-X theirs]`. On non-zero exit, check `git diff --name-only --diff-filter=U`. If files are conflicted and mode is `manual`, run `git merge --abort` and return `*ConflictError`. Other failures become `*RemoteError{Op:"pull"}`.
  5. `git rev-list --count @{u}..HEAD`; if > 0, `git push`. Failure becomes `*RemoteError{Op:"push"}`.
- **Scheduler:** runs a sync at start, then on a ticker. Runs never overlap (a mutex; a trigger during a run is coalesced). On `*ConflictError` in `manual` mode it pauses automatic runs; a successful manual run resumes them. Config changes restart the scheduler.
- **GUI wiring (`app/`)**
  - Wails methods: `SyncNow() (SyncStatus, error)`, `GetSyncStatus() SyncStatus`, `GetSyncSettings() SyncSettings`, `SaveSyncSettings(SyncSettings) error`.
  - `SyncStatus{LastRunAt, LastResult, LastError, ErrorKind, ConflictFiles, Paused, Running}`.
  - Every run emits `sync:status`. The frontend shows a toast for errors and for manual success.
  - The scheduler starts in `OnStartup` when enabled, stops in `OnShutdown`, restarts on settings or data directory changes.
- **Frontend**
  - New `Toast` component and a small toast queue hook (error toasts stay until dismissed, success toasts auto-dismiss after a few seconds).
  - Settings page gets a Sync section: enabled toggle, interval input, conflict select, "Sync now" button, last run time and result, paused notice with the conflicted files.
  - Bindings go through `bridge.ts`.
- **CLI:** `snipgo sync` runs `Syncer.Sync` once and prints a one-line summary (`committed, pulled, pushed` or `already up to date`). On conflict it lists the files and how to resolve them (`git -C <dir> pull`, fix, commit), exit code 1. Other typed errors map to short actionable messages.

## Implementation Steps
### Step 1: Config
- Add `Sync` settings to `internal/config` with defaults, validation, and `config set sync.*` support. Table-driven tests.

### Step 2: `internal/gitsync` Syncer
- Tests first with a fake `Runner` covering each step and every typed error.
- Integration tests with real git in temp dirs (bare remote + two clones, local identity configured per repo): clean sync, commit+push, pull from the other clone, `manual` conflict aborts and leaves the tree clean, `local` and `remote` resolve automatically, no upstream, not a repository, repo root above the snippet dir.

### Step 3: Scheduler
- Tests with a short interval and a fake Syncer: run at start, periodic runs, no overlap, pause on conflict, resume after manual success, stop.

### Step 4: CLI `snipgo sync`
- Command, output and exit codes, tests with a fake Syncer.

### Step 5: GUI backend
- Wails methods, `sync:status` event, scheduler lifecycle. Run `wails generate module`.

### Step 6: Frontend
- Toast component and hook, Settings Sync section, event subscription. Vitest + RTL tests first.

### Step 7: Docs and verification
- README (setup steps for the repository and remote, `snipgo sync`, settings), `docs/FEATURES.md` (Git Sync ✅), implementation summary.
- `go vet`, `go test ./...`, `vitest`, `tsc`, `make build-gui`.

## Critical Files
- `internal/config/config.go`, `cmd/snipgo/config.go`
- `internal/gitsync/` (new)
- `cmd/snipgo/sync.go` (new)
- `app/app.go`, `main.go`
- `frontend/src/bridge.ts`, `frontend/src/components/SettingsPage.tsx`, `frontend/src/components/Toast.tsx` (new), `frontend/src/App.tsx`
- `README.md`, `docs/FEATURES.md`

## Backward Compatibility
- Sync is off by default and existing configs without `sync:` load with defaults, so nothing changes for current users.
- `snipgo sync` is a new command. No existing command changes behavior.
- Snippet file format is unchanged.
