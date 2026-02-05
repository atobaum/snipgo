# Multiline Command Support for `snipgo new`

## Goal
Add multiline command input support to `snipgo new` command, following pet's approach with two options:
1. `--multiline` (`-m`) flag: Terminal-based multiline input
2. `--editor` (`-e`) flag: Open external editor ($EDITOR)

## Design Decisions

### Approach: Follow pet's dual-mode pattern
- **Multiline mode**: State machine with continuation prompt, two blank lines to finish
- **Editor mode**: Create temp file with template, open $EDITOR, parse result

### Why both options?
- `--multiline`: Quick inline input for short multi-line scripts
- `--editor`: Full editing experience for complex snippets with metadata

## Implementation Steps

### Step 1: Add flags to new command
**File:** `cmd/snipgo/new.go`

```go
var (
    useMultiLine bool
    useEditor    bool
)

func init() {
    newCmd.Flags().BoolVarP(&useMultiLine, "multiline", "m", false, "Enable multiline input (two blank lines to finish)")
    newCmd.Flags().BoolVarP(&useEditor, "editor", "e", false, "Open $EDITOR to create snippet")
}
```

### Step 2: Implement multiline input function
**File:** `cmd/snipgo/new.go`

```go
func scanMultiLine(firstPrompt, continuationPrompt string) (string, error) {
    // State machine: start -> lastLineNotEmpty -> lastLineEmpty -> done
    // Two consecutive blank lines = done
}
```

### Step 3: Implement editor mode for new snippets
**File:** `cmd/snipgo/new.go` (reuse patterns from edit.go)

```go
func createSnippetWithEditor(title string) (*core.Snippet, error) {
    // 1. Create new snippet with title
    // 2. Serialize to temp file
    // 3. Open $EDITOR
    // 4. Parse result
    // 5. Return snippet
}
```

### Step 4: Update runNew to use flags
**File:** `cmd/snipgo/new.go`

```go
func runNew(cmd *cobra.Command, args []string) error {
    // Get description first (always required)
    description := ...

    if useEditor {
        // Open editor with template
        snippet, err := createSnippetWithEditor(description)
        ...
    } else if useMultiLine {
        // Multiline terminal input
        command, err := scanMultiLine("Command> ", ".......> ")
        ...
    } else {
        // Current single-line behavior (unchanged)
        command, err := readline.Line("Command> ")
        ...
    }
}
```

## Critical Files to Modify

| File | Changes |
|------|---------|
| `cmd/snipgo/new.go` | Add flags, multiline scanner, editor mode |
| `cmd/snipgo/helpers.go` | (optional) Extract shared editor logic if needed |

## Reusable Code from edit.go

- Editor detection: `os.Getenv("EDITOR")` with "vi" fallback
- Temp file creation: `os.CreateTemp("", "snipgo-*.md")`
- Editor execution: `exec.Command(editor, tmpPath)` with stdin/stdout/stderr
- File modification check: Compare `ModTime()` before/after
- Parsing: `parseSnippetFromEdit()` in helpers.go

## User Experience

```bash
# Single line (current, unchanged)
$ snipgo new
Description> My snippet
Command> echo hello

# Multiline mode
$ snipgo new -m
Description> Docker cleanup script
Command> docker system prune -af
.......> docker volume prune -f
.......> docker network prune -f
.......>
.......>
Snippet saved: Docker cleanup script

# Editor mode
$ snipgo new -e
Description> Complex deployment script
# Opens vim/nano with template, user edits and saves
Snippet saved: Complex deployment script
```

## Editor Template Format
When `-e` flag is used, create temp file with:
```markdown
---
title: {user-provided description}
description: ""
tags: []
language: ""
is_favorite: false
---

# Enter your command/snippet below

```

## Verification
1. `snipgo new` - Single line still works (backward compatible)
2. `snipgo new -m` - Multiline input with two blank lines to finish
3. `snipgo new -e` - Opens $EDITOR, saves on close
4. `snipgo new -m -e` - Should error (mutually exclusive)
5. Created snippets appear in `snipgo list` and GUI

## Backward Compatibility
- Default behavior unchanged (single-line input)
- New flags are opt-in
- No changes to snippet file format

## References
- [pet snippet manager](https://github.com/knqyf263/pet) - Reference implementation for multiline support
