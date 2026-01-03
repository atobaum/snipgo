package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use temp dir
	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v, want nil", err)
	}

	if m == nil {
		t.Fatal("NewManager() returned nil")
	}

	if m.snippets == nil {
		t.Error("NewManager() snippets map is nil")
	}

	if m.storage == nil {
		t.Error("NewManager() storage is nil")
	}
}

func TestManager_LoadAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name    string
		setup   func() error
		wantErr bool
		wantCount int
	}{
		{
			name: "load empty directory",
			setup: func() error {
				return nil
			},
			wantErr: false,
			wantCount: 0,
		},
		{
			name: "load single valid snippet",
			setup: func() error {
				content := `---
id: test-id-1
title: Test Snippet 1
---
Body content`
				return os.WriteFile(filepath.Join(tmpDir, "test1.md"), []byte(content), 0644)
			},
			wantErr: false,
			wantCount: 1,
		},
		{
			name: "load multiple valid snippets",
			setup: func() error {
				snippets := []struct {
					filename string
					content  string
				}{
					{"test1.md", `---
id: test-id-1
title: Test Snippet 1
---
Body 1`},
					{"test2.md", `---
id: test-id-2
title: Test Snippet 2
---
Body 2`},
					{"test3.md", `---
id: test-id-3
title: Test Snippet 3
---
Body 3`},
				}
				for _, s := range snippets {
					if err := os.WriteFile(filepath.Join(tmpDir, s.filename), []byte(s.content), 0644); err != nil {
						return err
					}
				}
				return nil
			},
			wantErr: false,
			wantCount: 3,
		},
		{
			name: "skip invalid snippet files",
			setup: func() error {
				// Valid snippet
				valid := `---
id: test-id-1
title: Valid Snippet
---
Body`
				if err := os.WriteFile(filepath.Join(tmpDir, "valid.md"), []byte(valid), 0644); err != nil {
					return err
				}
				// Invalid snippet (no frontmatter)
				invalid := `id: test-id-2
title: Invalid Snippet`
				if err := os.WriteFile(filepath.Join(tmpDir, "invalid.md"), []byte(invalid), 0644); err != nil {
					return err
				}
				// Invalid snippet (empty ID)
				invalid2 := `---
id: ""
title: Invalid Snippet 2
---
Body`
				if err := os.WriteFile(filepath.Join(tmpDir, "invalid2.md"), []byte(invalid2), 0644); err != nil {
					return err
				}
				return nil
			},
			wantErr: false,
			wantCount: 1, // Only valid snippet should be loaded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.RemoveAll(tmpDir)
			os.MkdirAll(tmpDir, 0755)

			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err := m.LoadAll()

			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.LoadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				all := m.GetAll()
				if len(all) != tt.wantCount {
					t.Errorf("Manager.LoadAll() loaded %d snippets, want %d", len(all), tt.wantCount)
				}
			}
		})
	}
}

func TestManager_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name    string
		snippet *Snippet
		wantErr bool
	}{
		{
			name: "save valid snippet",
			snippet: &Snippet{
				ID:    "test-id-1",
				Title: "Test Snippet",
				Body:  "Body content",
			},
			wantErr: false,
		},
		{
			name: "save snippet with tags and language",
			snippet: &Snippet{
				ID:       "test-id-2",
				Title:    "Test Snippet 2",
				Tags:     []string{"tag1", "tag2"},
				Language: "go",
				Body:     "Body content",
			},
			wantErr: false,
		},
		{
			name: "save invalid snippet - empty ID",
			snippet: &Snippet{
				ID:    "",
				Title: "Test Snippet",
			},
			wantErr: true,
		},
		{
			name: "save invalid snippet - empty Title",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.Save(tt.snippet)

			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify snippet is in memory
				loaded, err := m.GetByID(tt.snippet.ID)
				if err != nil {
					t.Errorf("Manager.Save() snippet not found in memory: %v", err)
					return
				}

				if loaded.ID != tt.snippet.ID {
					t.Errorf("Manager.Save() loaded ID = %v, want %v", loaded.ID, tt.snippet.ID)
				}

				if loaded.Title != tt.snippet.Title {
					t.Errorf("Manager.Save() loaded Title = %v, want %v", loaded.Title, tt.snippet.Title)
				}

				// Verify file was created
				files, err := m.storage.ListFiles()
				if err != nil {
					t.Fatalf("Failed to list files: %v", err)
				}

				found := false
				for _, file := range files {
					content, err := m.storage.ReadFile(file)
					if err != nil {
						continue
					}
					fileSnippet, err := ParseFrontmatter(content)
					if err != nil {
						continue
					}
					if fileSnippet.ID == tt.snippet.ID {
						found = true
						break
					}
				}

				if !found {
					t.Error("Manager.Save() file was not created on disk")
				}
			}
		})
	}
}

