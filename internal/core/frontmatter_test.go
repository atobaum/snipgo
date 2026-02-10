package core

import (
	"snipgo/internal/tmpl"
	"testing"
	"time"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    *Snippet
		wantErr bool
	}{
		{
			name: "valid frontmatter with body",
			content: []byte(`---
id: test-id
title: Test Title
tags:
  - tag1
  - tag2
language: go
is_favorite: true
created_at: 2020-01-01T00:00:00Z
updated_at: 2020-01-02T00:00:00Z
---
This is the body content.`),
			want: &Snippet{
				ID:         "test-id",
				Title:      "Test Title",
				Tags:       []string{"tag1", "tag2"},
				Language:   "go",
				IsFavorite: true,
				Body:       "This is the body content.",
			},
			wantErr: false,
		},
		{
			name: "valid frontmatter with description",
			content: []byte(`---
id: test-id
title: Test Title
description: A short description
tags:
  - tag1
language: go
---
Body content.`),
			want: &Snippet{
				ID:          "test-id",
				Title:       "Test Title",
				Description: "A short description",
				Tags:        []string{"tag1"},
				Language:    "go",
				Body:        "Body content.",
			},
			wantErr: false,
		},
		{
			name: "valid frontmatter with empty body",
			content: []byte(`---
id: test-id
title: Test Title
---
`),
			want: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "",
			},
			wantErr: false,
		},
		{
			name: "valid frontmatter with multiline body",
			content: []byte(`---
id: test-id
title: Test Title
---
Line 1
Line 2
Line 3`),
			want: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "Line 1\nLine 2\nLine 3",
			},
			wantErr: false,
		},
		{
			name: "valid frontmatter with body starting with newline",
			content: []byte(`---
id: test-id
title: Test Title
---

Body content`),
			want: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "Body content",
			},
			wantErr: false,
		},
		{
			name: "no frontmatter delimiter",
			content: []byte(`id: test-id
title: Test Title`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty content",
			content: []byte(``),
			want:    nil,
			wantErr: true,
		},
		{
			name: "frontmatter not closed",
			content: []byte(`---
id: test-id
title: Test Title`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid YAML",
			content: []byte(`---
id: test-id
title: Test Title
invalid: [unclosed
---
Body`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "frontmatter with only delimiter",
			content: []byte(`---
---
Body`),
			want: &Snippet{
				Body: "Body",
			},
			wantErr: false,
		},
		{
			name: "frontmatter with empty lines",
			content: []byte(`---
id: test-id
title: Test Title

---
Body`),
			want: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "Body",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFrontmatter(tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("ParseFrontmatter() got = %v, want nil on error", got)
				}
				return
			}

			if got == nil {
				t.Error("ParseFrontmatter() got = nil, want non-nil")
				return
			}

			// Compare fields
			if got.ID != tt.want.ID {
				t.Errorf("ParseFrontmatter() ID = %v, want %v", got.ID, tt.want.ID)
			}

			if got.Title != tt.want.Title {
				t.Errorf("ParseFrontmatter() Title = %v, want %v", got.Title, tt.want.Title)
			}

			if got.Description != tt.want.Description {
				t.Errorf("ParseFrontmatter() Description = %v, want %v", got.Description, tt.want.Description)
			}

			if len(got.Tags) != len(tt.want.Tags) {
				t.Errorf("ParseFrontmatter() Tags length = %v, want %v", len(got.Tags), len(tt.want.Tags))
			} else {
				for i, tag := range got.Tags {
					if tag != tt.want.Tags[i] {
						t.Errorf("ParseFrontmatter() Tags[%d] = %v, want %v", i, tag, tt.want.Tags[i])
					}
				}
			}

			if got.Language != tt.want.Language {
				t.Errorf("ParseFrontmatter() Language = %v, want %v", got.Language, tt.want.Language)
			}

			if got.IsFavorite != tt.want.IsFavorite {
				t.Errorf("ParseFrontmatter() IsFavorite = %v, want %v", got.IsFavorite, tt.want.IsFavorite)
			}

			if got.Body != tt.want.Body {
				t.Errorf("ParseFrontmatter() Body = %v, want %v", got.Body, tt.want.Body)
			}
		})
	}
}

func TestSerializeFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		snippet *Snippet
		wantErr bool
		check   func(t *testing.T, content []byte)
	}{
		{
			name: "valid snippet",
			snippet: &Snippet{
				ID:         "test-id",
				Title:      "Test Title",
				Tags:       []string{"tag1", "tag2"},
				Language:   "go",
				IsFavorite: true,
				CreatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				Body:       "Body content",
			},
			wantErr: false,
			check: func(t *testing.T, content []byte) {
				contentStr := string(content)
				if len(contentStr) == 0 {
					t.Error("SerializeFrontmatter() content is empty")
				}
				// Check frontmatter delimiter
				if contentStr[:3] != "---" {
					t.Error("SerializeFrontmatter() content does not start with ---")
				}
				// Check body is included
				if !contains(contentStr, "Body content") {
					t.Error("SerializeFrontmatter() body not found in content")
				}
			},
		},
		{
			name: "snippet with empty body",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "",
			},
			wantErr: false,
			check: func(t *testing.T, content []byte) {
				contentStr := string(content)
				if len(contentStr) == 0 {
					t.Error("SerializeFrontmatter() content is empty")
				}
			},
		},
		{
			name: "snippet with multiline body",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
				Body:  "Line 1\nLine 2\nLine 3",
			},
			wantErr: false,
			check: func(t *testing.T, content []byte) {
				contentStr := string(content)
				if !contains(contentStr, "Line 1") {
					t.Error("SerializeFrontmatter() multiline body not found")
				}
				if !contains(contentStr, "Line 2") {
					t.Error("SerializeFrontmatter() multiline body not found")
				}
			},
		},
		{
			name: "invalid snippet - empty ID",
			snippet: &Snippet{
				ID:    "",
				Title: "Test Title",
			},
			wantErr: true,
		},
		{
			name: "invalid snippet - empty Title",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SerializeFrontmatter(tt.snippet)

			if (err != nil) != tt.wantErr {
				t.Errorf("SerializeFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("SerializeFrontmatter() got = %v, want nil on error", got)
				}
				return
			}

			if got == nil {
				t.Error("SerializeFrontmatter() got = nil, want non-nil")
				return
			}

			if tt.check != nil {
				tt.check(t, got)
			}

			// Verify we can parse it back
			parsed, err := ParseFrontmatter(got)
			if err != nil {
				t.Errorf("SerializeFrontmatter() generated content cannot be parsed: %v", err)
				return
			}

			// Compare key fields
			if parsed.ID != tt.snippet.ID {
				t.Errorf("SerializeFrontmatter() round-trip ID = %v, want %v", parsed.ID, tt.snippet.ID)
			}

			if parsed.Title != tt.snippet.Title {
				t.Errorf("SerializeFrontmatter() round-trip Title = %v, want %v", parsed.Title, tt.snippet.Title)
			}

			if parsed.Body != tt.snippet.Body {
				t.Errorf("SerializeFrontmatter() round-trip Body = %v, want %v", parsed.Body, tt.snippet.Body)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseFrontmatter_WithVariables(t *testing.T) {
	content := []byte(`---
id: test-var-id
title: Deploy Script
description: ""
tags: []
language: bash
is_favorite: false
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
variables:
  SERVER:
    description: Target server
    default: prod-01
  ENV:
    description: Environment
    choices:
      - staging
      - production
---

ssh ${SERVER} "deploy --env ${ENV}"`)

	snippet, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}

	// Verify Variables map is populated
	if snippet.Variables == nil {
		t.Fatal("ParseFrontmatter() Variables is nil, want non-nil map")
	}

	if len(snippet.Variables) != 2 {
		t.Errorf("ParseFrontmatter() Variables length = %v, want 2", len(snippet.Variables))
	}

	// Verify SERVER variable
	serverVar, exists := snippet.Variables["SERVER"]
	if !exists {
		t.Fatal("ParseFrontmatter() Variables[SERVER] does not exist")
	}

	if serverVar.Name != "SERVER" {
		t.Errorf("ParseFrontmatter() Variables[SERVER].Name = %v, want SERVER", serverVar.Name)
	}

	if serverVar.Description != "Target server" {
		t.Errorf("ParseFrontmatter() Variables[SERVER].Description = %v, want 'Target server'", serverVar.Description)
	}

	if serverVar.Default != "prod-01" {
		t.Errorf("ParseFrontmatter() Variables[SERVER].Default = %v, want 'prod-01'", serverVar.Default)
	}

	// Verify ENV variable
	envVar, exists := snippet.Variables["ENV"]
	if !exists {
		t.Fatal("ParseFrontmatter() Variables[ENV] does not exist")
	}

	if envVar.Name != "ENV" {
		t.Errorf("ParseFrontmatter() Variables[ENV].Name = %v, want ENV", envVar.Name)
	}

	if envVar.Description != "Environment" {
		t.Errorf("ParseFrontmatter() Variables[ENV].Description = %v, want 'Environment'", envVar.Description)
	}

	if len(envVar.Choices) != 2 {
		t.Errorf("ParseFrontmatter() Variables[ENV].Choices length = %v, want 2", len(envVar.Choices))
	} else {
		if envVar.Choices[0] != "staging" {
			t.Errorf("ParseFrontmatter() Variables[ENV].Choices[0] = %v, want 'staging'", envVar.Choices[0])
		}
		if envVar.Choices[1] != "production" {
			t.Errorf("ParseFrontmatter() Variables[ENV].Choices[1] = %v, want 'production'", envVar.Choices[1])
		}
	}

	// Verify body
	expectedBody := `ssh ${SERVER} "deploy --env ${ENV}"`
	if snippet.Body != expectedBody {
		t.Errorf("ParseFrontmatter() Body = %v, want %v", snippet.Body, expectedBody)
	}
}

func TestParseFrontmatter_WithoutVariables(t *testing.T) {
	content := []byte(`---
id: test-id
title: Simple Script
description: A simple script
tags: []
language: bash
is_favorite: false
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
---

echo "hello world"`)

	snippet, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}

	// Verify Variables is nil (backward compatibility)
	if snippet.Variables != nil {
		t.Errorf("ParseFrontmatter() Variables = %v, want nil for backward compatibility", snippet.Variables)
	}
}

