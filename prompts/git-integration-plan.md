# SnipGo Git Integration - Implementation Plan

## Project Goal
SnipGo 스니펫 데이터를 Git으로 버전 관리하고 여러 기기 간 동기화할 수 있는 기능 추가.
기존 local-first 철학을 유지하면서 Git을 opt-in 확장 기능으로 제공.

## Design Principles
1. **Git은 Optional**: Git 없이도 기존처럼 완벽히 동작
2. **Local-First**: 네트워크 없이도 모든 기능 사용 가능
3. **Manual Commit, Manual Sync**: 사용자가 명시적으로 제어 (auto_commit 옵션으로 자동화 가능)
4. **Transparent**: 사용자가 원하면 직접 git 명령어 사용 가능

## Implementation Phases

### Phase 1: Core Infrastructure (MVP) ✅ COMPLETED
- [x] Create `internal/git/git.go` package
- [x] Implement GitManager struct and methods
- [x] Add Git config to `internal/config/config.go`
- [x] Implement `snipgo git init` command
- [x] Implement `snipgo git <command>` pass-through

### Phase 2: Commit Integration ✅ COMPLETED
- [x] Integrate GitManager with Manager (Save/Delete hooks)
- [x] Implement `snipgo commit` command
- [x] Implement `snipgo git status` command
- [x] Add commit message template support

### Phase 3: Sync Commands ✅ COMPLETED
- [x] Implement `snipgo sync` (pull + push)
- [x] Implement `snipgo push` / `snipgo pull`
- [x] Implement `snipgo git clone` command
- [x] Add error handling and guidance messages

### Phase 4: History & Restore ✅ COMPLETED
- [x] Implement `snipgo history <id>` command
- [x] Implement `snipgo restore <id> <commit>` command
- [x] Add diff display support

### Phase 5: GUI Integration 🔲 PENDING
- [ ] Add Git status display in status bar
- [ ] Add Sync button
- [ ] Add conflict notification Modal
- [ ] Add history viewer (optional)

---

## Critical Files

### New Files (Created)
- `internal/git/git.go` - GitManager core logic ✅
- `internal/git/git_test.go` - Tests ✅
- `cmd/snipgo/git.go` - git subcommand ✅
- `cmd/snipgo/commit.go` - commit command ✅
- `cmd/snipgo/sync.go` - sync command ✅
- `cmd/snipgo/push.go` - push command ✅
- `cmd/snipgo/pull.go` - pull command ✅
- `cmd/snipgo/history.go` - history command ✅
- `cmd/snipgo/restore.go` - restore command ✅

### Modified Files
- `internal/core/manager.go` - GitManager integration ✅
- `internal/config/config.go` - Git config addition ✅
- `CLAUDE.md` - Updated workflow guidelines ✅
- `app/app.go` - GUI Git method exposure (Phase 5 - pending)

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI / GUI                            │
├─────────────────────────────────────────────────────────┤
│                    Manager Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │    Save()    │  │   Delete()   │  │  LoadAll()   │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │           │
│         ▼                 ▼                 ▼           │
│  ┌──────────────────────────────────────────────────┐  │
│  │              GitManager (NEW)                     │  │
│  │  - AddAndCommit()                                 │  │
│  │  - Sync() / Push() / Pull()                       │  │
│  │  - GetHistory()                                   │  │
│  │  - IsGitRepo()                                    │  │
│  └──────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│                   FileSystem Layer                       │
└─────────────────────────────────────────────────────────┘
```

---

## Configuration

```yaml
# ~/.config/snipgo/config.yaml
git:
  enabled: true              # Git 기능 활성화 여부
  auto_commit: false         # 저장 시 자동 커밋 (기본: 비활성)
  auto_push: false           # 커밋 후 자동 push (기본: 비활성)
  commit_message_template: "Update: {{.Title}}"
  remote: "origin"
  branch: "main"
```

---

## CLI Commands Summary

| Command | Description |
|---------|-------------|
| `snipgo git init` | Initialize git repository for snippets |
| `snipgo git clone <url>` | Clone existing snippets repository |
| `snipgo git status` | Show git status |
| `snipgo git <command>` | Pass-through any git command |
| `snipgo commit [-m msg]` | Commit all changes |
| `snipgo push` | Push to remote |
| `snipgo pull` | Pull from remote |
| `snipgo sync` | Pull + Push |
| `snipgo history <id>` | Show snippet version history |
| `snipgo restore <id> <hash>` | Restore snippet to previous version |
| `snipgo config set git.enabled true` | Enable git features |
| `snipgo config set git.auto_commit true` | Enable auto-commit on save |

---

## Progress Log

### 2026-02-04
- Started implementation
- Explored codebase architecture
- Created this plan document
- Completed Phase 1-4 (CLI implementation)
- Added comprehensive tests for GitManager
- 7 commits total on feature/git-integration branch

### Commits:
1. `feat(core): add GitManager package for git integration`
2. `docs: add git integration plan and update workflow guidelines`
3. `feat(cli): add git subcommands (init, status, clone, pass-through)`
4. `feat(cli): add commit, push, pull, sync commands`
5. `feat(cli): add history and restore commands for snippet versioning`
6. `test(git): add comprehensive tests for GitManager`
7. `feat(core): integrate GitManager with Manager for auto-commit`
