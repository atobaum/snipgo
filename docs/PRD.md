# Product Requirements Document: SnipGo

| 항목 | 내용 |
|------|------|
| Project Name | SnipGo |
| Version | 1.1.0 |
| Type | Hybrid Snippet Manager (CLI + GUI) |
| Philosophy | Local First, File over App |
| Stack | Go, Wails v2, Cobra, React |
| Storage | Local File System (Markdown) |
| Status | Approved for Development |

## 1. Overview

**SnipGo**는 개발자를 위한 **Local-First** 명령어 스니펫 관리 도구입니다. 터미널에서 실행하거나 붙여 넣는 명령어를 모아 두고 바로 꺼내 씁니다. 스니펫은 데이터베이스나 독자 포맷이 아니라 **Plain Text**(Markdown) 파일로 로컬에 저장합니다. 앱이 사라져도 데이터는 남아야 한다는 **File over App** 철학을 따릅니다.

터미널(CLI, TUI)과 데스크톱(GUI)을 자유롭게 오가며 같은 스니펫을 다룰 수 있습니다. 앱을 지워도 스니펫은 평범한 텍스트 파일로 그대로 남습니다.

### Why

터미널용 스니펫 도구는 명령어를 빠르게 꺼내 쓰기 좋습니다. 하지만 스니펫이 쌓이면 둘러보고 정리하고 고치기가 번거롭습니다. GUI 스니펫 도구는 보고 고치기 편하지만 터미널 작업 흐름과 따로 놉니다. SnipGo는 명령어에 집중합니다. 터미널에서는 바로 찾아 실행하고, 쌓인 명령어는 TUI나 GUI로 둘러보며 정리합니다.

### Who

- 하루 대부분을 터미널에서 일하고, 자주 쓰는 명령어를 외우기보다 꺼내 쓰고 싶은 개발자
- 셸 명령어와 터미널에 붙여 넣는 쿼리나 설정 조각을 모아 두고, 가끔은 한눈에 둘러보며 정리하고 싶은 사람
- 여러 기기를 오가며 dotfiles 같은 개인 자산을 git으로 관리하는 데 익숙한 사람
- 특정 앱에 데이터를 맡기기보다 plain text와 손에 익은 에디터를 선호하는 사람

### Core Principles

1. **Local First & Ownership.** 모든 데이터는 먼저 사용자 로컬 머신에 저장합니다. 클라우드에 기대지 않으므로 데이터는 온전히 사용자 것입니다.
2. **File over App.** Markdown 파일이 원본(Source of Truth)입니다. SnipGo 없이도 VS Code나 Obsidian 같은 다른 에디터로 자유롭게 열어 고칠 수 있습니다.
3. **Terminal First.** 가장 자주 쓰는 흐름은 터미널에서 찾아 실행하는 것입니다. 새 기능은 CLI와 TUI에 먼저 넣고, GUI는 둘러보고 정리하는 보조 화면으로 유지합니다.
4. **Performance.** 무거운 IDE 기능은 넣지 않고 검색과 편집이 즉시 반응하는 데 집중합니다.
5. **Git as Storage.** 별도 서버나 계정 없이 git 저장소 하나로 이력 관리, 백업, 기기 간 동기화를 해결합니다.

### Non-Goals

- 자체 클라우드 서버나 계정 시스템
- 실시간 팀 협업 (공유가 필요하면 git 저장소를 함께 씁니다)
- 독자 포맷이나 데이터베이스 저장소
- 일반 코드 조각이나 노트 보관 (노트 도구에 맡깁니다)
- 모바일 앱
- LSP나 디버깅 같은 IDE 수준의 편집 기능

## 2. Tech Stack

- **Language:** Go
- **CLI Framework:** Cobra (github.com/spf13/cobra)
- **GUI Framework:** Wails v2 (Go + Web Frontend)
- **Frontend:** React + Vite + TypeScript
- **Editor Library:** CodeMirror 6
- **Styling:** TailwindCSS
- **Data Parsing:** gopkg.in/yaml.v3 (Frontmatter)
- **File Watcher:** fsnotify/fsnotify (실시간 파일 변경 감지)

## 3. Data Schema

- **저장 위치:** `~/.config/snipgo/snippets/`
- **파일명 규칙:** `{Title}.md`. 제목을 바꾸면 파일 이름도 바뀌고, 같은 이름이 있으면 `{Title}-2.md`처럼 번호를 붙입니다. 글자, 숫자, `-`, `_`, `.` 외의 문자는 `_`로 바꾸고 최대 100자까지 씁니다.
- **파일 포맷:** YAML Frontmatter + Markdown Body

### Example: docker-compose-setup.md

```yaml
---
id: "550e8400-e29b-41d4-a716-446655440000"   # UUID v4
title: "Docker Compose Setup"                # 스니펫 제목
tags: ["docker", "devops"]                   # 태그 목록
language: "yaml"                             # 문법 강조용 언어
is_favorite: true                            # 즐겨찾기 여부
created_at: 2025-12-20T10:00:00Z             # 만든 시각 (바뀌지 않음)
updated_at: 2025-12-25T14:30:00Z             # 고친 시각 (저장할 때마다 갱신)
---

version: '3'
services:
  web:
    image: nginx
```