func TestSerializeFrontmatter_WithVariables(t *testing.T) {
	snippet := &Snippet{
		ID:          "test-var-id",
		Title:       "Deploy Script",
		Description: "",
		Tags:        []string{},
		Language:    "bash",
		IsFavorite:  false,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Variables: map[string]*tmpl.Variable{
			"SERVER": {
				Name:        "SERVER",
				Description: "Target server",
				Default:     "prod-01",
			},
			"ENV": {
				Name:        "ENV",
				Description: "Environment",
				Choices:     []string{"staging", "production"},
			},
		},
		Body: `ssh ${SERVER} "deploy --env ${ENV}"`,
	}

	content, err := SerializeFrontmatter(snippet)
	if err != nil {
		t.Fatalf("SerializeFrontmatter() error = %v", err)
	}

	// Verify round-trip: parse -> serialize -> parse gives same result
	parsed, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("SerializeFrontmatter() round-trip parse error = %v", err)
	}

	// Verify Variables survived round-trip
	if parsed.Variables == nil {
		t.Fatal("SerializeFrontmatter() round-trip Variables is nil")
	}

	if len(parsed.Variables) != 2 {
		t.Errorf("SerializeFrontmatter() round-trip Variables length = %v, want 2", len(parsed.Variables))
	}

	// Verify SERVER variable
	serverVar, exists := parsed.Variables["SERVER"]
	if !exists {
		t.Fatal("SerializeFrontmatter() round-trip Variables[SERVER] does not exist")
	}

	if serverVar.Name != "SERVER" {
		t.Errorf("SerializeFrontmatter() round-trip Variables[SERVER].Name = %v, want SERVER", serverVar.Name)
	}

	if serverVar.Description != "Target server" {
		t.Errorf("SerializeFrontmatter() round-trip Variables[SERVER].Description = %v, want 'Target server'", serverVar.Description)
	}

	if serverVar.Default != "prod-01" {
		t.Errorf("SerializeFrontmatter() round-trip Variables[SERVER].Default = %v, want 'prod-01'", serverVar.Default)
	}

	// Verify ENV variable
	envVar, exists := parsed.Variables["ENV"]
	if !exists {
		t.Fatal("SerializeFrontmatter() round-trip Variables[ENV] does not exist")
	}

	if envVar.Name != "ENV" {
		t.Errorf("SerializeFrontmatter() round-trip Variables[ENV].Name = %v, want ENV", envVar.Name)
	}

	if len(envVar.Choices) != 2 {
		t.Errorf("SerializeFrontmatter() round-trip Variables[ENV].Choices length = %v, want 2", len(envVar.Choices))
	}
}

