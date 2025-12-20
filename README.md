# SnipGo

Local-First Snippet Manager built with Go, Wails v2, and React.

## Features

- **Local First**: All data stored in `~/.snipgo/snippets/` as Markdown files
- **File over App**: Edit snippets with any text editor (VS Code, Obsidian, etc.)
- **CLI + GUI**: Use both terminal and desktop interface
- **Fuzzy Search**: Fast in-memory search with fuzzy matching
- **CodeMirror Editor**: Syntax highlighting for various languages

## Installation

### Prerequisites

- Go 1.21+
- Node.js 18+ and pnpm
- Wails v2 CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Build

```bash
# Install dependencies
go mod download
cd frontend && pnpm install && cd ..

# Build frontend
cd frontend && pnpm run build && cd ..

# Build application
wails build
```

## Configuration

SnipGo supports configuration through environment variables and config file:

1. **Environment Variable** (highest priority):
   ```bash
   export SNIPGO_DATA_DIR="/path/to/your/snippets"
   ```

2. **Config File** (`~/.config/snip-go/config.yaml`):
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
# Add a new snippet (opens $EDITOR)
snipgo add

# List all snippets
snipgo list

# Search snippets
snipgo search "docker"

# Copy snippet body to clipboard
snipgo copy "docker"
```

### GUI

Run the GUI application:

```bash
wails dev  # Development mode
# or
./bin/snip-go  # Built application
```

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
snip-go/
├── cmd/snipgo/       # CLI entry point
├── internal/
│   ├── core/        # Business logic (includes frontmatter parsing)
│   ├── config/       # Configuration management
│   └── storage/      # File system operations
├── app/              # Wails backend
├── frontend/         # React frontend
└── main.go           # Wails entry point
```

## License

MIT

