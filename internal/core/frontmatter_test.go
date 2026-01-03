package core

import (
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
			name:    "no frontmatter delimiter",
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
			name:    "frontmatter not closed",
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

