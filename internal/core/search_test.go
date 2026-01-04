package core

import (
	"os"
	"testing"
)

func TestManager_Search(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test config
	cleanup, err := setupTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}
	defer cleanup()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Setup test snippets
	snippets := []*Snippet{
		{
			ID:    "id-1",
			Title: "Go Programming",
			Tags:  []string{"go", "programming"},
			Body:  "This is about Go programming language",
		},
		{
			ID:    "id-2",
			Title: "Python Tutorial",
			Tags:  []string{"python", "tutorial"},
			Body:  "Learn Python programming",
		},
		{
			ID:    "id-3",
			Title: "JavaScript Basics",
			Tags:  []string{"javascript", "web"},
			Body:  "JavaScript for web development",
		},
		{
			ID:    "id-4",
			Title: "Database Design",
			Tags:  []string{"database", "sql"},
			Body:  "Designing database schemas",
		},
		{
			ID:    "id-5",
			Title: "Web Development",
			Tags:  []string{"web", "html"},
			Body:  "Building web applications with HTML and CSS",
		},
	}

	// Save all snippets
	for _, s := range snippets {
		if err := m.Save(s); err != nil {
			t.Fatalf("Failed to save snippet: %v", err)
		}
	}

	tests := []struct {
		name         string
		query        string
		wantCount    int
		wantIDs      []string // Expected snippet IDs in result (order matters for score)
		checkScores  bool
		minScore     int // Minimum score for first result
	}{
		{
			name:      "empty query returns all snippets",
			query:     "",
			wantCount: 5,
			wantIDs:   []string{"id-1", "id-2", "id-3", "id-4", "id-5"},
			checkScores: false,
		},
		{
			name:      "fuzzy search on title - exact match",
			query:     "Go Programming",
			wantCount: 1,
			wantIDs:   []string{"id-1"},
			checkScores: true,
			minScore:   1, // Fuzzy search gives positive score
		},
		{
			name:      "fuzzy search on title - partial match",
			query:     "Go",
			wantCount: 1,
			wantIDs:   []string{"id-1"},
			checkScores: true,
			minScore:   1,
		},
		{
			name:      "fuzzy search on title - case insensitive",
			query:     "python",
			wantCount: 1,
			wantIDs:   []string{"id-2"},
			checkScores: true,
			minScore:   1,
		},
		{
			name:      "fuzzy search on title - typo tolerance",
			query:     "Javascrpt", // Typo: missing 'i'
			wantCount: 1,
			wantIDs:   []string{"id-3"},
			checkScores: true,
			minScore:   1,
		},
		{
			name:      "search by tag",
			query:     "web",
			wantCount: 2, // id-3 and id-5 have "web" tag
			wantIDs:   []string{"id-3", "id-5"}, // Both should match
			checkScores: true,
			minScore:   10, // Tag match gives score 10
		},
		{
			name:      "search by body",
			query:     "database",
			wantCount: 1,
			wantIDs:   []string{"id-4"},
			checkScores: true,
			minScore:   5, // Body match gives score 5
		},
		{
			name:      "search by tag and body",
			query:     "programming",
			wantCount: 2, // id-1 (tag) and id-2 (body)
			wantIDs:   []string{"id-1", "id-2"},
			checkScores: true,
			minScore:   5,
		},
		{
			name:      "title match excludes from tag/body search",
			query:     "Python",
			wantCount: 1, // Only id-2, not id-1 (which has "programming" in body)
			wantIDs:   []string{"id-2"},
			checkScores: true,
			minScore:   1, // Title match
		},
		{
			name:      "no matches",
			query:     "nonexistent",
			wantCount: 0,
			wantIDs:   []string{},
			checkScores: false,
		},
		{
			name:      "multiple tag matches",
			query:     "web",
			wantCount: 2,
			wantIDs:   []string{"id-3", "id-5"},
			checkScores: true,
			minScore:   10,
		},
		{
			name:      "tag match with body match",
			query:     "web",
			wantCount: 2,
			wantIDs:   []string{"id-3", "id-5"},
			checkScores: true,
			minScore:   10, // Tag match (10) is higher than body match (5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := m.Search(tt.query)

			if len(results) != tt.wantCount {
				t.Errorf("Manager.Search(%q) returned %d results, want %d", tt.query, len(results), tt.wantCount)
				return
			}

			// Check that all expected IDs are present
			resultIDs := make(map[string]bool)
			for _, result := range results {
				resultIDs[result.Snippet.ID] = true
			}

			for _, wantID := range tt.wantIDs {
				if !resultIDs[wantID] {
					t.Errorf("Manager.Search(%q) missing expected ID: %s", tt.query, wantID)
				}
			}

			// Check scores
			if tt.checkScores {
				if len(results) > 0 {
					// First result should have highest score
					firstScore := results[0].Score
					if firstScore < tt.minScore {
						t.Errorf("Manager.Search(%q) first result score = %d, want >= %d", tt.query, firstScore, tt.minScore)
					}

					// Results should be sorted by score (descending)
					for i := 1; i < len(results); i++ {
						if results[i].Score > results[i-1].Score {
							t.Errorf("Manager.Search(%q) results not sorted by score: %d > %d", tt.query, results[i].Score, results[i-1].Score)
						}
					}
				}
			}

			// Verify empty query returns all with score 0
			if tt.query == "" {
				for _, result := range results {
					if result.Score != 0 {
						t.Errorf("Manager.Search(\"\") result score = %d, want 0", result.Score)
					}
				}
			}

			// Verify results are copies (not references)
			if len(results) > 0 {
				originalTags := len(results[0].Snippet.Tags)
				results[0].Snippet.Tags = append(results[0].Snippet.Tags, "new-tag")
				// Reload to check original wasn't modified
				m.LoadAll()
				reloaded, _ := m.GetByID(results[0].Snippet.ID)
				if len(reloaded.Tags) != originalTags {
					t.Error("Manager.Search() returned references, not copies")
				}
			}
		})
	}
}

