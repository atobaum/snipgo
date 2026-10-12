# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SnipGo is a local-first snippet manager with dual interfaces (CLI + GUI). All snippets are stored as Markdown files with YAML frontmatter in `~/.config/snipgo/snippets/`, allowing editing with any text editor.

## Important Conventions

### Git Workflow

**Language:** All commits, PR titles, PR descriptions, and branch names must be written in **English**.

**CRITICAL: Always branch before coding!**
- NEVER commit directly to main branch
- ALWAYS create a feature branch BEFORE making any code changes
- All changes go through PR review

**Starting work:**
1. Create a new branch from the latest main branch
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feature/branch-name
   ```
2. Make your changes on the feature branch
3. Push and create PR into main

**Creating PR:**
- Use `gh` CLI to create PR into main branch
  ```bash
  git push -u origin feature/branch-name
  gh pr create --base main
  ```

### Feature Development Workflow

**Plan Document Management:**
- Create a plan document in `prompts/` directory before starting new feature development
- File naming: `{feature-name}-plan.md` (e.g., `search-filter-plan.md`)
- After completion: `{feature-name}-implementation-summary.md`

**Plan Document Structure:**
```markdown
# {Feature Name} Implementation Plan

## Project Goal
- Brief description of the goal

## Requirements
1. Requirement 1
2. Requirement 2

## Design Decisions
- Key design decisions

## Implementation Steps
### Step 1: ...
### Step 2: ...

## Critical Files
- List of files to modify

## Backward Compatibility
- Impact on existing functionality
```

**Workflow:**
1. Gather requirements → Create `prompts/{feature}-plan.md`
2. Review and finalize plan
3. Implement step by step
4. After completion, write `prompts/{feature}-implementation-summary.md`

**Claude Code Plan Mode:**
- When using Claude Code's plan mode, save the finalized plan to `prompts/` directory before exiting plan mode
- File naming: `{feature-name}-plan.md`

### Test-Driven Development (TDD)

**CRITICAL: Write tests BEFORE implementation code!** Follow Red-Green-Refactor.

**When to apply TDD:**
- New features with clear input/output behavior
- Bug fixes (write test that reproduces bug first)
- Refactoring existing code

**Exceptions (where TDD may not apply):**
- Exploratory/prototype code (but add tests before merging)
- Pure UI layout changes (use visual testing instead)
- Third-party integration spikes

### Commit Message Format

```
<scope>(<action>): <subject>

<body> (optional)
```

**Scope:** feat, fix, refactor, test, style, docs
**Actions:** cli, gui, core, docs, ci

**Examples:**
- `feat(gui): add fuzzy search for snippet titles`
- `fix(cli): resolve frontmatter parsing error with special characters`
- `refactor(core): extract snippet validation logic`
- `docs: add API documentation`

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

## Common Pitfalls

1. **Never commit `.snipgo/` directory** (user data)
2. **Always use RWMutex** when accessing Manager.snippets map
3. **Return copies, not pointers** from Manager methods (prevent mutations)
4. **Call `ReloadSnippets()`** after external saves to sync state
5. **Use bridge.ts functions** for type conversion, not raw Wails bindings
6. **Don't use `--no-verify` or `--no-gpg-sign`** in git commits unless explicitly requested
7. **Test files belong in same package** as tested code (not separate `_test` package)
