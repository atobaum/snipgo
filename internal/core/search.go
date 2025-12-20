package core

import (
	"strings"

	"github.com/sahilm/fuzzy"
)

// SearchResult represents a search result with a score
type SearchResult struct {
	Snippet *Snippet
	Score   int
}

// Search searches snippets using fuzzy search for titles and substring matching for tags/body
func (m *Manager) Search(query string) []*SearchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if query == "" {
		// Return all snippets if query is empty
		results := make([]*SearchResult, 0, len(m.snippets))
		for _, snippet := range m.snippets {
			results = append(results, &SearchResult{
				Snippet: copySnippet(snippet),
				Score:   0,
			})
		}
		return results
	}

	queryLower := strings.ToLower(query)
	results := make([]*SearchResult, 0)

	// Build a list of titles for fuzzy search
	titles := make([]string, 0, len(m.snippets))
	snippetMap := make(map[string]*Snippet)
	for _, snippet := range m.snippets {
		titles = append(titles, snippet.Title)
		snippetMap[snippet.Title] = snippet
	}

	// Fuzzy search on titles
	matches := fuzzy.Find(query, titles)
	titleMatches := make(map[string]bool)
	for _, match := range matches {
		snippet := snippetMap[match.Str]
		results = append(results, &SearchResult{
			Snippet: copySnippet(snippet),
			Score:   match.Score,
		})
		titleMatches[snippet.ID] = true
	}

	// Substring matching on tags and body for snippets not already matched
	for _, snippet := range m.snippets {
		if titleMatches[snippet.ID] {
			continue // Already matched by title
		}

		score := 0

		// Check tags
		for _, tag := range snippet.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				score += 10
				break
			}
		}

		// Check body
		if strings.Contains(strings.ToLower(snippet.Body), queryLower) {
			score += 5
		}

		// Only include if there's a match
		if score > 0 {
			results = append(results, &SearchResult{
				Snippet: copySnippet(snippet),
				Score:   score,
			})
		}
	}

	// Sort by score (higher is better)
	// Simple bubble sort for small datasets
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

