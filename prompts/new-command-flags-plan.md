# `new` Command Flags Implementation Plan

## Project Goal
- Make it a one-liner to save a command drafted elsewhere as a snippet, e.g. `pbpaste | snipgo new -t "List ports" --tags net,mac -L bash` or `snipgo new -c -t "List ports"`.
- Implement the planned `new` flags that pre-fill frontmatter.

## Requirements
1. Flags: `-t/--title`, `--tags` (comma-separated), `-L/--language` with `--lang` alias (same as `search`), `-c/--clipboard`. Keep `-m/--multiline` and `-e/--editor`.
2. Body source priority: `--clipboard`, then piped stdin, then the existing interactive prompt (single-line, `-m`, or `-e`).
3. Conflicts are rejected before any prompt: `--clipboard` with piped stdin, `--clipboard` or piped stdin with `-m`, `-m` with `-e`.
4. Body is preserved exactly except one trailing newline is stripped. A body that is empty or whitespace-only is an error.
5. Title: `--title` skips the prompt. Otherwise prompt as today. When stdin is piped, the prompt reads from `/dev/tty`; if no TTY is available, fail with an error suggesting `-t`.
6. With `--title` plus clipboard/stdin the command is fully non-interactive.
7. Saving goes through `Manager.Save`.
8. Clipboard, stdin, TTY, line prompt and editor are injectable for tests.

## Design Decisions
- `-e` always opens a full document (frontmatter from flags + body). With clipboard/stdin the body is prefilled. This replaces the previous body-only editor so tags/language set from flags are visible and editable. ID and created_at are kept from the generated snippet.
- The editor's stdin uses `/dev/tty` (via the existing TTY helper) so `-e` works when stdin is piped.
- Tags are split on commas, trimmed, empty entries dropped, duplicates removed preserving first occurrence.
- `--language` wins over `--lang` when both are given, same as `search`.
- `openTTYOrStdin` is rebuilt on a new `openTTY` helper so the title prompt can fail instead of falling back to the piped stdin.
- Stdin is treated as piped when it is not a character device (`os.ModeCharDevice`), avoiding a new dependency.

## Implementation Steps
### Step 1: Tests first (`cmd/snipgo/new_test.go`)
- `parseTagsFlag`, `trimSingleTrailingNewline`, `chooseBodySource` (conflict table), `buildNewSnippet` (table with fake clipboard/stdin/TTY/editor).
### Step 2: Implement in `cmd/snipgo/new.go`
- `newOptions`, `newIO`, `chooseBodySource`, `readTitle`, `buildNewSnippet`, `runNew` wiring real dependencies.
### Step 3: `cmd/snipgo/prompt.go` and `helpers.go`
- Add `openTTY`; editor reads from TTY.
### Step 4: Docs
- `docs/FEATURES.md`, `docs/PRD.md` 4.2, `README.md` CLI usage.

## Critical Files
- `cmd/snipgo/new.go`
- `cmd/snipgo/new_test.go`
- `cmd/snipgo/prompt.go`
- `cmd/snipgo/helpers.go`
- `docs/FEATURES.md`, `docs/PRD.md`, `README.md`

## Backward Compatibility
- `snipgo new`, `-m` keep their behavior when stdin is a terminal.
- `-e` now shows frontmatter plus body instead of body only.
- Running `snipgo new` with non-terminal stdin now reads the body from stdin instead of prompting.
