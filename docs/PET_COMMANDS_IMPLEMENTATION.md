# pet 명령어 구현 계획

## pet --help 결과 분석

pet의 명령어 목록:
- `clip` - Copy the selected commands (snipgo의 `copy`와 유사) ✅
- `completion` - Generate autocompletion script (미구현)
- `configure` - Edit config file (snipgo의 `config`와 유사) ✅
- `edit` - Edit snippet file (미구현)
- `exec` - Run the selected commands ✅
- `help` - Help about any command (cobra 기본 제공) ✅
- `list` - Show all snippets ✅
- `new` - Create a new snippet ✅
- `search` - Search snippets ✅
- `sync` - Sync snippets (roadmap에 추가됨, 구현 제외)
- `version` - Print the version number (미구현)

## 현재 상태 분석

### 구현된 명령어

- `new`: 새 스니펫 생성
- `list`: 스니펫 목록 출력
- `search`: 스니펫 검색 (fzf 사용)
- `copy`: 스니펫 본문을 클립보드로 복사 (pet의 `clip`과 유사)
- `exec`: 스니펫을 셸 명령어로 실행
- `config`: 설정 관리 (show, set) (pet의 `configure`와 유사)

### 미구현 명령어

1. **`edit`**: CLI에서 스니펫 편집
   - `serializeSnippetForEdit`, `parseSnippetFromEdit` 함수는 이미 존재하지만 CLI 명령어로 연결되지 않음
   - `$EDITOR` 환경 변수를 사용하여 임시 파일을 열고 편집 후 저장

2. **`version`**: 버전 정보 출력
   - `rootCmd.Version`은 설정되어 있지만 명령어로는 없음
   - `snipgo version` 명령어로 버전 정보 출력

3. **`completion`**: 자동완성 스크립트 생성 (zsh)
   - Cobra의 completion 기능 활용
   - `snipgo completion zsh` 명령어로 zsh 자동완성 스크립트 생성

## 구현 계획

### 1. `edit` 명령어 구현

**파일**: `cmd/snipgo/main.go`

**기능**:
- fzf를 사용하여 편집할 스니펫 선택
- `$EDITOR` 환경 변수 확인 (없으면 기본 에디터 사용)
- 임시 파일에 스니펫 내용 저장 (frontmatter 포함)
- 에디터로 파일 열기
- 편집 완료 후 파일 읽어서 파싱
- `Manager.Save`로 저장
- 임시 파일 삭제

**구현 세부사항**:
- `editCmd` 변수 추가 및 `init()`에 등록
- `runEdit` 함수 구현
- `$EDITOR` 환경 변수 확인 (기본값: `vi`)
- `serializeSnippetForEdit`, `parseSnippetFromEdit` 함수 활용
- 에러 처리: 에디터 실행 실패, 파싱 실패 등

**코드 예시**:
```go
var editCmd = &cobra.Command{
    Use:   "edit",
    Short: "Edit a snippet",
    Long:  "Interactively select a snippet using fzf and edit it with $EDITOR",
    Args:  cobra.NoArgs,
    RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
    // 1. fzf로 스니펫 선택
    // 2. 임시 파일 생성 및 스니펫 내용 저장
    // 3. $EDITOR 실행
    // 4. 편집된 내용 파싱
    // 5. Manager.Save로 저장
    // 6. 임시 파일 삭제
}
```

### 2. `version` 명령어 구현

**파일**: `cmd/snipgo/main.go`

**기능**:
- 버전 정보 출력 (version, commit, date)
- `rootCmd.Version`을 활용하거나 별도 명령어로 구현

**구현 세부사항**:
- `versionCmd` 변수 추가 및 `init()`에 등록
- `runVersion` 함수 구현
- 버전, 커밋 해시, 빌드 날짜 출력

**코드 예시**:
```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number",
    Long:  "Print the version, commit hash, and build date",
    Args:  cobra.NoArgs,
    RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
    fmt.Printf("snipgo version %s\n", version)
    fmt.Printf("commit: %s\n", commit)
    fmt.Printf("date: %s\n", date)
    return nil
}
```

### 3. `completion` 명령어 구현

**파일**: `cmd/snipgo/main.go`