func TestManager_Search_ScoreOrdering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test config
	cleanup, err := setupTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}
	defer cleanup()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create snippets with different match types
	snippets := []*Snippet{
		{
			ID:    "title-match",
			Title: "test query",
			Body:  "some content",
		},
		{
			ID:    "tag-match",
			Title: "Other Title",
			Tags:  []string{"test"},
			Body:  "some content",
		},
		{
			ID:    "body-match",
			Title: "Other Title",
			Body:  "test query content",
		},
		{
			ID:    "tag-body-match",
			Title: "Other Title",
			Tags:  []string{"test"},
			Body:  "test query content",
		},
	}

	for _, s := range snippets {
		if err := m.Save(s); err != nil {
			t.Fatalf("Failed to save snippet: %v", err)
		}
	}

	results := m.Search("test")

	// Verify ordering: title match (highest fuzzy score) > tag match (10) > body match (5)
	// Tag+body match (15) should be highest
	if len(results) < 4 {
		t.Fatalf("Manager.Search() returned %d results, want at least 4", len(results))
	}

	// Find each result
	resultMap := make(map[string]*SearchResult)
	for _, result := range results {
		resultMap[result.Snippet.ID] = result
	}

	// Tag+body match should have highest score (15)
	if tagBodyScore := resultMap["tag-body-match"].Score; tagBodyScore < 15 {
		t.Errorf("Manager.Search() tag-body-match score = %d, want >= 15", tagBodyScore)
	}

	// Tag match should have score 10
	if tagScore := resultMap["tag-match"].Score; tagScore != 10 {
		t.Errorf("Manager.Search() tag-match score = %d, want 10", tagScore)
	}

	// Body match should have score 5
	if bodyScore := resultMap["body-match"].Score; bodyScore != 5 {
		t.Errorf("Manager.Search() body-match score = %d, want 5", bodyScore)
	}

	// Title match should have positive fuzzy score
	if titleScore := resultMap["title-match"].Score; titleScore <= 0 {
		t.Errorf("Manager.Search() title-match score = %d, want > 0", titleScore)
	}
}

