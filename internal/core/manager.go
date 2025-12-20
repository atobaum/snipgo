package core

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"snip-go/internal/storage"
)

// Manager manages snippets in memory and on disk
type Manager struct {
	snippets map[string]*Snippet // key: snippet ID
	storage  *storage.FileSystem
	mu       sync.RWMutex
}

// NewManager creates a new Manager instance
func NewManager() (*Manager, error) {
	fs, err := storage.NewFileSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}

	m := &Manager{
		snippets: make(map[string]*Snippet),
		storage:  fs,
	}

	return m, nil
}

// LoadAll loads all snippets from disk into memory
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := m.storage.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Clear existing snippets
	m.snippets = make(map[string]*Snippet)

	// Load each file
	for _, filepath := range files {
		content, err := m.storage.ReadFile(filepath)
		if err != nil {
			// Log error but continue loading other files
			fmt.Printf("Warning: failed to read file %s: %v\n", filepath, err)
			continue
		}

		snippet, err := ParseFrontmatter(content)
		if err != nil {
			// Log error but continue loading other files
			fmt.Printf("Warning: failed to parse file %s: %v\n", filepath, err)
			continue
		}

		if err := snippet.Validate(); err != nil {
			fmt.Printf("Warning: invalid snippet in file %s: %v\n", filepath, err)
			continue
		}

		m.snippets[snippet.ID] = snippet
	}

	return nil
}

// Save saves a snippet to disk
func (m *Manager) Save(snippet *Snippet) error {
	if err := snippet.Validate(); err != nil {
		return fmt.Errorf("invalid snippet: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update timestamp
	snippet.UpdateTimestamp()

	// Serialize to markdown
		content, err := SerializeFrontmatter(snippet)
	if err != nil {
		return fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Generate filename: {Title}_{Timestamp}.md
	filename := generateFilename(snippet)
	filepath := filepath.Join(m.storage.GetSnippetsDir(), filename)

	// Write to disk
	if err := m.storage.WriteFile(filepath, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update in-memory index
	m.snippets[snippet.ID] = snippet

	return nil
}

// Delete deletes a snippet from disk and memory
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snippets[id]; !exists {
		return fmt.Errorf("snippet with ID %s not found", id)
	}

	// Find and delete the file
	files, err := m.storage.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, filepath := range files {
		content, err := m.storage.ReadFile(filepath)
		if err != nil {
			continue
		}

		fileSnippet, err := ParseFrontmatter(content)
		if err != nil {
			continue
		}

		if fileSnippet.ID == id {
			if err := m.storage.DeleteFile(filepath); err != nil {
				return fmt.Errorf("failed to delete file: %w", err)
			}
			break
		}
	}

	// Remove from memory
	delete(m.snippets, id)

	return nil
}

// GetByID returns a snippet by ID
func (m *Manager) GetByID(id string) (*Snippet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snippet, exists := m.snippets[id]
	if !exists {
		return nil, fmt.Errorf("snippet with ID %s not found", id)
	}

	// Return a copy to prevent external modifications
	return copySnippet(snippet), nil
}

// GetAll returns all snippets
func (m *Manager) GetAll() []*Snippet {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snippets := make([]*Snippet, 0, len(m.snippets))
	for _, snippet := range m.snippets {
		snippets = append(snippets, copySnippet(snippet))
	}

	return snippets
}

// generateFilename generates a filename for a snippet
func generateFilename(snippet *Snippet) string {
	// Sanitize title for filename
	title := strings.ReplaceAll(snippet.Title, " ", "_")
	title = strings.ReplaceAll(title, "/", "_")
	title = strings.ReplaceAll(title, "\\", "_")
	title = strings.ReplaceAll(title, ":", "_")
	title = strings.ReplaceAll(title, "*", "_")
	title = strings.ReplaceAll(title, "?", "_")
	title = strings.ReplaceAll(title, "\"", "_")
	title = strings.ReplaceAll(title, "<", "_")
	title = strings.ReplaceAll(title, ">", "_")
	title = strings.ReplaceAll(title, "|", "_")

	// Use timestamp for uniqueness
	timestamp := snippet.UpdatedAt.Format("20060102_150405")

	return fmt.Sprintf("%s_%s.md", title, timestamp)
}

// copySnippet creates a deep copy of a snippet
func copySnippet(s *Snippet) *Snippet {
	tags := make([]string, len(s.Tags))
	copy(tags, s.Tags)

	return &Snippet{
		ID:         s.ID,
		Title:      s.Title,
		Tags:       tags,
		Language:   s.Language,
		IsFavorite: s.IsFavorite,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
		Body:       s.Body,
	}
}