func TestSerializeFrontmatter_WithoutVariables(t *testing.T) {
	snippet := &Snippet{
		ID:          "test-id",
		Title:       "Simple Script",
		Description: "A simple script",
		Tags:        []string{},
		Language:    "bash",
		IsFavorite:  false,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Variables:   nil,
		Body:        `echo "hello world"`,
	}

	content, err := SerializeFrontmatter(snippet)
	if err != nil {
		t.Fatalf("SerializeFrontmatter() error = %v", err)
	}

	contentStr := string(content)

	// Verify NO "variables:" line in output (due to omitempty)
	if contains(contentStr, "variables:") {
		t.Error("SerializeFrontmatter() output contains 'variables:' when Variables is nil, want no variables field")
	}
}

func TestParseFrontmatter_VariablesWithChoices(t *testing.T) {
	content := []byte(`---
id: test-choices-id
title: Test Choices
description: ""
tags: []
language: bash
is_favorite: false
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T00:00:00Z
variables:
  OPTION:
    description: Select an option
    choices:
      - option1
      - option2
      - option3
---

echo ${OPTION}`)

	snippet, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("ParseFrontmatter() error = %v", err)
	}

	optionVar, exists := snippet.Variables["OPTION"]
	if !exists {
		t.Fatal("ParseFrontmatter() Variables[OPTION] does not exist")
	}

	if len(optionVar.Choices) != 3 {
		t.Errorf("ParseFrontmatter() Variables[OPTION].Choices length = %v, want 3", len(optionVar.Choices))
	}

	expectedChoices := []string{"option1", "option2", "option3"}
	for i, choice := range optionVar.Choices {
		if choice != expectedChoices[i] {
			t.Errorf("ParseFrontmatter() Variables[OPTION].Choices[%d] = %v, want %v", i, choice, expectedChoices[i])
		}
	}
}

func TestSerializeFrontmatter_RoundTripPreservesChoicesOrdering(t *testing.T) {
	snippet := &Snippet{
		ID:          "test-ordering-id",
		Title:       "Test Ordering",
		Description: "",
		Tags:        []string{},
		Language:    "bash",
		IsFavorite:  false,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Variables: map[string]*tmpl.Variable{
			"COLOR": {
				Name:        "COLOR",
				Description: "Select a color",
				Choices:     []string{"red", "green", "blue", "yellow"},
			},
		},
		Body: `echo ${COLOR}`,
	}

	content, err := SerializeFrontmatter(snippet)
	if err != nil {
		t.Fatalf("SerializeFrontmatter() error = %v", err)
	}

	parsed, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("SerializeFrontmatter() round-trip parse error = %v", err)
	}

	colorVar, exists := parsed.Variables["COLOR"]
	if !exists {
		t.Fatal("SerializeFrontmatter() round-trip Variables[COLOR] does not exist")
	}

	if len(colorVar.Choices) != 4 {
		t.Errorf("SerializeFrontmatter() round-trip Variables[COLOR].Choices length = %v, want 4", len(colorVar.Choices))
	}

	expectedChoices := []string{"red", "green", "blue", "yellow"}
	for i, choice := range colorVar.Choices {
		if choice != expectedChoices[i] {
			t.Errorf("SerializeFrontmatter() round-trip Variables[COLOR].Choices[%d] = %v, want %v", i, choice, expectedChoices[i])
		}
	}
}