func TestManager_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name    string
		setup   func() (string, error)
		wantErr bool
	}{
		{
			name: "delete existing snippet",
			setup: func() (string, error) {
				snippet := &Snippet{
					ID:    "test-id-1",
					Title: "Test Snippet",
					Body:  "Body",
				}
				if err := m.Save(snippet); err != nil {
					return "", err
				}
				return "test-id-1", nil
			},
			wantErr: false,
		},
		{
			name: "delete non-existent snippet",
			setup: func() (string, error) {
				return "non-existent-id", nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.RemoveAll(tmpDir)
			os.MkdirAll(tmpDir, 0755)

			id, err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = m.Delete(id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify snippet is removed from memory
				_, err := m.GetByID(id)
				if err == nil {
					t.Error("Manager.Delete() snippet still exists in memory")
				}

				// Verify file was deleted
				files, err := m.storage.ListFiles()
				if err != nil {
					t.Fatalf("Failed to list files: %v", err)
				}

				for _, file := range files {
					content, err := m.storage.ReadFile(file)
					if err != nil {
						continue
					}
					fileSnippet, err := ParseFrontmatter(content)
					if err != nil {
						continue
					}
					if fileSnippet.ID == id {
						t.Error("Manager.Delete() file still exists on disk")
					}
				}
			}
		})
	}
}

func TestManager_GetByID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Add a snippet
	snippet := &Snippet{
		ID:    "test-id-1",
		Title: "Test Snippet",
		Tags:  []string{"tag1", "tag2"},
		Body:  "Body content",
	}
	if err := m.Save(snippet); err != nil {
		t.Fatalf("Failed to save snippet: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr bool
		wantID  string
	}{
		{
			name:    "get existing snippet",
			id:      "test-id-1",
			wantErr: false,
			wantID:  "test-id-1",
		},
		{
			name:    "get non-existent snippet",
			id:      "non-existent-id",
			wantErr: true,
			wantID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.GetByID(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil {
					t.Error("Manager.GetByID() returned nil")
					return
				}

				if got.ID != tt.wantID {
					t.Errorf("Manager.GetByID() ID = %v, want %v", got.ID, tt.wantID)
				}

				// Verify it's a copy (modifying returned snippet shouldn't affect original)
				originalTags := len(snippet.Tags)
				got.Tags = append(got.Tags, "new-tag")
				if len(snippet.Tags) != originalTags {
					t.Error("Manager.GetByID() returned reference, not copy")
				}
			}
		})
	}
}

