# Git Sync Implementation Summary

## What Was Built
- `snipgo sync` and an optional automatic sync in the GUI keep the snippet directory in sync with a git remote. One run is commit (snippet directory only), pull (`--no-rebase`, `-X ours` or `-X theirs` for `local` or `remote`), then push when ahead of upstream.
- The user sets up the repository and upstream once. SnipGo never runs `git init` or touches remotes.

## Config (Task 1)
- `sync.enabled` (default false), `sync.interval_minutes` (default 10, minimum 1), `sync.conflict` (`manual`, `local`, `remote`; default `manual`).
- `config.SetValue` backs `snipgo config set` for all keys. `ValidateSync` rejects bad input on save. `NormalizeSync` replaces bad values in an existing file with defaults and logs a warning.

## Engine (Tasks 2 and 3)
- New package `internal/gitsync` with no Wails dependency. `Runner` interface with `ExecRunner`, which sets `GIT_TERMINAL_PROMPT=0` and `cmd.WaitDelay` so a child holding the pipes cannot outlive the deadline.
- `Syncer.Sync` checks git is installed, the directory is in a repository, no merge is in progress, and the branch has an upstream. Each failure has a typed error (`ErrGitNotFound`, `ErrNotRepository`, `ErrMergeInProgress`, `ErrNoUpstream`, `ErrLocked`, `GitError`, `ConflictError`).
- On a failed pull, unmerged files are listed (`core.quotePath=false`, scoped to the snippet directory) and `merge --abort` runs on a fresh 10s context, so a timeout or cancel mid-pull never leaves a merge in progress. Every mode aborts, because `-X` cannot resolve modify/delete conflicts.
- `Describe` maps errors to a stable kind and an actionable message. `Summary` and `NewStatus` feed the CLI and GUI.
- Integration tests run against real repositories (bare remote plus two clones): commit/push/pull, deletion sync, manual conflict abort, local and remote strategies, modify/delete conflict, commit scoping to the snippet directory, repository errors, Korean file names.

## Scheduler (Task 4)
- `Scheduler` runs automatic syncs at start and every interval, and `RunNow` for manual runs. Runs never overlap. An automatic tick is skipped when a run is in progress or when paused.
- A `ConflictError` pauses automatic sync. A successful manual run clears the pause. A canceled manual run emits no event and leaves the pause unchanged.

## CLI (Task 5)
- `snipgo sync` loads the config, runs one sync with a 2 minute timeout, prints the summary, and returns an error (exit code 1) with the described message on failure.

## GUI backend (Task 6)
- `app.App` starts the scheduler on startup when `sync.enabled` is true and restarts it on `SetDataDirectory` and `SaveSyncSettings`. `OnShutdown` stops it first.
- New bound methods: `SyncNow`, `GetSyncStatus`, `GetSyncSettings`, `SaveSyncSettings`. Status is emitted to the frontend as an event.
- `SyncNow` works regardless of `enabled`. It uses an App-owned cancelable context, an App-level run mutex across scheduler instances, and a generation counter so stale events from a replaced scheduler are dropped. Restart and shutdown cancel and wait for an in-flight manual run.

## Frontend (Tasks 7 and 8)
- Toasts: `useToasts`, `ToastViewport`, `useSyncNotifications`. Manual results always show. Automatic runs show errors only, and the same automatic error is shown once until a success resets it.
- Settings has a Git Sync section with the enabled toggle, interval (validated, minimum 1), conflict select ("Stop and let me resolve", "Prefer this machine", "Prefer remote"), a "Sync now" button, the last run, and a notice when automatic sync is paused. A failure to load sync settings does not disable the rest of the Settings page.

## Decisions During Implementation
- Parallel worktrees per task (`gs/tN`), cherry-picked onto the integration branch. Each implementer ran only its own package tests plus `go vet`, and the full suite ran after integration.
- Locked message reworded to "The repository is busy because another git process is using it. Try again in a moment." because the plan's code and test disagreed on case.
- Task 2 fix round: `WaitDelay` for the timeout guarantee, quotePath off for Korean file names, merge abort on a fresh context, `conflictFiles` never null.
- Task 6 fix round: manual runs are canceled and serialized across scheduler restarts. The conflict pause resets on restart and the stored status `Paused` flag is cleared with it, since saving settings is a deliberate user action and a persisting conflict re-pauses on the next run.
- Task 7: the toast mock is typed `Mock<(toast: Omit<Toast, "id">) => void>` because `ReturnType<typeof vi.fn>` failed `tsc`.
- Task 8 fix round: a rejected `SyncNow` is surfaced, and a sync load failure is isolated from the data directory setting.

## Tests
- `internal/gitsync`: 27 test functions (fake runner unit tests, scheduler, describe, real-repository integration, runner timeout). `internal/config`: 8. `cmd/snipgo`: 28 including `TestSyncAndReport`.
- Frontend: 78 tests passed in 9 files after Task 8.
- `app/` has no automated tests. The Wails wiring was verified by `make build-gui` only.

## Known Limitations
- Credentials are whatever the user's git setup provides. With no credential available, a run fails instead of prompting.
- Automatic sync runs only while the GUI is open. There is no background daemon, and the CLI never syncs on its own.
- After a `manual` conflict, the user resolves it with git outside SnipGo. Pausing resets when the app restarts or settings are saved, and a persisting conflict re-pauses on the next run.
- Deferred minors
  - `ValidateSync` is untested and duplicates `SetValue` checks. `sync.conflict` is not trimmed.
  - `context.Canceled` is described as kind `git`. Several error paths lack tests (Sync ctx pass-through, `ExecRunner`, `merge --abort` failure).
  - The runner test's fake `git` is POSIX-only.
  - cobra prints usage after `snipgo sync` errors (`SilenceUsage` not set).
  - A tick racing `Stop` can start one extra run. A just-finished old run can store one stale status. `SyncNow` with a context canceled by a concurrent restart silently does nothing.
  - Success toasts use `role="alert"` and dismiss buttons share one label.
  - Settings validation error is not cleared when the interval is fixed, and the Sync now button lacks `type="button"`.
