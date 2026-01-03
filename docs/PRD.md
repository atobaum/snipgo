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

## 1. 개요 (Overview)

**SnipGo**는 개발자를 위한 **Local-First** 고성능 스니펫 관리 도구입니다.  
이 소프트웨어는 "앱은 사라져도 데이터는 남아야 한다"는 **File over App** 철학을 따릅니다. 데이터베이스나 독자 규격 파일이 아닌, 가장 보편적이고 지속 가능한 포맷인 **Plain Text (Markdown)**를 로컬 저장소(`~/.snipgo`)에 직접 저장합니다.  
사용자는 터미널(CLI)과 데스크탑(GUI)을 자유롭게 오가며 작업할 수 있으며, 앱을 삭제하더라도 귀하의 데이터는 단순한 텍스트 파일로 온전히 남아있습니다.

### 핵심 원칙 (Core Principles)

1. **Local First & Ownership:** 모든 데이터는 사용자 로컬 머신에 우선 저장됩니다. 클라우드 의존성 없이 완전한 데이터 소유권을 보장합니다.
2. **File over App:** 앱은 데이터를 보여주는 '렌즈'일 뿐입니다. 소스 파일(Markdown)이 진실의 원천(Source of Truth)이며, 다른 에디터(VS Code, Obsidian 등)에서도 자유롭게 열고 수정할 수 있습니다.
3. **Universal Access:** CLI(Cobra)의 신속함과 GUI(Wails)의 가시성을 동시에 제공하여, 어떤 환경에서든 끊김 없는 워크플로우를 보장합니다.
4. **Performance:** 무거운 IDE 기능을 배제하고, 즉각적인 검색과 편집 속도(Low Latency)에 집중합니다.

## 2. 기술 스택 (Tech Stack)

- **Language:** **Go (Golang)**
- **CLI Framework:** **Cobra** (github.com/spf13/cobra)
- **GUI Framework:** **Wails v2** (Go + Web Frontend)
- **Frontend:** React + Vite + TypeScript
- **Editor Library:** **CodeMirror 6** (Raw Editing Mode)
- **Styling:** TailwindCSS
- **Data Parsing:** gopkg.in/yaml.v3 (Frontmatter Handling)
- **File Watcher:** fsnotify/fsnotify (실시간 파일 변경 감지용)

## 3. 데이터 구조 (Data Schema)

- **저장 위치:** `~/.config/snipgo/snippets/`
- **파일명 규칙:** `{Title}_{Timestamp}.md` (중복 방지 및 가독성 확보)
- **파일 포맷:** YAML Frontmatter + Markdown Body

### 예시 파일 (docker-compose-setup.md)

```yaml
---
id: "550e8400-e29b-41d4-a716-446655440000"   # UUID v4
title: "Docker Compose Setup"                # 스니펫 제목
tags: ["docker", "devops"]                   # 태그 목록
language: "yaml"                             # Syntax Highlighting용
is_favorite: true                            # 즐겨찾기 여부 (Boolean)
created_at: 2025-12-20T10:00:00Z             # 생성 일시 (불변)
updated_at: 2025-12-25T14:30:00Z             # 수정 일시 (저장 시 갱신)
---

version: '3'
services:
  web:
    image: nginx
```

## 4. 상세 기능 명세 (Functional Requirements)

### 4.1 Backend Core (Go)

- **Snippet Manager:**
  - `LoadAll()`: 실행 시 `~/.snipgo` 내 모든 `.md` 파일을 파싱하여 메모리 인덱싱.
  - `Save()`: Frontmatter와 본문을 합성하여 저장. (들여쓰기 등 YAML 문법 검증 포함).
  - `Delete()`: 파일 영구 삭제.
- **File Watcher (Hot Reload):**
  - fsnotify를 사용하여 저장소 폴더 감시.
  - 외부(CLI, 타 에디터)에서 파일 변경/추가/삭제 시, GUI 리스트를 실시간으로 갱신하여 데이터 무결성 유지.
- **Search Engine (In-Memory):**
  - **Title:** Fuzzy Search 적용 (오타 및 약어 허용).
  - **Tags/Body:** Substring Matching.

### 4.2 CLI (cmd/snipgo)

- **`snipgo add`:**
  - 기본: `$EDITOR`를 열어 빈 파일 생성.
  - **개선 (Interactive):** 플래그 지원 (`-t "Title" --tags "go,api"`). 실행 시 Frontmatter가 미리 채워진 상태로 에디터 오픈.
- **`snipgo list`:**
  - ID(Short), Title, Tags, Language, Favorite 여부를 테이블로 출력.
- **`snipgo search <query>`:**
  - 검색 결과를 리스트로 출력.
- **`snipgo copy <query>`:**
  - 검색 결과(Top 1)의 **코드 본문(Body)**만 시스템 클립보드에 복사.

### 4.3 GUI (Wails)

- **View & Edit:**
  - **Split UI:** (상단) Title, Tags(Chips Input), Meta Info / (하단) CodeMirror Editor.
  - **Raw Mode:** 필요시 전체 텍스트(Frontmatter 포함)를 직접 수정하는 모드 지원.
- **Sync:**
  - Backend의 File Watcher 이벤트를 수신하여 리스트 자동 갱신.
- **Convenience:**
  - Copy to Clipboard 버튼.
  - Is Favorite 토글 버튼.

## 5. 개발 로드맵 (Milestones)

### 🚨 Phase 1: MVP (Must Have) - "핵심 가치 검증"

*배포 가능한 최소 기능 제품*

- **Core:** Go 프로젝트 구조(Clean Architecture), Markdown I/O, YAML 파싱 로직.
- **CLI:** add(기본), list, search, copy 구현.
- **GUI:** Wails 초기화, 리스트 뷰, 상세 보기/수정(Raw), 클립보드 복사.

### ⚠️ Phase 2: Usability (Should Have) - "사용성 강화"

*실사용 시 불편함 제거*

- **Sync:** fsnotify 기반 Hot Reload 구현 (CLI 수정 → GUI 반영).
- **CLI:** add 명령어의 Interactive Flag (`-t`, `--tags`) 구현.
- **GUI:** 태그 입력 UI 개선 (Chips 형태), `is_favorite` 필터링 및 정렬.

### 🎡 Phase 3: Polish (Could Have) - "완성도"

*심미적 요소 및 고급 기능*

- **Export:** 코드 스니펫 이미지 캡처 (Carbon 스타일).
- **CLI:** bubbletea 등을 활용한 Interactive Search/Select TUI.
- **Theme:** Light/Dark 모드 및 에디터 테마 커스텀.
- **Sync:** 클라우드 동기화 기능 (GitHub Gist, Git 저장소 등). Local-first 원칙을 유지하면서 선택적 동기화 제공.
