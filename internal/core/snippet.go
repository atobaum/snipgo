package core

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// Snippet represents a code snippet with metadata
type Snippet struct {
	ID         string    `yaml:"id" json:"id"`
	Title      string    `yaml:"title" json:"title"`
	Tags       []string  `yaml:"tags" json:"tags"`
	Language   string    `yaml:"language" json:"language"`
	IsFavorite bool      `yaml:"is_favorite" json:"is_favorite"`
	CreatedAt  time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt  time.Time `yaml:"updated_at" json:"updated_at"`
	Body       string    `yaml:"-" json:"body"` // Body is not in frontmatter
}

// generateID generates a ULID (26 characters, lexicographically sortable)
func generateID() string {
	entropy := rand.Reader
	ms := ulid.Timestamp(time.Now())
	return ulid.MustNew(ms, entropy).String()
}

// NewSnippet creates a new snippet with generated ULID and timestamps
func NewSnippet(title string) *Snippet {
	now := time.Now()
	return &Snippet{
		ID:         generateID(),
		Title:      title,
		Tags:       []string{},
		Language:   "",
		IsFavorite: false,
		CreatedAt:  now,
		UpdatedAt:  now,
		Body:       "",
	}
}

// Validate checks if the snippet has required fields
func (s *Snippet) Validate() error {
	if s.ID == "" {
		return ErrInvalidSnippet{Field: "id", Reason: "ID cannot be empty"}
	}
	if s.Title == "" {
		return ErrInvalidSnippet{Field: "title", Reason: "Title cannot be empty"}
	}
	return nil
}

// UpdateTimestamp updates the UpdatedAt field to current time
func (s *Snippet) UpdateTimestamp() {
	s.UpdatedAt = time.Now()
}

// ErrInvalidSnippet represents a validation error for a snippet
type ErrInvalidSnippet struct {
	Field  string
	Reason string
}

func (e ErrInvalidSnippet) Error() string {
	return "invalid snippet: " + e.Field + ": " + e.Reason
}
