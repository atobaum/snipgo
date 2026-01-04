# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SnipGo is a local-first snippet manager with dual interfaces (CLI + GUI). All snippets are stored as Markdown files with YAML frontmatter in `~/.config/snipgo/snippets/`, allowing editing with any text editor.

## Essential Commands

### Development

```bash
# Install dependencies
make install-deps

# Development mode (GUI with hot reload)
make dev
# or: wails dev

# Build CLI only (standalone binary)
make build-cli

# Build GUI (requires wails CLI)
make build-gui

# Build both
make build
```

### Testing

```bash
# Go tests
go test ./...
go test -v -cover ./...

# Run single Go test
go test -v -run TestParseFrontmatter ./internal/core

# Frontend tests
cd frontend && pnpm test
cd frontend && pnpm test:watch
cd frontend && pnpm test:coverage
```

### Linting & Formatting

```bash
# Lint Go code (checks gofmt, go vet, goimports)
make lint-go

# Lint TypeScript
make lint-ts

# Lint both
make lint

# Auto-fix linting issues
make lint-fix

# Type-check TypeScript
make type-check
```

### CLI Usage

```bash
# Create snippet interactively
./bin/snipgo new

# List all snippets
./bin/snipgo list

# Search snippets (requires fzf)
./bin/snipgo search                              # List all snippets
./bin/snipgo search -q "docker"                  # Search with query
./bin/snipgo search --tag golang                 # Filter by tag
./bin/snipgo search --tag golang --tag web       # Multiple tags (AND logic)
./bin/snipgo search --lang bash                  # Filter by language
./bin/snipgo search --tag devops --lang bash -q "deploy"  # Combined filters and query

# Execute snippet as shell command (requires fzf)
./bin/snipgo exec

# Copy snippet to clipboard
./bin/snipgo copy "docker"

# Configuration
./bin/snipgo config show
./bin/snipgo config set data_directory /path/to/snippets
```

## Architecture

### High-Level Structure

```
Frontend (React/TypeScript)
    ↓ IPC (Wails)
App Struct (app/app.go) - Bridge Layer
    ↓
Manager (internal/core/manager.go) - Business Logic
    ↓
FileSystem (internal/storage/filesystem.go) - Disk I/O
```

### Core Components

#### internal/core/manager.go - Central Coordinator

The `Manager` struct is the heart of the application:
- Maintains in-memory cache of all snippets (`map[string]*Snippet`)
- Provides thread-safe operations via `sync.RWMutex`
- Coordinates between storage layer and application logic
- All snippets loaded into memory on startup for fast search/access

**Key Operations:**
- `LoadAll()` - Loads all `.md` files from disk into memory cache
- `Save()` - Persists to disk, updates memory, sets `updated_at` timestamp
- `GetByID()` - Returns copy (not pointer) to prevent mutations
- `Search()` - Fuzzy search with scoring (title > tags > body)

#### internal/core/frontmatter.go - Data Format

Snippets stored as Markdown files with YAML frontmatter:
- `ParseFrontmatter()` - Splits YAML metadata from body content
- `SerializeFrontmatter()` - Converts Snippet struct back to file format
- Format: `---\nYAML\n---\n\nBody`

#### internal/storage/filesystem.go - File Operations

File naming: `{sanitized-title}_{timestamp}.md`
- All files stored in configured `snippetsDir` (default: `~/.config/snipgo/snippets/`)
- No subdirectories - flat structure
- Filename sanitization removes special characters

#### app/app.go - Wails Bridge

Exposes Go methods to frontend via Wails IPC:
- Methods must be public (capitalized) to be callable from frontend
- `OnStartup(ctx)` captures Wails runtime context for clipboard operations
- Type conversion happens in frontend bridge layer (frontend/src/bridge.ts)

### Frontend Architecture

#### State Management (App.tsx)

- **No Redux/Context** - Local component state only
- `selectedSnippet` - Currently edited snippet
- `isDirtyRef` - useRef for tracking unsaved changes (non-state for performance)
- `listRefreshKey` - Forces SnippetList re-render after save/delete

#### Smart Save Strategy (SnippetEditor.tsx)

**Two save modes:**
1. **Immediate save** - Tags and favorites (`saveTagsAndFavorite()`)
   - Triggers on tag/favorite change
   - No confirmation needed
2. **Explicit save** - Title, body, language (Save button)
   - Shows "수정됨" (Modified) indicator when dirty
   - Confirmation dialog on navigation if unsaved

