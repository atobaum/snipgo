# SnipGo

Local-First Snippet Manager built with Go, Wails v2, and React.

## Features

- **Local First**: All data stored in `~/.config/snipgo/snippets/` as Markdown files
- **File over App**: Edit snippets with any text editor (VS Code, Obsidian, etc.)
- **CLI + GUI**: Use both terminal and desktop interface
- **Fuzzy Search**: Fast in-memory search with fuzzy matching
- **CodeMirror Editor**: Syntax highlighting for various languages
- **Smart Save**: Unsaved changes indicator with confirmation dialog
- **Auto-save**: Tags and favorites are saved immediately

## Installation

### Homebrew (Recommended)

```bash
brew tap atobaum/tap
brew install snipgo
```

### Prerequisites

- Go 1.25+
- Node.js 24 LTS and pnpm
- Wails v2 CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Build

#### Using Make (Recommended)

```bash
# Install dependencies
make install-deps

# Build CLI only
make build-cli

# Build GUI (requires wails CLI)
make build-gui

# Build both
make build
```

#### Manual Build

```bash
# Install dependencies
go mod download
cd frontend && pnpm install && cd ..

# Build CLI
go build -o bin/snipgo ./cmd/snipgo

# Build GUI (requires wails CLI)
cd frontend && pnpm run build && cd ..
wails build
```

**Note**: GUI application must be built using `wails build`, not `go build`. The `wails` CLI handles build tags and asset embedding automatically.

## Configuration

SnipGo supports configuration through environment variables and config file:

1. **Config File Path** (environment variable):
   ```bash
   export SNIPGO_CONFIG_PATH="/path/to/config.yaml"
   ```
   Default: `~/.config/snipgo/config.yaml`

2. **Config File** (`~/.config/snipgo/config.yaml` or path set by `SNIPGO_CONFIG_PATH`):
   ```yaml
   data_directory: ~/my-snippets
   ```

3. **Default**: `~/.config/snipgo/snippets/`

### CLI Configuration Commands

```bash
# Show current configuration
snipgo config show

# Set data directory
snipgo config set data_directory /path/to/snippets
```

## Usage

### CLI

```bash
# Create a new snippet interactively
snipgo new

# List all snippets
snipgo list

# Search snippets interactively with fzf (no args = all snippets)
snipgo search
snipgo search "docker"

# Execute a snippet as shell command (interactive selection with fzf)
snipgo exec

# Edit a snippet (interactive selection with fzf, opens in $EDITOR)
snipgo edit

# Copy snippet body to clipboard
snipgo copy "docker"

# Show version information
snipgo version

# Generate zsh completion script
snipgo completion zsh
```

**Note**: `exec`, `search`, and `edit` commands require [fzf](https://github.com/junegunn/fzf) to be installed for interactive selection.

### Zsh Shortcut

You can set up a keyboard shortcut to quickly search and insert snippets. Add the following to your `~/.zshrc`:

```bash
function snipgo-select() {
  BUFFER=$(snipgo search "$LBUFFER" 2>/dev/null)
  CURSOR=$#BUFFER
  zle redisplay
}
zle -N snipgo-select
stty -ixon
bindkey '^s' snipgo-select
```

This binds `Ctrl+S` to search snippets. When you press `Ctrl+S`, it will:
1. Use the text before your cursor as a search query
2. Open fzf to select a snippet
3. Insert the snippet body at your cursor position

**Note**: The `stty -ixon` command disables flow control to allow `Ctrl+S` to be used as a shortcut. You can change `'^s'` to any key combination you prefer (e.g., `'^r'` for Ctrl+R).

### GUI

Run the GUI application:

```bash
wails dev  # Development mode
# or
./bin/snipgo  # Built application
```

**GUI Features:**
- Edit snippet title, body, language with CodeMirror syntax highlighting
- "수정됨" (Modified) indicator when there are unsaved changes
- Confirmation dialog when switching snippets with unsaved changes
- Tags and favorites are auto-saved immediately when changed
- Selected snippet is highlighted in the list

## Data Format

Snippets are stored as Markdown files with YAML frontmatter:

```yaml
---
id: "550e8400-e29b-41d4-a716-446655440000"
title: "Docker Compose Setup"
tags: ["docker", "devops"]
language: "yaml"
is_favorite: true
created_at: 2025-12-20T10:00:00Z
updated_at: 2025-12-25T14:30:00Z
---

version: '3'
services:
  web:
    image: nginx
```

## Project Structure

```
snipgo/
├── cmd/snipgo/       # CLI entry point
├── internal/
│   ├── core/        # Business logic (includes frontmatter parsing)
│   ├── config/       # Configuration management
│   └── storage/      # File system operations
├── app/              # Wails backend
├── frontend/         # React frontend
└── main.go           # Wails entry point
```

## Roadmap

### ✅ Phase 1: MVP (Completed)

**Core Features:**
- ✅ Go project structure with Clean Architecture
- ✅ Markdown I/O with YAML frontmatter parsing
- ✅ CLI commands: `new`, `list`, `search`, `copy`, `exec`, `edit`, `version`, `completion`, `config`
- ✅ GUI with Wails v2: list view, detail view/edit, clipboard copy
- ✅ Configuration management
- ✅ Fuzzy search with in-memory indexing
- ✅ Tag input UI with chips/badges style (GUI)

### 🚧 Phase 2: Usability (In Progress)

**Planned Features:**
- ⏳ **Hot Reload**: fsnotify-based file watcher for real-time GUI updates when files are modified externally (CLI → GUI sync)
- ⏳ **CLI Enhancements**: Add flags to `new` command (`-t "Title" --tags "go,api"`) to pre-fill frontmatter before opening editor
- ⏳ **GUI Improvements**: 
  - Filtering and sorting by `is_favorite`
  - Enhanced tag management and filtering

### 📋 Phase 3: Polish (Planned)

**Future Features:**
- 📋 **Export**: Code snippet image capture (Carbon-style)
- 📋 **Interactive TUI**: Enhanced CLI with bubbletea for interactive search/select (alternative to fzf)
- 📋 **Theme**: Light/Dark mode and editor theme customization
- 📋 **Cloud Sync**: Optional synchronization with GitHub Gist or Git repositories (maintaining local-first principle)

## Related Projects

- [pet](https://github.com/knqyf263/pet) - Simple command-line snippet manager (inspiration for `exec` and `search` commands)
- [MassCode](https://github.com/massCodeIO/massCode) - A free and open source code snippets manager for developers

## License

MIT

