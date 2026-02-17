package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"snipgo/internal/core"
	"snipgo/internal/history"
	"snipgo/internal/tmpl"
)

func TestExpandSnippetBody_WithAllFlagsProvided(t *testing.T) {
	snippet := &core.Snippet{
		ID:        "test1",
		Title:     "Test Snippet",
		Body:      "Hello ${A} and ${B}",
		Variables: map[string]*tmpl.Variable{},
	}

	providedVars := map[string]string{
		"A": "World",
		"B": "Universe",
	}

	result, err := expandSnippetBody(snippet, providedVars, false, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello World and Universe"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandSnippetBody_WithDefaults(t *testing.T) {
	snippet := &core.Snippet{
		ID:    "test2",
		Title: "Test Snippet with Defaults",
		Body:  "Environment: ${ENV}, Region: ${REGION}",
		Variables: map[string]*tmpl.Variable{
			"ENV": {
				Name:    "ENV",
				Default: "production",
			},
			"REGION": {
				Name:    "REGION",
				Default: "us-west-2",
			},
		},
	}

	// No provided vars - should use defaults
	result, err := expandSnippetBody(snippet, map[string]string{}, false, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Environment: production, Region: us-west-2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandSnippetBody_WithRawFlag(t *testing.T) {
	snippet := &core.Snippet{
		ID:    "test3",
		Title: "Test Raw",
		Body:  "Hello ${NAME}",
		Variables: map[string]*tmpl.Variable{
			"NAME": {Name: "NAME", Default: "World"},
		},
	}

	// Raw flag should return body unchanged
	result, err := expandSnippetBody(snippet, map[string]string{}, true, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello ${NAME}"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandSnippetBody_WithMissingValues_NonInteractive(t *testing.T) {
	snippet := &core.Snippet{
		ID:        "test4",
		Title:     "Test Missing",
		Body:      "Hello ${NAME}",
		Variables: map[string]*tmpl.Variable{},
	}

	// No provided vars, non-interactive - should error
	_, err := expandSnippetBody(snippet, map[string]string{}, false, false, nil)
	if err == nil {
		t.Fatal("expected error for missing values, got nil")
	}

	expectedMsg := "missing values for variables: NAME"
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestMergeVariables_ExcludesFrontmatterOnlyVars(t *testing.T) {
	bodyVarNames := []string{"A", "B"}
	frontmatterVars := map[string]*tmpl.Variable{
		"A": {Name: "A", Default: "alpha"},
		"B": {Name: "B", Default: "beta"},
		"C": {Name: "C", Default: "gamma"}, // Frontmatter-only, should be excluded
	}

	result := mergeVariables(bodyVarNames, frontmatterVars)

	// Should only include A and B
	if len(result) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(result))
	}

	if result[0].Name != "A" {
		t.Errorf("expected first variable to be A, got %s", result[0].Name)
	}
	if result[1].Name != "B" {
		t.Errorf("expected second variable to be B, got %s", result[1].Name)
	}

	// Verify metadata is preserved
	if result[0].Default != "alpha" {
		t.Errorf("expected A default to be 'alpha', got %q", result[0].Default)
	}
	if result[1].Default != "beta" {
		t.Errorf("expected B default to be 'beta', got %q", result[1].Default)
	}
}

func TestParseVarFlags_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		flags       []string
		want        map[string]string
		expectError bool
	}{
		{
			name:        "empty flags",
			flags:       []string{},
			want:        map[string]string{},
			expectError: false,
		},
		{
			name:        "single flag",
			flags:       []string{"KEY=VALUE"},
			want:        map[string]string{"KEY": "VALUE"},
			expectError: false,
		},
		{
			name:        "multiple flags",
			flags:       []string{"A=1", "B=2", "C=3"},
			want:        map[string]string{"A": "1", "B": "2", "C": "3"},
			expectError: false,
		},
		{
			name:        "value with equals",
			flags:       []string{"URL=https://example.com?a=b"},
			want:        map[string]string{"URL": "https://example.com?a=b"},
			expectError: false,
		},
		{
			name:        "empty value",
			flags:       []string{"EMPTY="},
			want:        map[string]string{"EMPTY": ""},
			expectError: false,
		},
		{
			name:        "missing equals",
			flags:       []string{"INVALID"},
			want:        nil,
			expectError: true,
		},
		{
			name:        "equals at start",
			flags:       []string{"=VALUE"},
			want:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVarFlags(tt.flags)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d entries, got %d", len(tt.want), len(got))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("expected %s=%q, got %q", k, v, got[k])
				}
			}
		})
	}
}

