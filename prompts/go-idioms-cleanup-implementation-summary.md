# Go Idioms Cleanup Implementation Summary

## Changes
- `go.mod` bumped to Go 1.27
- `tmpl.MergeWithMetadata()` replaces the duplicated merge logic in `app/app.go` and `cmd/snipgo/template_helpers.go`. Tests moved to `internal/tmpl/merge_test.go`, plus a no-mutation test.
- Template variable regex compiled once as `tmpl.varPattern`
- fzf selection resolved by a hidden index prefix (`--delimiter '\t' --with-nth 2..`) via `buildFzfInput` / `snippetFromFzfSelection`. Fixes wrong picks when display lines or titles collide.
- `core.ErrSnippetNotFound` returned (wrapped with the ID) from `GetByID` and `Delete`
- `filepath.WalkDir` in `storage.ListFiles`
- `promptForVariables` and `openTTYOrStdin` moved to `cmd/snipgo/prompt.go`

## Not Included
- `Storage` interface, `cliApp` struct, ULID rewrite, `LoadAll` move into constructor, cosmetic renames from `refactor/phase1-foundation`

## Known Issues (pre-existing)
- `make lint-go` fails on gofmt for `internal/core/search.go` and `internal/core/manager_test.go`, both already unformatted on main.
