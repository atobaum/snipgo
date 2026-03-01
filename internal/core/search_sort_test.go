package core

import (
	"testing"
	"time"
)

func makeSearchResult(id, title string, score int) *SearchResult {
	return &SearchResult{
		Snippet: &Snippet{ID: id, Title: title},
		Score:   score,
	}
}

func titlesOf(results []*SearchResult) []string {
	out := make([]string, len(results))
	for i, r := range results {
		out[i] = r.Snippet.Title
	}
	return out
}

func TestApplySort_FrequencyDesc(t *testing.T) {
	now := time.Now()
	results := []*SearchResult{
		makeSearchResult("id-zebra", "Zebra", 0),
		makeSearchResult("id-apple", "Apple", 0),
		makeSearchResult("id-mango", "Mango", 0),
	}
	opts := SearchOptions{
		SortBy:    "frequency",
		SortOrder: "desc",
		UsageData: map[string]UsageData{
			"id-mango": {Count: 5, LastUsed: now},
			"id-apple": {Count: 2, LastUsed: now},
			"id-zebra": {Count: 0},
		},
	}
	applySort(results, opts, false)
	got := titlesOf(results)
	want := []string{"Mango", "Apple", "Zebra"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("frequency desc: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestApplySort_FrequencyAsc(t *testing.T) {
	now := time.Now()
	results := []*SearchResult{
		makeSearchResult("id-mango", "Mango", 0),
		makeSearchResult("id-apple", "Apple", 0),
		makeSearchResult("id-zebra", "Zebra", 0),
	}
	opts := SearchOptions{
		SortBy:    "frequency",
		SortOrder: "asc",
		UsageData: map[string]UsageData{
			"id-mango": {Count: 5, LastUsed: now},
			"id-apple": {Count: 2, LastUsed: now},
			"id-zebra": {Count: 0},
		},
	}
	applySort(results, opts, false)
	got := titlesOf(results)
	want := []string{"Zebra", "Apple", "Mango"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("frequency asc: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestApplySort_NameAsc(t *testing.T) {
	results := []*SearchResult{
		makeSearchResult("1", "Zebra", 0),
		makeSearchResult("2", "Apple", 0),
		makeSearchResult("3", "Mango", 0),
	}
	opts := SearchOptions{SortBy: "name", SortOrder: "asc"}
	applySort(results, opts, false)
	got := titlesOf(results)
	want := []string{"Apple", "Mango", "Zebra"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("name asc: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestApplySort_NameDesc(t *testing.T) {
	results := []*SearchResult{
		makeSearchResult("1", "Apple", 0),
		makeSearchResult("2", "Mango", 0),
		makeSearchResult("3", "Zebra", 0),
	}
	opts := SearchOptions{SortBy: "name", SortOrder: "desc"}
	applySort(results, opts, false)
	got := titlesOf(results)
	want := []string{"Zebra", "Mango", "Apple"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("name desc: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestApplySort_LastUsedDesc(t *testing.T) {
	now := time.Now()
	results := []*SearchResult{
		makeSearchResult("id-zebra", "Zebra", 0),
		makeSearchResult("id-apple", "Apple", 0),
		makeSearchResult("id-mango", "Mango", 0),
	}
	opts := SearchOptions{
		SortBy:    "last_used",
		SortOrder: "desc",
		UsageData: map[string]UsageData{
			"id-mango": {Count: 1, LastUsed: now.Add(-1 * time.Minute)},
			"id-apple": {Count: 1, LastUsed: now.Add(-2 * time.Hour)},
			"id-zebra": {Count: 0, LastUsed: time.Time{}},
		},
	}
	applySort(results, opts, false)
	got := titlesOf(results)
	want := []string{"Mango", "Apple", "Zebra"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("last_used desc: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestApplySort_WithQuery_ScoreFirst(t *testing.T) {
	// When query present, score is primary; frequency is tiebreaker
	now := time.Now()
	results := []*SearchResult{
		{Snippet: &Snippet{ID: "id-a", Title: "Alpha"}, Score: 10},
		{Snippet: &Snippet{ID: "id-b", Title: "Beta"}, Score: 10},
		{Snippet: &Snippet{ID: "id-c", Title: "Gamma"}, Score: 50},
	}
	opts := SearchOptions{
		SortBy: "frequency",
		UsageData: map[string]UsageData{
			"id-a": {Count: 100, LastUsed: now},
			"id-b": {Count: 5, LastUsed: now},
			"id-c": {Count: 1, LastUsed: now},
		},
	}
	applySort(results, opts, true)
	got := titlesOf(results)
	// Gamma(score=50) first, then Alpha(score=10,count=100), then Beta(score=10,count=5)
	want := []string{"Gamma", "Alpha", "Beta"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("query+score: position %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestResolveOrder_Defaults(t *testing.T) {
	cases := []struct {
		sortBy    string
		sortOrder string
		want      string
	}{
		{"frequency", "", "desc"},
		{"last_used", "", "desc"},
		{"name", "", "asc"},
		{"frequency", "asc", "asc"},
		{"name", "desc", "desc"},
	}
	for _, c := range cases {
		got := resolveOrder(c.sortBy, c.sortOrder)
		if got != c.want {
			t.Errorf("resolveOrder(%q, %q) = %q, want %q", c.sortBy, c.sortOrder, got, c.want)
		}
	}
}