func TestManager_Search_TitleMatchExclusion(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test config
	cleanup, err := setupTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}
	defer cleanup()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create snippet that matches both title and tag/body
	snippet := &Snippet{
		ID:    "test-id",
		Title: "test",
		Tags:  []string{"test"},
		Body:  "test content",
	}

	if err := m.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	results := m.Search("test")

	// Should only appear once (from title match, not tag/body)
	if len(results) != 1 {
		t.Errorf("Manager.Search() returned %d results, want 1 (title match should exclude tag/body match)", len(results))
	}

	if results[0].Snippet.ID != "test-id" {
		t.Errorf("Manager.Search() returned wrong snippet ID: %s", results[0].Snippet.ID)
	}

	// Should have fuzzy score from title match, not tag/body score
	if results[0].Score <= 0 {
		t.Errorf("Manager.Search() score = %d, want > 0 (fuzzy title match)", results[0].Score)
	}
}

func TestMatchesTags(t *testing.T) {
	tests := []struct {
		name        string
		snippet     *Snippet
		filterTags  []string
		wantMatch   bool
	}{
		{
			name:        "empty filter matches all",
			snippet:     &Snippet{Tags: []string{"go", "web"}},
			filterTags:  []string{},
			wantMatch:   true,
		},
		{
			name:        "single tag match",
			snippet:     &Snippet{Tags: []string{"go", "web"}},
			filterTags:  []string{"go"},
			wantMatch:   true,
		},
		{
			name:        "multiple tags match (AND)",
			snippet:     &Snippet{Tags: []string{"go", "web", "api"}},
			filterTags:  []string{"go", "web"},
			wantMatch:   true,
		},
		{
			name:        "case insensitive match",
			snippet:     &Snippet{Tags: []string{"Go", "Web"}},
			filterTags:  []string{"go", "web"},
			wantMatch:   true,
		},
		{
			name:        "missing one tag (AND logic fails)",
			snippet:     &Snippet{Tags: []string{"go"}},
			filterTags:  []string{"go", "web"},
			wantMatch:   false,
		},
		{
			name:        "no matching tags",
			snippet:     &Snippet{Tags: []string{"python"}},
			filterTags:  []string{"go"},
			wantMatch:   false,
		},
		{
			name:        "empty snippet tags",
			snippet:     &Snippet{Tags: []string{}},
			filterTags:  []string{"go"},
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesTags(tt.snippet, tt.filterTags)
			if got != tt.wantMatch {
				t.Errorf("matchesTags() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesLanguage(t *testing.T) {
	tests := []struct {
		name       string
		snippet    *Snippet
		filterLang string
		wantMatch  bool
	}{
		{
			name:       "empty filter matches all",
			snippet:    &Snippet{Language: "go"},
			filterLang: "",
			wantMatch:  true,
		},
		{
			name:       "exact match",
			snippet:    &Snippet{Language: "go"},
			filterLang: "go",
			wantMatch:  true,
		},
		{
			name:       "case insensitive match",
			snippet:    &Snippet{Language: "Python"},
			filterLang: "python",
			wantMatch:  true,
		},
		{
			name:       "case insensitive match uppercase",
			snippet:    &Snippet{Language: "bash"},
			filterLang: "BASH",
			wantMatch:  true,
		},
		{
			name:       "no match",
			snippet:    &Snippet{Language: "go"},
			filterLang: "python",
			wantMatch:  false,
		},
		{
			name:       "empty snippet language",
			snippet:    &Snippet{Language: ""},
			filterLang: "go",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesLanguage(tt.snippet, tt.filterLang)
			if got != tt.wantMatch {
				t.Errorf("matchesLanguage() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestManager_SearchWithFilters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup test config
	cleanup, err := setupTestConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to setup test config: %v", err)
	}
	defer cleanup()

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Setup test snippets with tags and languages
	snippets := []*Snippet{
		{
			ID:       "id-1",
			Title:    "Go Web Server",
			Tags:     []string{"go", "web"},
			Language: "go",
			Body:     "Building web servers in Go",
		},
		{
			ID:       "id-2",
			Title:    "Python Script",
			Tags:     []string{"python", "automation"},
			Language: "python",
			Body:     "Automation script in Python",
		},
		{
			ID:       "id-3",
			Title:    "Bash Deploy",
			Tags:     []string{"bash", "devops", "deploy"},
			Language: "bash",
			Body:     "Deployment script using bash",
		},
		{
			ID:       "id-4",
			Title:    "Go API",
			Tags:     []string{"go", "api"},
			Language: "go",
			Body:     "REST API in Go",
		},
		{
			ID:       "id-5",
			Title:    "Docker Setup",
			Tags:     []string{"docker", "devops"},
			Language: "bash",
			Body:     "Docker configuration",
		},
	}

	// Save all snippets
	for _, s := range snippets {
		if err := m.Save(s); err != nil {
			t.Fatalf("Failed to save snippet: %v", err)
		}
	}

	tests := []struct {
		name      string
		opts      SearchOptions
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "no filters or query returns all",
			opts:      SearchOptions{},
			wantCount: 5,
		},
		{
			name:      "filter by single tag",
			opts:      SearchOptions{Tags: []string{"go"}},
			wantCount: 2,
			wantIDs:   []string{"id-1", "id-4"},
		},
		{
			name:      "filter by multiple tags (AND logic)",
			opts:      SearchOptions{Tags: []string{"go", "web"}},
			wantCount: 1,
			wantIDs:   []string{"id-1"},
		},
		{
			name:      "filter by language",
			opts:      SearchOptions{Language: "bash"},
			wantCount: 2,
			wantIDs:   []string{"id-3", "id-5"},
		},
		{
			name:      "filter by language case insensitive",
			opts:      SearchOptions{Language: "Python"},
			wantCount: 1,
			wantIDs:   []string{"id-2"},
		},
		{
			name:      "combined tag and language filters",
			opts:      SearchOptions{Tags: []string{"devops"}, Language: "bash"},
			wantCount: 2,
			wantIDs:   []string{"id-3", "id-5"},
		},
		{
			name:      "query only (no filters)",
			opts:      SearchOptions{Query: "Go"},
			wantCount: 2,
			wantIDs:   []string{"id-1", "id-4"},
		},
		{
			name:      "query with tag filter",
			opts:      SearchOptions{Query: "API", Tags: []string{"go"}},
			wantCount: 1,
			wantIDs:   []string{"id-4"},
		},
		{
			name: "query with language filter",
			opts: SearchOptions{Query: "deploy", Language: "bash"},
			wantCount: 1,
			wantIDs:   []string{"id-3"},
		},
		{
			name: "all filters combined",
			opts: SearchOptions{
				Query:    "deploy",
				Tags:     []string{"devops"},
				Language: "bash",
			},
			wantCount: 1,
			wantIDs:   []string{"id-3"},
		},
		{
			name:      "no matches - tag filter",
			opts:      SearchOptions{Tags: []string{"nonexistent"}},
			wantCount: 0,
		},
		{
			name:      "no matches - language filter",
			opts:      SearchOptions{Language: "ruby"},
			wantCount: 0,
		},
		{
			name:      "no matches - filters with query",
			opts:      SearchOptions{Query: "test", Tags: []string{"go"}, Language: "python"},
			wantCount: 0,
		},
		{
			name:      "tag filter case insensitive",
			opts:      SearchOptions{Tags: []string{"Go", "Web"}},
			wantCount: 1,
			wantIDs:   []string{"id-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := m.SearchWithFilters(tt.opts)

			if len(results) != tt.wantCount {
				t.Errorf("SearchWithFilters() returned %d results, want %d", len(results), tt.wantCount)
				return
			}

			// Check that all expected IDs are present
			if len(tt.wantIDs) > 0 {
				resultIDs := make(map[string]bool)
				for _, result := range results {
					resultIDs[result.Snippet.ID] = true
				}

				for _, wantID := range tt.wantIDs {
					if !resultIDs[wantID] {
						t.Errorf("SearchWithFilters() missing expected ID: %s", wantID)
					}
				}
			}

			// Verify empty query with filters returns score 0
			if tt.opts.Query == "" && len(results) > 0 {
				for _, result := range results {
					if result.Score != 0 {
						t.Errorf("SearchWithFilters() with empty query, result score = %d, want 0", result.Score)
					}
				}
			}

			// Verify results are copies (not references)
			if len(results) > 0 {
				originalTags := len(results[0].Snippet.Tags)
				results[0].Snippet.Tags = append(results[0].Snippet.Tags, "new-tag")
				// Reload to check original wasn't modified
				m.LoadAll()
				reloaded, _ := m.GetByID(results[0].Snippet.ID)
				if len(reloaded.Tags) != originalTags {
					t.Error("SearchWithFilters() returned references, not copies")
				}
			}
		})
	}
}

