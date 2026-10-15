# `new` Command Flags Implementation Summary

## What Changed
- `snipgo new` gained `-t/--title`, `--tags`, `-L/--language` (`--lang` alias, same as `search`) and `-c/--clipboard`. `-m` and `-e` are kept.
- Body source priority: `--clipboard`, then piped stdin, then the interactive prompt. Clipboard/stdin bodies keep their content except one trailing newline. Empty or whitespace-only bodies are rejected.
- `--title` with clipboard/stdin is fully non-interactive. Without `--title`, a piped-stdin run prompts on `/dev/tty` and fails with a hint to use `-t` when no TTY exists.
- `-e` now opens a full document (frontmatter from flags + body, prefilled from clipboard/stdin when given). ID and created_at are preserved from the generated draft.
- The editor process reads from `/dev/tty` when available so `-e` works with piped stdin.

## Conflict Rules
- `--multiline` + `--editor`
- `--clipboard` + piped stdin
- `--clipboard` + `--multiline`
- piped stdin + `--multiline`

All are checked before any prompt or clipboard read.

## Structure
- `cmd/snipgo/new.go`: `newOptions`, `newIO` (stdin, stdinPiped, readClipboard, openTTY, promptLine, editContent), `chooseBodySource`, `buildNewSnippet`, `readTitle`, `promptCommand`, `editNewSnippet`, `parseTagsFlag`, `trimSingleTrailingNewline`. `runNew` wires real dependencies and saves through `Manager.Save`.
- `cmd/snipgo/prompt.go`: `openTTY` extracted; `openTTYOrStdin` uses it.
- `cmd/snipgo/helpers.go`: editor stdin uses `openTTYOrStdin`.
- Piped stdin is detected via `os.ModeCharDevice`, no new dependency.

## Tests
- `TestParseTagsFlag`, `TestTrimSingleTrailingNewline`, `TestChooseBodySource`, `TestBuildNewSnippet` (16 cases), `TestBuildNewSnippet_EditorShowsFrontmatterAndBody`. All fakes; no real clipboard, terminal or editor.

## Docs
- `docs/FEATURES.md`, `docs/PRD.md` 4.2, `README.md` CLI usage and roadmap.

## Not Covered by Automated Tests
- Real `readline` prompts, `-m` terminal input, real `/dev/tty` title prompt and a real `$EDITOR` session with piped stdin.
