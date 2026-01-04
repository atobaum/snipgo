# Search Filter Implementation Summary

## 작업 개요

`snipgo search` 명령어에 tag와 language 필터링 기능을 추가하고, query를 positional argument에서 flag 기반으로 변경했습니다.

## 구현된 기능

### 1. CLI 인터페이스 변경

**이전:**
```bash
snipgo search [query]
```

**이후:**
```bash
snipgo search -q "query"              # Query는 이제 -q 플래그로
snipgo search --tag golang            # 단일 태그 필터
snipgo search --tag golang --tag web  # 여러 태그 필터 (AND 로직)
snipgo search --lang bash             # 언어 필터
snipgo search --tag devops --lang bash -q "deploy"  # 모든 필터 조합
```

### 2. 주요 플래그

- `-q, --query`: 검색 쿼리 (fuzzy match)
- `-t, --tag`: 태그 필터 (반복 가능, AND 로직)
- `-L, --language`: 언어 필터
- `--lang`: --language의 별칭

### 3. 필터 로직

**적용 순서:**
1. 태그 + 언어 필터 적용 → 후보 snippet set
2. 후보 set에서 fuzzy search 수행
3. 점수순 정렬

**필터 의미:**
- **Tags**: AND 로직 (모든 태그를 포함해야 함), 대소문자 구분 없음
- **Language**: 정확한 매칭, 대소문자 구분 없음
- **Query**: Fuzzy title match + substring tag/body match (기존 로직)

## 수정된 파일

### 1. `internal/core/search.go`

**추가된 구조체:**
```go
type SearchOptions struct {
    Query    string   // Fuzzy search query (can be empty)
    Tags     []string // Filter by tags (AND logic, case-insensitive)
    Language string   // Filter by language (case-insensitive, empty means no filter)
}
```

**추가된 함수:**
- `matchesTags(snippet *Snippet, filterTags []string) bool` - 태그 필터링 헬퍼
- `matchesLanguage(snippet *Snippet, filterLang string) bool` - 언어 필터링 헬퍼
- `SearchWithFilters(opts SearchOptions) []*SearchResult` - 필터 지원 검색

**리팩토링:**
- `Search(query string)` → `SearchWithFilters`를 호출하는 래퍼로 변경 (역호환성 유지)

### 2. `cmd/snipgo/search.go`

**변경사항:**
- Command 정의에서 `Args: cobra.MaximumNArgs(1)` 제거
- 플래그 추가: `--query`, `--tag`, `--language`, `--lang`
- `runSearch` 함수 완전 재작성:
  - 플래그 파싱
  - SearchOptions 구조체 생성
  - SearchWithFilters 호출
  - 매칭되지 않을 경우 상세한 에러 메시지 출력

### 3. `internal/core/search_test.go`

**추가된 테스트:**
- `TestMatchesTags` - matchesTags 헬퍼 함수 테스트 (7개 케이스)
- `TestMatchesLanguage` - matchesLanguage 헬퍼 함수 테스트 (6개 케이스)
- `TestManager_SearchWithFilters` - SearchWithFilters 메서드 테스트 (14개 케이스)

**테스트 커버리지:**
- 단일 태그 필터
- 여러 태그 필터 (AND 로직)
- 언어 필터 (대소문자 구분 없음)
- 태그 + 언어 조합 필터
- 필터 + 쿼리 조합
- 필터만 사용 (쿼리 없음)
- 매칭 없음 케이스
- 대소문자 구분 없는 매칭

### 4. `CLAUDE.md`

**업데이트된 섹션:**
CLI Usage 섹션에 새로운 search 명령어 사용 예제 추가

## 테스트 결과

**모든 core 테스트 통과:**
- ✅ TestMatchesTags (7/7 케이스 통과)
- ✅ TestMatchesLanguage (6/6 케이스 통과)
- ✅ TestManager_SearchWithFilters (14/14 케이스 통과)
- ✅ 기존 Search 테스트 모두 통과 (역호환성 확인)

**빌드 성공:**
- ✅ CLI 바이너리 빌드 완료
- ✅ Help 메시지에 새로운 플래그 정상 표시

## 역호환성

- **GUI**: 변경 없음 - 기존 `Search(query string)` 메서드 계속 사용 가능
- **다른 CLI 명령어**: `copy`, `exec` 명령어 영향 없음
- **기존 Search 메서드**: 래퍼로 유지되어 기존 호출 코드 모두 정상 작동

## 사용 예시

```bash
# 1. 태그로 필터링
./bin/snipgo search --tag golang

# 2. 여러 태그로 필터링 (AND)
./bin/snipgo search --tag golang --tag web

# 3. 언어로 필터링
./bin/snipgo search --lang bash

# 4. 태그 + 언어 필터
./bin/snipgo search --tag devops --lang bash

# 5. 필터 + 쿼리
./bin/snipgo search --tag devops --lang bash -q "deploy"

# 6. 쿼리만 사용
./bin/snipgo search -q "docker"

# 7. 필터/쿼리 없이 (모든 snippet 표시)
./bin/snipgo search
```

## 구현 특징

1. **필터 우선 적용**: 필터를 먼저 적용해서 검색 대상을 줄인 후 fuzzy search 수행 (성능 향상)
2. **AND 로직**: 여러 태그 필터는 AND 로직으로 동작 (점진적 필터링에 유용)
3. **대소문자 구분 없음**: 모든 필터는 대소문자를 구분하지 않음
4. **확장 가능한 설계**: SearchOptions 구조체로 향후 추가 필터 확장 용이
5. **상세한 에러 메시지**: 매칭 실패 시 어떤 필터가 적용되었는지 표시

## 다음 단계

1. PR 생성을 위한 브랜치 생성
2. 변경사항 커밋
3. PR 생성

---

**작업 완료 일시:** 2026-01-04
**구현 파일:**
- internal/core/search.go
- cmd/snipgo/search.go
- internal/core/search_test.go
- CLAUDE.md
