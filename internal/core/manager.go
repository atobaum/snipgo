package core

import (
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"snipgo/internal/storage"
	"snipgo/internal/tmpl"
)

// ErrSnippetNotFound is returned when no snippet exists for the given ID.
var ErrSnippetNotFound = errors.New("snippet not found")

const (
	maxFilenameRunes = 100
	fallbackFilename = "untitled"
)

// Manager manages snippets in memory and on disk
type Manager struct {
	snippets  map[string]*Snippet // key: snippet ID
	filePaths map[string]string   // key: snippet ID, value: absolute file path
	storage   *storage.FileSystem
	mu        sync.RWMutex
}

// NewManager creates a new Manager instance
func NewManager() (*Manager, error) {
	fs, err := storage.NewFileSystem()
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}

	m := &Manager{
		snippets:  make(map[string]*Snippet),
		filePaths: make(map[string]string),
		storage:   fs,
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

	// Clear existing state
	m.snippets = make(map[string]*Snippet)
	m.filePaths = make(map[string]string)

	// Load each file
	for _, fp := range files {
		content, err := m.storage.ReadFile(fp)
		if err != nil {
			slog.Warn("failed to read file", "path", fp, "error", err)
			continue
		}

		snippet, err := ParseFrontmatter(content)
		if err != nil {
			slog.Warn("failed to parse file", "path", fp, "error", err)
			continue
		}

		if err := snippet.Validate(); err != nil {
			slog.Warn("invalid snippet in file", "path", fp, "error", err)
			continue
		}

		m.snippets[snippet.ID] = snippet
		m.filePaths[snippet.ID] = fp
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

	currentPath, exists := m.filePaths[snippet.ID]
	filePath := m.availableFilePath(snippet.Title, currentPath)
	if exists && filePath != currentPath && m.storage.FileExists(currentPath) {
		if err := m.storage.RenameFile(currentPath, filePath); err != nil {
			return err
		}
		m.filePaths[snippet.ID] = filePath
	}

	// Write to disk
	if err := m.storage.WriteFile(filePath, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update in-memory index
	m.snippets[snippet.ID] = snippet
	m.filePaths[snippet.ID] = filePath

	return nil
}

// Delete deletes a snippet from disk and memory
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.snippets[id]; !exists {
		return fmt.Errorf("%w: %s", ErrSnippetNotFound, id)
	}

	// Delete the file using the cached path
	if fp, ok := m.filePaths[id]; ok {
		if err := m.storage.DeleteFile(fp); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// Remove from memory
	delete(m.snippets, id)
	delete(m.filePaths, id)

	return nil
}

// GetByID returns a snippet by ID
func (m *Manager) GetByID(id string) (*Snippet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snippet, exists := m.snippets[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSnippetNotFound, id)
	}

	// Return a copy to prevent external modifications
	return copySnippet(snippet), nil
}

// FilePath returns the file path backing the snippet with the given ID.
func (m *Manager) FilePath(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fp, ok := m.filePaths[id]
	return fp, ok
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

// generateFilename builds a filename from a title, appending -N for N > 1
func generateFilename(title string, n int) string {
	base := sanitizeFilename(title)
	if n > 1 {
		return fmt.Sprintf("%s-%d.md", base, n)
	}
	return base + ".md"
}

// sanitizeFilename keeps letters, digits, '-', '_' and '.', replacing other runs with a single '_'.
func sanitizeFilename(title string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '-' || r == '.' {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}

	const edgeChars = "_.-"
	name := strings.Trim(b.String(), edgeChars)
	if runes := []rune(name); len(runes) > maxFilenameRunes {
		name = strings.TrimRight(string(runes[:maxFilenameRunes]), edgeChars)
	}
	if name == "" {
		return fallbackFilename
	}
	return name
}

// availableFilePath returns the first title-based path not taken by another file.
// currentPath counts as available so a snippet keeps its own file.
func (m *Manager) availableFilePath(title, currentPath string) string {
	dir := m.storage.GetSnippetsDir()
	for n := 1; ; n++ {
		candidate := filepath.Join(dir, generateFilename(title, n))
		if candidate == currentPath || !m.storage.FileExists(candidate) || m.storage.SameFile(candidate, currentPath) {
			return candidate
		}
	}
}

// copySnippet creates a deep copy of a snippet
func copySnippet(s *Snippet) *Snippet {
	tags := make([]string, len(s.Tags))
	copy(tags, s.Tags)

	copy_ := &Snippet{
		ID:          s.ID,
		Title:       s.Title,
		Description: s.Description,
		Tags:        tags,
		Language:    s.Language,
		IsFavorite:  s.IsFavorite,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Body:        s.Body,
	}

	// Deep copy Variables map
	if s.Variables != nil {
		copy_.Variables = make(map[string]*tmpl.Variable, len(s.Variables))
		for k, v := range s.Variables {
			varCopy := *v
			if v.Choices != nil {
				varCopy.Choices = make([]string, len(v.Choices))
				copy(varCopy.Choices, v.Choices)
			}
			copy_.Variables[k] = &varCopy
		}
	}

	return copy_
}

// GetDataDir returns the directory where snippet data files are stored.
func (m *Manager) GetDataDir() string {
	return m.storage.GetSnippetsDir()
}

// GetAllTags returns all unique tags across all snippets, sorted alphabetically
func (m *Manager) GetAllTags() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tagSet := make(map[string]bool)
	for _, snippet := range m.snippets {
		for _, tag := range snippet.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)
	return tags
}
