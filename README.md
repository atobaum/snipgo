# SnipGo

Local-First Snippet Manager built with Go, Wails v2, and React.

## Features

- **Local First**: All data stored in `~/.snipgo/snippets/` as Markdown files
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

- Go 1.21+
- Node.js 18+ and pnpm
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

1. **Environment Variable** (highest priority):
   ```bash
   export SNIPGO_DATA_DIR="/path/to/your/snippets"
   ```

2. **Config File** (`~/.config/snipgo/config.yaml`):
   ```yaml
   data_directory: ~/my-snippets
   ```

3. **Default**: `~/.snipgo/snippets/`

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

# Copy snippet body to clipboard
snipgo copy "docker"
```

**Note**: `exec` and `search` commands require [fzf](https://github.com/junegunn/fzf) to be installed for interactive selection.

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

## Related Projects

- [pet](https://github.com/knqyf263/pet) - Simple command-line snippet manager (inspiration for `exec` and `search` commands)
- [MassCode](https://github.com/massCodeIO/massCode) - A free and open source code snippets manager for developers

## License

MIT

