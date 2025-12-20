package parser

import (
	"fmt"
	"strings"

	"snip-go/internal/core"

	"gopkg.in/yaml.v3"
)

const (
	frontmatterDelimiter = "---"
)

// ParseFrontmatter parses a markdown file with YAML frontmatter
func ParseFrontmatter(content []byte) (*core.Snippet, error) {
	text := string(content)
	lines := strings.Split(text, "\n")

	// Check if file starts with frontmatter delimiter
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != frontmatterDelimiter {
		return nil, fmt.Errorf("file does not start with frontmatter delimiter")
	}

	// Find the end of frontmatter
	var frontmatterLines []string
	var bodyStartIndex int
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelimiter {
			bodyStartIndex = i + 1
			break
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	if bodyStartIndex == 0 {
		return nil, fmt.Errorf("frontmatter delimiter not closed")
	}

	// Parse YAML frontmatter
	frontmatterText := strings.Join(frontmatterLines, "\n")
	snippet := &core.Snippet{}
	if err := yaml.Unmarshal([]byte(frontmatterText), snippet); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract body (everything after frontmatter)
	bodyLines := lines[bodyStartIndex:]
	body := strings.Join(bodyLines, "\n")
	// Remove leading newline if present
	body = strings.TrimPrefix(body, "\n")
	snippet.Body = body

	return snippet, nil
}

// SerializeFrontmatter converts a snippet to markdown with YAML frontmatter
func SerializeFrontmatter(snippet *core.Snippet) ([]byte, error) {
	// Validate snippet
	if err := snippet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid snippet: %w", err)
	}

	// Marshal frontmatter to YAML
	frontmatterBytes, err := yaml.Marshal(snippet)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Build the document
	var parts []string
	parts = append(parts, frontmatterDelimiter)
	parts = append(parts, string(frontmatterBytes))
	parts = append(parts, frontmatterDelimiter)
	parts = append(parts, "") // Empty line between frontmatter and body
	parts = append(parts, snippet.Body)

	result := strings.Join(parts, "\n")
	return []byte(result), nil
}

