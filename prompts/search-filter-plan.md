# Search Filter 기능 구현 계획

## 프로젝트 목표

`snipgo search` 명령어에 tag와 language 필터링 기능 추가

## 요구사항

1. Tag 필터 (--tag, 반복 가능, AND 로직)
2. Language 필터 (--language, --lang)
3. Query를 positional argument → flag로 변경 (-q, --query)
4. 필터만 사용 가능 (query 없이도 동작)

## 설계 결정

### CLI 인터페이스

```bash
snipgo search -q "query"
snipgo search --tag golang --tag web  # AND 로직
snipgo search --lang bash
snipgo search --tag devops --lang bash -q "deploy"
```

### 필터 적용 순서

1. **필터 먼저**: Tag + Language 필터로 후보 snippet 추출
2. **검색 실행**: 후보 snippet에서 fuzzy search
3. **정렬**: 점수순 정렬

### 백엔드 구조

**SearchOptions 구조체:**
```go
type SearchOptions struct {
    Query    string   // 검색 쿼리 (선택)
    Tags     []string // 태그 필터 (AND 로직)
    Language string   // 언어 필터
}
```

**새 메서드:**
- `SearchWithFilters(opts SearchOptions) []*SearchResult`

**리팩토링:**
- 기존 `Search(query string)` → `SearchWithFilters` 래퍼

## 구현 단계

### 1단계: Core Logic - SearchOptions 구조체
**파일:** `internal/core/search.go`
- SearchOptions 구조체 정의

### 2단계: Core Logic - 필터 헬퍼 함수
**파일:** `internal/core/search.go`

**matchesTags:**
- 빈 필터 → true
- 대소문자 구분 없이 모든 태그 포함 확인 (AND)

**matchesLanguage:**
- 빈 필터 → true
- 대소문자 구분 없이 정확히 매칭

### 3단계: Core Logic - SearchWithFilters 메서드
**파일:** `internal/core/search.go`

1. 필터 적용 → 후보 snippet set
2. Query 없으면 모든 후보 반환 (score: 0)
3. Query 있으면 fuzzy search 수행
4. 점수순 정렬

### 4단계: Core Logic - Search 메서드 리팩토링
**파일:** `internal/core/search.go`

기존 Search를 SearchWithFilters 래퍼로 변경:
```go
func (m *Manager) Search(query string) []*SearchResult {
    return m.SearchWithFilters(SearchOptions{Query: query})
}
```

### 5단계: CLI - 플래그 추가
**파일:** `cmd/snipgo/search.go`

- `Args: cobra.MaximumNArgs(1)` 제거
- 플래그 추가:
  ```go
  searchCmd.Flags().StringP("query", "q", "", "Search query (fuzzy match)")
  searchCmd.Flags().StringSliceP("tag", "t", []string{}, "Filter by tags (repeatable, AND logic)")
  searchCmd.Flags().StringP("language", "L", "", "Filter by language")
  searchCmd.Flags().StringP("lang", "", "", "Alias for --language")
  ```

### 6단계: CLI - runSearch 함수 업데이트
**파일:** `cmd/snipgo/search.go`

1. 플래그 파싱
2. language 별칭 처리
3. SearchOptions 구조체 생성
4. SearchWithFilters 호출
5. 매칭 없으면 상세 에러 메시지

### 7단계: CLI - Help 텍스트 업데이트
**파일:** `cmd/snipgo/search.go`

Long 설명에 예제 추가

### 8단계: 테스트 작성
**파일:** `internal/core/search_test.go`

**TestMatchesTags:**
- 빈 필터, 단일 태그, 여러 태그 AND, 대소문자, 실패 케이스

**TestMatchesLanguage:**
- 빈 필터, 정확 매칭, 대소문자, 실패 케이스

**TestManager_SearchWithFilters:**
- 필터 조합 (tag, language, query)
- AND 로직 확인
- 대소문자 구분 없음 확인
- 매칭 없음 케이스

### 9단계: 문서 업데이트
**파일:** `CLAUDE.md`

CLI Usage 섹션에 새로운 예제 추가

## 중요 파일

1. **internal/core/search.go** - 핵심 검색 로직
2. **cmd/snipgo/search.go** - CLI 인터페이스
3. **internal/core/search_test.go** - 테스트
4. **CLAUDE.md** - 사용자 문서

## 역호환성

- GUI: `Search(query)` 메서드 계속 사용 가능
- 다른 CLI 명령어: 영향 없음
- 기존 테스트: 모두 통과해야 함

## 구현 노트

**필터 순서의 중요성:**
필터를 먼저 적용하면 성능 향상 + 점수 로직이 깔끔함

**AND 로직 선택 이유:**
점진적 필터링에 유용 (--tag go → --tag go --tag web)

**Query를 플래그로 변경:**
모든 검색 파라미터가 플래그로 통일되어 일관성 향상