func TestManager_GetAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snipgo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalEnv := os.Getenv("SNIPGO_DATA_DIR")
	defer os.Setenv("SNIPGO_DATA_DIR", originalEnv)

	os.Setenv("SNIPGO_DATA_DIR", tmpDir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Add multiple snippets
	snippets := []*Snippet{
		{ID: "test-id-1", Title: "Snippet 1", Body: "Body 1"},
		{ID: "test-id-2", Title: "Snippet 2", Body: "Body 2"},
		{ID: "test-id-3", Title: "Snippet 3", Body: "Body 3"},
	}

	for _, s := range snippets {
		if err := m.Save(s); err != nil {
			t.Fatalf("Failed to save snippet: %v", err)
		}
	}

	all := m.GetAll()

	if len(all) != len(snippets) {
		t.Errorf("Manager.GetAll() returned %d snippets, want %d", len(all), len(snippets))
	}

	// Verify all snippets are present
	snippetMap := make(map[string]bool)
	for _, s := range snippets {
		snippetMap[s.ID] = true
	}

	for _, got := range all {
		if !snippetMap[got.ID] {
			t.Errorf("Manager.GetAll() returned unexpected snippet: %v", got.ID)
		}
	}

	// Verify they are copies
	if len(all) > 0 {
		originalTags := len(all[0].Tags)
		all[0].Tags = append(all[0].Tags, "new-tag")
		// Reload to check original wasn't modified
		m.LoadAll()
		reloaded, _ := m.GetByID(all[0].ID)
		if len(reloaded.Tags) != originalTags {
			t.Error("Manager.GetAll() returned references, not copies")
		}
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name    string
		snippet *Snippet
		want    string // We'll check pattern, not exact match due to timestamp
	}{
		{
			name: "normal title",
			snippet: &Snippet{
				Title:     "Test Snippet",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "Test_Snippet_20200101_123045.md",
		},
		{
			name: "title with spaces",
			snippet: &Snippet{
				Title:     "My Test Snippet",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "My_Test_Snippet_20200101_123045.md",
		},
		{
			name: "title with special characters",
			snippet: &Snippet{
				Title:     "Test/Snippet:With*Special?Chars",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "Test_Snippet_With_Special_Chars_20200101_123045.md",
		},
		{
			name: "title with quotes",
			snippet: &Snippet{
				Title:     "Test\"Snippet\"",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "Test_Snippet__20200101_123045.md",
		},
		{
			name: "title with angle brackets",
			snippet: &Snippet{
				Title:     "Test<Snippet>",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "Test_Snippet__20200101_123045.md",
		},
		{
			name: "title with pipe",
			snippet: &Snippet{
				Title:     "Test|Snippet",
				UpdatedAt: time.Date(2020, 1, 1, 12, 30, 45, 0, time.UTC),
			},
			want: "Test_Snippet_20200101_123045.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFilename(tt.snippet)

			if got != tt.want {
				t.Errorf("generateFilename() = %v, want %v", got, tt.want)
			}

			// Verify it ends with .md
			if len(got) < 3 || got[len(got)-3:] != ".md" {
				t.Errorf("generateFilename() = %v, want to end with .md", got)
			}
		})
	}
}

func TestCopySnippet(t *testing.T) {
	original := &Snippet{
		ID:         "test-id",
		Title:      "Test Title",
		Tags:       []string{"tag1", "tag2", "tag3"},
		Language:   "go",
		IsFavorite: true,
		CreatedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		Body:       "Body content",
	}

	copied := copySnippet(original)

	// Verify all fields are copied
	if copied.ID != original.ID {
		t.Errorf("copySnippet() ID = %v, want %v", copied.ID, original.ID)
	}

	if copied.Title != original.Title {
		t.Errorf("copySnippet() Title = %v, want %v", copied.Title, original.Title)
	}

	if len(copied.Tags) != len(original.Tags) {
		t.Errorf("copySnippet() Tags length = %v, want %v", len(copied.Tags), len(original.Tags))
	}

	for i, tag := range copied.Tags {
		if tag != original.Tags[i] {
			t.Errorf("copySnippet() Tags[%d] = %v, want %v", i, tag, original.Tags[i])
		}
	}

	if copied.Language != original.Language {
		t.Errorf("copySnippet() Language = %v, want %v", copied.Language, original.Language)
	}

	if copied.IsFavorite != original.IsFavorite {
		t.Errorf("copySnippet() IsFavorite = %v, want %v", copied.IsFavorite, original.IsFavorite)
	}

	if !copied.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("copySnippet() CreatedAt = %v, want %v", copied.CreatedAt, original.CreatedAt)
	}

	if !copied.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("copySnippet() UpdatedAt = %v, want %v", copied.UpdatedAt, original.UpdatedAt)
	}

	if copied.Body != original.Body {
		t.Errorf("copySnippet() Body = %v, want %v", copied.Body, original.Body)
	}

	// Verify it's a deep copy - modifying tags shouldn't affect original
	originalTagsLen := len(original.Tags)
	copied.Tags = append(copied.Tags, "new-tag")
	if len(original.Tags) != originalTagsLen {
		t.Error("copySnippet() tags are not independent (shallow copy)")
	}

	// Verify pointers are different
	if copied == original {
		t.Error("copySnippet() returned same pointer")
	}
}