**Dirty tracking:**
- Only title, body, language tracked (not tags/favorites)
- Prevents accidental data loss

#### Type Bridge (frontend/src/bridge.ts)

- Converts between Go time.Time and JavaScript Date/ISO strings
- Re-exports Wails-generated bindings with proper TypeScript types
- **Important:** Always use bridge functions, not raw Wails bindings

### CLI vs GUI

**CLI** (`cmd/snipgo/main.go`):
- Uses Cobra for command framework
- Direct access to Manager (no Wails layer)
- Integrates with fzf for interactive selection
- Uses same core logic as GUI

**GUI** (`main.go`):
- Entry point calls `wails.Run()`
- Must be built with `wails build` (not `go build`)
- Asset embedding handled by Wails CLI

## Important Conventions

### Go Code Style (from .cursorrules)

- **File naming:** lowercase with underscores (`snippet_test.go`, `file_watcher.go`)
- **Package naming:** lowercase, single word
- **Public functions:** Start with capital letter, require doc comments
- **Error handling:** Explicit, early return pattern preferred
- **Testing:** Table-driven tests, `*_test.go` files in same package

### TypeScript Code Style (from .cursorrules)

- **Components:** Functional components only, PascalCase naming
- **File naming:** Match component name (`SnippetList.tsx`)
- **Props:** Define with `interface`
- **Variables:** `const` preferred, `let` only when necessary

### Commit Message Format (from .cursorrules)

```
<type>: <subject>

<body> (optional)
```

**Types:** feat, fix, refactor, docs, style, test, chore

**Examples:**
- `feat: add fuzzy search for snippet titles`
- `fix: resolve frontmatter parsing error with special characters`

### Testing Guidelines (from .cursorrules)

**Go:**
- Table-driven tests preferred
- Test coverage target: 70%+
- Test files: `*_test.go` (e.g., `snippet_test.go`)
- Mock external dependencies with interfaces

**TypeScript:**
- Vitest + React Testing Library
- Test files: `*.test.tsx` (e.g., `SnippetList.test.tsx`)
- User-centric testing (query by text, simulate events)
- Arrange-Act-Assert pattern

## Critical Build Notes

1. **GUI must use `wails build`**, not `go build`
   - Wails CLI handles build tags and asset embedding
   - See build error in git history for context

2. **Generate Wails bindings** after changing Go methods:
   ```bash
   wails generate module
   ```
   Updates TypeScript bindings in `frontend/wailsjs/go/`

3. **Frontend dist directory** must exist before GUI build:
   ```bash
   mkdir -p frontend/dist && touch frontend/dist/.gitkeep
   ```

## Data Flow Examples

### Creating a Snippet (GUI)
1. User edits in SnippetEditor → clicks Save
2. `app.SaveSnippet(snippet)` via Wails IPC
3. `Manager.Save()` validates, updates timestamp
4. `FileSystem.WriteFile()` serializes to markdown
5. Manager updates in-memory cache
6. Frontend reloads snippets, updates UI

### Search Flow
1. User types query → SnippetList calls `app.SearchSnippets(query)`
2. `Manager.Search()` uses fuzzy matching (sahilm/fuzzy)
3. Scoring: fuzzy title match > substring tag match (+10) > substring body match (+5)
4. Results sorted by score descending
5. Frontend displays filtered list

### External File Edit Sync
1. User edits `.md` file in VS Code/Obsidian
2. GUI calls `app.ReloadSnippets()`
3. Manager clears cache, rescans filesystem
4. Parses all `.md` files again
5. UI reflects latest state

## Configuration

**Config file path priority:**
1. Environment variable: `SNIPGO_CONFIG_PATH`
2. Default: `~/.config/snipgo/config.yaml`

**Data directory priority:**
1. Config file: `data_directory` field in config.yaml
2. Default: `~/.config/snipgo/snippets/`

**Config management:**
- `internal/config/config.go` handles all config logic
- `expandPath()` resolves `~` and environment variables
- CLI: `snipgo config set data_directory /path`

## Common Pitfalls

1. **Never commit `.snipgo/` directory** (user data)
2. **Always use RWMutex** when accessing Manager.snippets map
3. **Return copies, not pointers** from Manager methods (prevent mutations)
4. **Call `ReloadSnippets()`** after external saves to sync state
5. **Use bridge.ts functions** for type conversion, not raw Wails bindings
6. **Don't use `--no-verify` or `--no-gpg-sign`** in git commits unless explicitly requested
7. **Test files belong in same package** as tested code (not separate `_test` package)