## 4. Functional Requirements

### 4.1 Backend Core

- **Snippet Manager**
  - `LoadAll()` 시작할 때 저장소 디렉토리 안의 모든 `.md` 파일을 읽어 메모리에 인덱싱합니다.
  - `Save()` 필수 필드를 검사한 뒤 Frontmatter와 본문을 합쳐 저장합니다.
  - `Delete()` 파일을 영구 삭제합니다.
- **File Watcher (Hot Reload)**
  - fsnotify로 저장소 폴더를 감시합니다.
  - CLI나 다른 에디터에서 파일을 추가, 수정, 삭제하면 GUI 목록을 바로 갱신해서 화면과 파일 내용을 일치시킵니다.
- **Search Engine (In-Memory)**
  - **Title** Fuzzy Search (오타와 약어 허용)
  - **Tags/Description/Body** Substring Matching
  - 태그(AND)와 언어로 필터링하고, 사용 빈도, 이름, 최근 사용 순으로 정렬합니다.

### 4.2 CLI

- **`snipgo new`**
  - 프롬프트로 제목과 본문을 입력받아 스니펫을 만듭니다. `-m`이면 여러 줄 본문을 입력받습니다.
  - `-t "Title" --tags "go,api" -L bash` 같은 플래그로 Frontmatter를 미리 채웁니다. `-t`를 주면 제목을 묻지 않습니다.
  - 본문은 `-c`(클립보드), 파이프로 들어온 stdin, 프롬프트 순으로 고릅니다. 끝의 줄바꿈 하나만 지우고 그대로 저장하며, 비어 있으면 오류를 냅니다. `pbpaste | snipgo new -t "List ports"`처럼 프롬프트 없이 저장할 수 있습니다.
  - `-e`이면 Frontmatter와 본문(클립보드나 stdin에서 읽은 내용)을 채운 문서를 `$EDITOR`로 엽니다.
  - `-c`와 파이프 stdin, `-c`나 파이프 stdin과 `-m`, `-m`과 `-e`는 함께 쓸 수 없습니다. stdin이 파이프일 때 제목 프롬프트는 `/dev/tty`에서 읽고, 터미널이 없으면 `-t`를 요구합니다.
- **`snipgo list`**
  - 짧은 ID, Title, Tags, Language, 즐겨찾기 여부를 표로 출력합니다.
- **`snipgo search`**
  - fzf로 스니펫을 골라 본문을 출력합니다. `-q`로 검색어를, `--tag`와 `--lang`으로 필터를 줍니다.
- **`snipgo exec`**
  - fzf로 스니펫을 골라 셸 명령으로 실행합니다. 변수가 있으면 실행 전에 미리보기를 보여줍니다.
- **`snipgo edit`**
  - fzf로 스니펫을 골라 `$EDITOR`로 엽니다.
- **`snipgo copy <query>`**
  - 가장 잘 맞는 검색 결과 하나의 **코드 본문**(Body)만 클립보드에 복사합니다.
- **변수 템플릿**
  - 본문의 `${VAR}`는 `search`, `exec`, `copy` 때 값을 입력받아 채웁니다. `-v KEY=VALUE`로 값을 넘기거나 `--raw`로 그대로 출력할 수 있습니다.
- **`snipgo config`**
  - 설정을 보고(`show`) 바꿉니다(`set`).

### 4.3 GUI

- **View & Edit**
  - 위쪽에는 Title, Tags(Chips 입력), 메타 정보를, 아래쪽에는 CodeMirror 에디터를 둡니다.
  - **Raw Mode**에서는 Frontmatter를 포함한 전체 텍스트를 직접 볼 수 있습니다.
- **Sync**
  - Backend File Watcher의 이벤트를 받아 목록을 자동으로 갱신합니다.
- **Convenience**
  - 클립보드 복사 버튼
  - 즐겨찾기 토글 버튼

## 5. Milestones

남은 기능의 상세 내용과 진행 상태는 [FEATURES](FEATURES.md)에서 관리합니다.

### 🚨 Phase 1: MVP

*핵심 가치를 검증할 최소 기능 제품*

- **Core:** Go 프로젝트 구조(Clean Architecture), Markdown 입출력, YAML 파싱
- **CLI:** new, list, search, copy
- **GUI:** Wails 초기 설정, 목록 화면, 상세 보기와 편집(Raw), 클립보드 복사

### ⚠️ Phase 2: Usability

*실제로 쓰면서 겪는 불편 없애기*

- **Sync:** fsnotify 기반 Hot Reload (CLI에서 고치면 GUI에 반영)
- **CLI:** new 명령어 플래그(`-t`, `--tags`)
- **GUI:** Chips 형태 태그 입력, 즐겨찾기 필터와 정렬

### 🎡 Phase 3: Polish

*완성도를 높이는 부가 기능*

- **Export:** 코드 스니펫 이미지 캡처
- **CLI:** bubbletea 등을 활용한 TUI (대화형 검색과 선택)
- **Theme:** Light/Dark 모드와 에디터 테마 설정
- **Sync:** git 저장소 동기화. Local First 원칙을 지키면서 원하는 사람만 켜서 씁니다.