**참고 문서**:
- [Cobra Shell Completion Guide](https://cobra.dev/docs/how-to-guides/shell-completion/)
- [pet zsh completion 예시](https://github.com/knqyf263/pet/blob/main/misc/completions/zsh/_pet)

**기능**:
- zsh 자동완성 스크립트 생성
- Cobra의 내장 completion 기능 활용
- `snipgo completion zsh` 실행 시 zsh 자동완성 스크립트 출력
- 모든 명령어와 플래그에 대한 자동완성 지원

**구현 방법**:

Cobra는 두 가지 방법을 제공합니다:
1. **자동 생성 방식** (권장): Cobra의 `GenZshCompletion()` 메서드 사용
2. **수동 작성 방식**: pet처럼 수동으로 zsh completion 스크립트 작성

snipgo는 **자동 생성 방식**을 사용합니다. Cobra가 모든 명령어와 플래그를 자동으로 인식하여 completion 스크립트를 생성합니다.

**구현 세부사항**:

1. **Completion 명령어 추가**:
   - `completionCmd` 루트 명령어 생성
   - `completionZshCmd` zsh 서브커맨드 생성
   - `rootCmd.GenZshCompletion()` 메서드 호출

2. **코드 구조**:
```go
var completionCmd = &cobra.Command{
    Use:   "completion",
    Short: "Generate completion script",
    Long: `Generate shell completion scripts for snipgo.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, execute once:
  # Linux:
  snipgo completion zsh > "${fpath[1]}/_snipgo"
  
  # macOS:
  snipgo completion zsh > $(brew --prefix)/share/zsh/site-functions/_snipgo
`,
    Args: cobra.NoArgs,
}

var completionZshCmd = &cobra.Command{
    Use:   "zsh",
    Short: "Generate zsh completion script",
    Long: `Generate the autocompletion script for zsh shell.

To load completions in your current shell session:
  source <(snipgo completion zsh)

To load completions for every new session, add to your ~/.zshrc:
  echo 'source <(snipgo completion zsh)' >> ~/.zshrc

Or install to a system-wide location:
  snipgo completion zsh > ~/.zsh/completions/_snipgo
  echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  echo 'autoload -U compinit && compinit' >> ~/.zshrc
`,
    Args: cobra.NoArgs,
    RunE: runCompletionZsh,
}

func runCompletionZsh(cmd *cobra.Command, args []string) error {
    return rootCmd.GenZshCompletion(os.Stdout)
}

func init() {
    completionCmd.AddCommand(completionZshCmd)
    rootCmd.AddCommand(completionCmd)
}
```

3. **자동완성되는 항목**:
   - 모든 서브커맨드: `new`, `list`, `search`, `copy`, `exec`, `edit`, `version`, `config`, `completion`
   - 각 명령어의 플래그: `--help`, `--version` 등
   - `config` 서브커맨드: `show`, `set`
   - `completion` 서브커맨드: `zsh` (향후 bash, fish 등 추가 가능)

**설치 방법**:

1. **임시 활성화** (현재 세션만):
```bash
source <(snipgo completion zsh)
```

2. **사용자별 영구 설치** (권장):
```bash
# completion 디렉토리 생성
mkdir -p ~/.zsh/completions

# completion 스크립트 생성
snipgo completion zsh > ~/.zsh/completions/_snipgo

# ~/.zshrc에 추가
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit && compinit' >> ~/.zshrc

# 또는 한 줄로
echo 'source <(snipgo completion zsh)' >> ~/.zshrc
```

3. **시스템 전체 설치** (macOS Homebrew):
```bash
snipgo completion zsh > $(brew --prefix)/share/zsh/site-functions/_snipgo
```

4. **시스템 전체 설치** (Linux):
```bash
sudo mkdir -p /usr/local/share/zsh/site-functions
sudo snipgo completion zsh > /usr/local/share/zsh/site-functions/_snipgo
```

**테스트 방법**:

1. completion 스크립트 생성 확인:
```bash
snipgo completion zsh | head -20
```

2. 실제 자동완성 테스트:
```bash
# 설치 후
snipgo <TAB>  # 모든 명령어 목록 표시
snipgo new <TAB>  # new 명령어의 플래그 표시
snipgo config <TAB>  # config 서브커맨드 표시
snipgo completion <TAB>  # completion 서브커맨드 표시
```

**향후 확장 가능성**:

1. **Custom Completions**: 특정 플래그에 대한 동적 자동완성 추가
   - `config set` 명령어에서 설정 키 자동완성:
   ```go
   configSetCmd.RegisterFlagCompletionFunc("key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
       return []string{"data_directory"}, cobra.ShellCompDirectiveNoFileComp
   })
   ```
   - `copy` 명령어에서 스니펫 제목 자동완성 (동적):
   ```go
   copyCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
       snippets := manager.GetAll()
       titles := make([]string, 0, len(snippets))
       for _, s := range snippets {
           if strings.HasPrefix(s.Title, toComplete) {
               titles = append(titles, s.Title)
           }
       }
       return titles, cobra.ShellCompDirectiveNoFileComp
   }
   ```

2. **다른 셸 지원**: bash, fish, PowerShell 등
   ```go
   var completionBashCmd = &cobra.Command{
       Use:   "bash",
       Short: "Generate bash completion script",
       RunE:  func(cmd *cobra.Command, args []string) error {
           return rootCmd.GenBashCompletion(os.Stdout)
       },
   }
   ```

3. **파일 경로 자동완성**: `MarkFlagFilename()` 사용
   ```go
   configSetCmd.MarkFlagFilename("config", "yaml", "yml")
   ```

**참고사항**:
- Cobra의 자동 생성 기능을 사용하면 모든 명령어와 플래그가 자동으로 completion에 포함됨
- pet처럼 수동으로 작성할 필요 없이 Cobra가 자동으로 처리
- 필요시 `RegisterFlagCompletionFunc()`를 사용하여 특정 플래그에 대한 동적 자동완성 추가 가능

## 파일 변경 사항

### `cmd/snipgo/main.go`
- `editCmd` 변수 추가 및 `init()`에 등록
- `versionCmd` 변수 추가 및 `init()`에 등록
- `completionCmd` 및 `completionZshCmd` 변수 추가 및 `init()`에 등록
- `runEdit` 함수 구현
- `runVersion` 함수 구현
- `runCompletionZsh` 함수 구현
- `$EDITOR` 환경 변수 처리 로직 추가

**주의사항**:
- `runCompletionZsh` 함수는 `manager`를 사용하지 않으므로 `main()` 함수에서 `manager` 초기화 전에 호출되어도 됨
- 하지만 completion 명령어는 독립적으로 실행 가능해야 하므로, `manager` 초기화 없이도 동작해야 함
- Cobra의 `GenZshCompletion()`은 명령어 구조만 필요하므로 `manager`가 필요 없음

## 테스트 계획

1. `edit` 명령어 테스트:
   - 스니펫 선택 및 편집
   - 에디터 실행 확인
   - 저장 후 변경사항 반영 확인

2. `version` 명령어 테스트:
   - 버전 정보 출력 확인
   - 커밋 해시 및 빌드 날짜 확인

3. `completion zsh` 명령어 테스트:
   - zsh 자동완성 스크립트 생성 확인
     ```bash
     snipgo completion zsh | head -50  # 스크립트 내용 확인
     ```
   - 실제 zsh에서 자동완성 동작 확인
     ```bash
     # 설치 후 새 터미널에서
     snipgo <TAB>  # 모든 명령어 목록 표시 확인
     snipgo new <TAB>  # new 명령어 플래그 표시 확인
     snipgo config <TAB>  # config 서브커맨드 표시 확인
     snipgo completion <TAB>  # completion 서브커맨드 표시 확인
     ```
   - 모든 서브커맨드 자동완성 확인
   - 플래그 자동완성 확인 (`--help`, `--version` 등)
   - 에러 처리 확인 (manager 초기화 없이도 동작)

## 참고사항

- pet의 `edit` 명령어는 `$EDITOR`를 사용하여 임시 파일을 열고 편집하는 방식
- 기존 코드의 `serializeSnippetForEdit`, `parseSnippetFromEdit` 함수를 재사용
- `version` 명령어는 `rootCmd.Version`을 활용하거나 별도로 구현
- `completion` 명령어는 Cobra의 내장 기능을 활용하여 간단히 구현 가능
- `sync` 명령어는 pet이 Gist와 동기화하는 기능이지만, snipgo는 local-first이므로 roadmap에만 추가하고 구현하지 않음

