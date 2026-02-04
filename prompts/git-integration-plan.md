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

### Phase 1: Core Infrastructure (MVP) ✅ IN PROGRESS
- [x] Create `internal/git/git.go` package
- [x] Implement GitManager struct and methods
- [x] Add Git config to `internal/config/config.go`
- [ ] Implement `snipgo git init` command
- [ ] Implement `snipgo git <command>` pass-through

### Phase 2: Commit Integration
- [ ] Integrate GitManager with Manager (Save/Delete hooks)
- [ ] Implement `snipgo commit` command
- [ ] Implement `snipgo git status` command
- [ ] Add commit message template support

### Phase 3: Sync Commands
- [ ] Implement `snipgo sync` (pull + push)
- [ ] Implement `snipgo push` / `snipgo pull`
- [ ] Implement `snipgo git clone` command
- [ ] Add error handling and guidance messages

### Phase 4: History & Restore
- [ ] Implement `snipgo history <id>` command
- [ ] Implement `snipgo restore <id> <commit>` command
- [ ] Add diff display support

### Phase 5: GUI Integration
- [ ] Add Git status display in status bar
- [ ] Add Sync button
- [ ] Add conflict notification Modal
- [ ] Add history viewer (optional)

---

## Critical Files

### New Files
- `internal/git/git.go` - GitManager core logic
- `internal/git/git_test.go` - Tests
- `cmd/snipgo/git.go` - git subcommand
- `cmd/snipgo/commit.go` - commit command
- `cmd/snipgo/sync.go` - sync command
- `cmd/snipgo/push.go` - push command
- `cmd/snipgo/pull.go` - pull command
- `cmd/snipgo/history.go` - history command
- `cmd/snipgo/restore.go` - restore command

### Modified Files
- `internal/core/manager.go` - GitManager integration
- `internal/config/config.go` - Git config addition
- `cmd/snipgo/main.go` - New command registration
- `app/app.go` - GUI Git method exposure (Phase 5)

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

## Progress Log

### 2026-02-04
- Started implementation
- Explored codebase architecture
- Created this plan document