func TestHistoryRoundTrip(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "var_history.json")

	// Create history
	hist, err := history.NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	// Add some values
	if err := hist.Add("ENV", "production"); err != nil {
		t.Fatalf("failed to add ENV: %v", err)
	}
	if err := hist.Add("REGION", "us-west-2"); err != nil {
		t.Fatalf("failed to add REGION: %v", err)
	}
	if err := hist.Add("ENV", "staging"); err != nil {
		t.Fatalf("failed to add ENV again: %v", err)
	}

	// Verify values
	envValues := hist.Get("ENV")
	if len(envValues) != 2 {
		t.Fatalf("expected 2 ENV values, got %d", len(envValues))
	}
	if envValues[0] != "staging" {
		t.Errorf("expected most recent ENV to be 'staging', got %q", envValues[0])
	}
	if envValues[1] != "production" {
		t.Errorf("expected second ENV to be 'production', got %q", envValues[1])
	}

	regionValues := hist.Get("REGION")
	if len(regionValues) != 1 {
		t.Fatalf("expected 1 REGION value, got %d", len(regionValues))
	}
	if regionValues[0] != "us-west-2" {
		t.Errorf("expected REGION to be 'us-west-2', got %q", regionValues[0])
	}

	// Reload from disk
	hist2, err := history.NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("failed to reload history: %v", err)
	}

	// Verify persistence
	envValues2 := hist2.Get("ENV")
	if len(envValues2) != 2 {
		t.Fatalf("expected 2 ENV values after reload, got %d", len(envValues2))
	}
	if envValues2[0] != "staging" {
		t.Errorf("expected most recent ENV after reload to be 'staging', got %q", envValues2[0])
	}

	regionValues2 := hist2.Get("REGION")
	if len(regionValues2) != 1 {
		t.Fatalf("expected 1 REGION value after reload, got %d", len(regionValues2))
	}
	if regionValues2[0] != "us-west-2" {
		t.Errorf("expected REGION after reload to be 'us-west-2', got %q", regionValues2[0])
	}
}

func TestExpandSnippetBody_WithHistory(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "var_history.json")

	hist, err := history.NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	// Pre-populate history
	if err := hist.Add("ENV", "production"); err != nil {
		t.Fatalf("failed to add to history: %v", err)
	}

	snippet := &core.Snippet{
		ID:        "test5",
		Title:     "Test with History",
		Body:      "Environment: ${ENV}",
		Variables: map[string]*tmpl.Variable{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Provide explicit value (should override history)
	providedVars := map[string]string{"ENV": "staging"}
	result, err := expandSnippetBody(snippet, providedVars, false, false, hist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Environment: staging"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	// Verify staging was saved to history
	envValues := hist.Get("ENV")
	if len(envValues) == 0 {
		t.Fatal("expected ENV in history")
	}
	if envValues[0] != "staging" {
		t.Errorf("expected most recent ENV to be 'staging', got %q", envValues[0])
	}
}

func TestExpandSnippetBody_NilHistoryGracefulDegradation(t *testing.T) {
	snippet := &core.Snippet{
		ID:    "test6",
		Title: "Test Nil History",
		Body:  "Hello ${NAME}",
		Variables: map[string]*tmpl.Variable{
			"NAME": {Name: "NAME", Default: "World"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Nil history should work fine
	result, err := expandSnippetBody(snippet, map[string]string{}, false, false, nil)
	if err != nil {
		t.Fatalf("unexpected error with nil history: %v", err)
	}

	expected := "Hello World"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandSnippetBody_HistoryWithFrontmatterDefault(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "var_history.json")

	hist, err := history.NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	// Pre-populate history with different value
	if err := hist.Add("ENV", "staging"); err != nil {
		t.Fatalf("failed to add to history: %v", err)
	}

	snippet := &core.Snippet{
		ID:    "test7",
		Title: "Test Priority",
		Body:  "Environment: ${ENV}",
		Variables: map[string]*tmpl.Variable{
			"ENV": {
				Name:    "ENV",
				Default: "production", // Frontmatter default should take priority over history
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// No provided vars - should use frontmatter default (not history)
	result, err := expandSnippetBody(snippet, map[string]string{}, false, false, hist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Environment: production"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExpandSnippetBody_SavesAllValuesToHistory(t *testing.T) {
	tmpDir := t.TempDir()
	histPath := filepath.Join(tmpDir, "var_history.json")

	hist, err := history.NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("failed to create history: %v", err)
	}

	snippet := &core.Snippet{
		ID:    "test8",
		Title: "Test Multiple Variables",
		Body:  "${A} ${B} ${C}",
		Variables: map[string]*tmpl.Variable{
			"B": {Name: "B", Default: "beta"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	providedVars := map[string]string{
		"A": "alpha",
		"C": "gamma",
	}

	_, err = expandSnippetBody(snippet, providedVars, false, false, hist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all values saved to history
	aValues := hist.Get("A")
	if len(aValues) != 1 || aValues[0] != "alpha" {
		t.Errorf("expected A='alpha' in history, got %v", aValues)
	}

	bValues := hist.Get("B")
	if len(bValues) != 1 || bValues[0] != "beta" {
		t.Errorf("expected B='beta' in history, got %v", bValues)
	}

	cValues := hist.Get("C")
	if len(cValues) != 1 || cValues[0] != "gamma" {
		t.Errorf("expected C='gamma' in history, got %v", cValues)
	}
}

func TestExpandSnippetBody_NoVariables(t *testing.T) {
	snippet := &core.Snippet{
		ID:        "test9",
		Title:     "No Variables",
		Body:      "Hello World",
		Variables: map[string]*tmpl.Variable{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := expandSnippetBody(snippet, map[string]string{}, false, false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != snippet.Body {
		t.Errorf("expected body unchanged, got %q", result)
	}
}

// TestMain to ensure clean test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}
