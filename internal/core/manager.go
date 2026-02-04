package core

import (
	"bytes"
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	"snipgo/internal/git"
	"snipgo/internal/storage"
)

// Manager manages snippets in memory and on disk
type Manager struct {
	snippets map[string]*Snippet // key: snippet ID
	storage  *storage.FileSystem
	git      *git.GitManager
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

// SetGitManager sets the git manager for auto-commit functionality
func (m *Manager) SetGitManager(gm *git.GitManager) {
	m.git = gm
}

// GetGitManager returns the git manager (may be nil if not configured)
func (m *Manager) GetGitManager() *git.GitManager {
	return m.git
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
			slog.Warn("failed to read file", "path", filepath, "error", err)
			continue
		}

		snippet, err := ParseFrontmatter(content)
		if err != nil {
			// Log error but continue loading other files
			slog.Warn("failed to parse file", "path", filepath, "error", err)
			continue
		}

		if err := snippet.Validate(); err != nil {
			slog.Warn("invalid snippet in file", "path", filepath, "error", err)
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

	// Check if snippet already exists on disk
	existingFilePath := m.findFileBySnippetID(snippet.ID)

	// Update timestamp
	snippet.UpdateTimestamp()

	// Serialize to markdown
	content, err := SerializeFrontmatter(snippet)
	if err != nil {
		return fmt.Errorf("failed to serialize snippet: %w", err)
	}

	// Use existing file path if updating, otherwise generate new filename
	var filePath string
	if existingFilePath != "" {
		filePath = existingFilePath
	} else {
		filename := generateFilename(snippet)
		filePath = filepath.Join(m.storage.GetSnippetsDir(), filename)
	}

	// Write to disk
	if err := m.storage.WriteFile(filePath, content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update in-memory index
	m.snippets[snippet.ID] = snippet

	// Auto-commit if enabled
	m.autoCommitFile(filepath.Base(filePath), snippet, "save")

	return nil
}

// Delete deletes a snippet from disk and memory
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	snippet, exists := m.snippets[id]
	if !exists {
		return fmt.Errorf("snippet with ID %s not found", id)
	}

	// Find and delete the file
	filePath := m.findFileBySnippetID(id)
	if filePath != "" {
		if err := m.storage.DeleteFile(filePath); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}

		// Auto-commit deletion if enabled
		m.autoCommitFile(filepath.Base(filePath), snippet, "delete")
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

// findFileBySnippetID finds the file path for a snippet with the given ID.
// Returns empty string if not found.
func (m *Manager) findFileBySnippetID(id string) string {
	files, err := m.storage.ListFiles()
	if err != nil {
		return ""
	}

	for _, fp := range files {
		content, err := m.storage.ReadFile(fp)
		if err != nil {
			continue
		}

		fileSnippet, err := ParseFrontmatter(content)
		if err != nil {
			continue
		}

		if fileSnippet.ID == id {
			return fp
		}
	}

	return ""
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
}

// GetFilenameByID returns the filename (basename only) for a snippet ID
func (m *Manager) GetFilenameByID(id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filePath := m.findFileBySnippetID(id)
	if filePath == "" {
		return "", fmt.Errorf("file not found for snippet: %s", id)
	}

	return filepath.Base(filePath), nil
}

// autoCommitFile commits a file change if git auto-commit is enabled
func (m *Manager) autoCommitFile(filename string, snippet *Snippet, action string) {
	if m.git == nil || !m.git.IsEnabled() || !m.git.IsGitRepo() {
		return
	}

	cfg := m.git.GetConfig()
	if cfg == nil || !cfg.AutoCommit {
		return
	}

	// Generate commit message from template
	commitMsg := m.formatCommitMessage(snippet, action)

	// Add and commit the file
	if err := m.git.AddAndCommit(commitMsg, filename); err != nil {
		// Log warning but don't fail the save operation
		slog.Warn("git auto-commit failed", "file", filename, "error", err)
		return
	}

	slog.Debug("auto-committed file", "file", filename, "message", commitMsg)

	// Auto-push if enabled
	if cfg.AutoPush && m.git.HasRemote() {
		if err := m.git.Push(); err != nil {
			slog.Warn("git auto-push failed", "error", err)
		} else {
			slog.Debug("auto-pushed to remote")
		}
	}
}

// formatCommitMessage generates a commit message using the template
func (m *Manager) formatCommitMessage(snippet *Snippet, action string) string {
	cfg := m.git.GetConfig()
	if cfg == nil || cfg.CommitMessageTemplate == "" {
		// Default message based on action
		switch action {
		case "save":
			return fmt.Sprintf("Update: %s", snippet.Title)
		case "delete":
			return fmt.Sprintf("Delete: %s", snippet.Title)
		default:
			return fmt.Sprintf("Change: %s", snippet.Title)
		}
	}

	// Parse and execute template
	tmpl, err := template.New("commit").Parse(cfg.CommitMessageTemplate)
	if err != nil {
		return fmt.Sprintf("Update: %s", snippet.Title)
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Title":       snippet.Title,
		"ID":          snippet.ID,
		"Action":      action,
		"Description": snippet.Description,
		"Language":    snippet.Language,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Update: %s", snippet.Title)
	}

	return buf.String()
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
