# Go Idioms Cleanup Implementation Plan

## Project Goal
- Re-apply the worthwhile parts of the stale `refactor/phase1-foundation` branch on top of current main, and bump Go to 1.27.

## Requirements
1. Bump `go.mod` to Go 1.27
2. Hoist the template variable regex to a package-level variable in `internal/tmpl`
3. Extract `tmpl.MergeWithMetadata()` and use it from both the CLI and the Wails app
4. Fix fzf selection so snippets with identical display lines resolve to the right snippet
5. Use `filepath.WalkDir` instead of `filepath.Walk` in storage
6. Add `core.ErrSnippetNotFound` sentinel error, wrapped with the snippet ID
7. Move `promptForVariables` and `openTTYOrStdin` into `cmd/snipgo/prompt.go`

## Design Decisions
- fzf input lines are prefixed with the snippet index and a tab, hidden via `--with-nth 2..`. The selected index is parsed back, so duplicates and title collisions can't pick the wrong snippet.
- Dropped from the old branch: `Storage` interface, `cliApp` struct, ULID `MustNew` rewrite, moving `LoadAll` into the constructor, cosmetic renames.

## Implementation Steps
### Step 1: Go 1.27 bump
### Step 2: `tmpl` regex hoisting and `MergeWithMetadata` (test first)
### Step 3: fzf input/selection helpers (test first)
### Step 4: `ErrSnippetNotFound` (test first)
### Step 5: `WalkDir`, `prompt.go` split

## Critical Files
- `go.mod`
- `internal/tmpl/parser.go`, `internal/tmpl/expander.go`, `internal/tmpl/merge.go`
- `app/app.go`, `cmd/snipgo/template_helpers.go`, `cmd/snipgo/prompt.go`, `cmd/snipgo/helpers.go`
- `internal/core/errors.go`, `internal/core/manager.go`
- `internal/storage/filesystem.go`

## Backward Compatibility
- No behavior change for users except fzf selection correctness. Error messages for missing snippets keep the snippet ID.
