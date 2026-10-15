# Features

남은 기능과 새로 제안된 기능을 모아 둡니다. 방향과 원칙은 [PRD](PRD.md)를 따르고, 기능을 끝내면 상태를 바꿉니다.

상태 표기는 `⏳ 대기`, `🚧 진행 중`, `✅ 완료`를 씁니다.

## Phase 2: Usability

### ✅ Hot Reload

- **내용.** fsnotify로 저장소 디렉토리와 하위 디렉토리를 감시하고, CLI나 다른 에디터, git pull로 파일이 바뀌면 GUI 목록과 열린 스니펫을 바로 갱신합니다. `.git` 같은 숨김 디렉토리와 `.md`가 아닌 파일은 무시하고, 짧은 시간에 몰린 변경은 한 번에 반영합니다.
- **편집 중 충돌.** 저장하지 않은 수정이 있으면 덮어쓰지 않고 에디터에 "Changed on disk" 표시만 띄웁니다. 열린 스니펫이 밖에서 삭제되면 선택을 해제합니다.
- **자기 저장 무시.** 앱이 저장, 이름 변경, 삭제한 직후의 파일 상태를 기록해 두고, 그 상태 그대로인 파일의 이벤트는 무시합니다.
- **관련.** git 동기화의 사전 작업입니다.

### ✅ `new` Command Flags

- **내용.** `-t/--title`, `--tags`, `-L/--language`(`--lang` 별칭)로 frontmatter를 미리 채웁니다. `-e`와 함께 쓰면 채운 frontmatter와 본문을 에디터에서 바로 고칠 수 있습니다.
- **클립보드와 stdin 입력.** `-c/--clipboard`이면 클립보드에서, stdin이 파이프이면 stdin에서 본문을 읽습니다. 끝의 줄바꿈 하나만 지우고 나머지는 그대로 저장합니다. `-t`까지 주면 프롬프트 없이 끝나므로 `pbpaste | snipgo new -t "List ports" --tags net,mac -L bash`처럼 한 줄로 저장할 수 있습니다.
- **충돌 규칙.** `--clipboard`와 파이프 stdin, `--clipboard`나 파이프 stdin과 `-m`, `-m`과 `-e`는 함께 쓸 수 없습니다. stdin이 파이프인데 `-t`가 없으면 제목을 `/dev/tty`로 묻고, 터미널이 없으면 `-t`를 쓰라는 오류를 냅니다.

### ⏳ GUI Favorite Filter & Sort

- **내용.** 즐겨찾기만 보기, 즐겨찾기를 위로 올리는 정렬을 지원합니다.

## Phase 3: Polish

### ⏳ TUI

- **내용.** CLI에서도 GUI와 비슷하게 목록 탐색, 미리보기, 검색, 복사, 실행, 편집을 한 화면에서 할 수 있게 합니다. bubbletea 사용을 검토합니다.
- **fzf 연동.** 대체하지 않고 함께 둡니다. TUI는 둘러보고 편집하는 메인 화면으로, fzf는 zsh widget처럼 빠르게 고르는 용도로 씁니다.

### ✅ Git Sync

- **내용.** 스니펫 디렉토리를 git 저장소로 두고 `snipgo sync`나 GUI 자동 동기화로 commit, pull, push 를 실행합니다. commit은 스니펫 디렉토리 안의 `.md` 파일 변경만 포함하고(숨김 디렉토리 제외), pull은 `--no-rebase`로 merge 합니다. upstream이 있는 브랜치가 필요하고, 저장소 설정은 사용자가 한 번 직접 합니다.
- **설정.** `sync.enabled`(기본 꺼짐), `sync.interval_minutes`(기본 10, 최소 1), `sync.conflict`(`manual`, `local`, `remote`)를 `snipgo config set`이나 GUI 설정의 Git Sync 섹션에서 바꿉니다.
- **자동 동기화.** 켜져 있으면 GUI가 시작할 때와 설정한 주기마다 실행합니다. 수동 동기화는 켜짐 여부와 상관없이 `snipgo sync`나 GUI의 Sync now 버튼으로 실행하고, 실패하면 `snipgo sync`는 종료 코드 1을 돌려줍니다.
- **충돌 처리.** `manual`은 merge를 중단하고 충돌 파일을 알려 주며, 수동 동기화가 성공할 때까지 자동 동기화를 멈춥니다. `local`은 이 기기 내용을, `remote`는 원격 내용을 우선합니다. 한쪽에서 수정하고 다른 쪽에서 삭제한 충돌은 모드와 상관없이 항상 `manual`처럼 멈춥니다.
- **큰 저장소 안에 둘 때.** 스니펫 디렉토리가 dotfiles 저장소나 Obsidian vault 같은 더 큰 저장소 안에 있으면 sync는 브랜치 전체를 pull해서 merge 하고, 그 브랜치에서 아직 push 하지 않은 로컬 commit을 모두 push 합니다. `local`, `remote` 충돌 모드도 스니펫만이 아니라 merge 전체에 적용됩니다. 그래서 스니펫 전용 저장소를 따로 두기를 권장합니다.
- **인증과 제한.** 인증은 사용자의 git 설정(ssh 키, credential helper)을 그대로 씁니다. `GIT_TERMINAL_PROMPT=0`으로 실행해 프롬프트 없이 실패하고, 한 번의 실행은 2분 안에 끝나지 않으면 중단합니다.
- **알림.** 수동 동기화는 결과를 항상 toast로 보여 줍니다. 자동 동기화는 오류만 보여 주고, 같은 오류가 반복되면 한 번만 보여 줍니다.
- **의존.** Hot Reload

### ⏳ Theme

- **내용.** Light와 Dark 모드, 에디터 테마를 바꿀 수 있게 합니다.

### ⏳ Image Export

- **내용.** 코드 스니펫을 이미지로 캡처합니다.
