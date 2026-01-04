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

// SearchOptions contains search query and filter criteria
type SearchOptions struct {
	Query    string   // Fuzzy search query (can be empty)
	Tags     []string // Filter by tags (AND logic, case-insensitive)
	Language string   // Filter by language (case-insensitive, empty means no filter)
}

// matchesTags checks if snippet contains ALL specified tags (AND logic)
func matchesTags(snippet *Snippet, filterTags []string) bool {
	if len(filterTags) == 0 {
		return true // No filter
	}

	// Build lowercase tag set for efficient lookup
	snippetTagSet := make(map[string]bool)
	for _, tag := range snippet.Tags {
		snippetTagSet[strings.ToLower(tag)] = true
	}

	// Check ALL filter tags are present
	for _, filterTag := range filterTags {
		if !snippetTagSet[strings.ToLower(filterTag)] {
			return false
		}
	}
	return true
}

// matchesLanguage checks if snippet language matches filter (case-insensitive)
func matchesLanguage(snippet *Snippet, filterLang string) bool {
	if filterLang == "" {
		return true // No filter
	}
	return strings.EqualFold(snippet.Language, filterLang)
}

// SearchWithFilters searches snippets with optional query and filters
func (m *Manager) SearchWithFilters(opts SearchOptions) []*SearchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Step 1: Apply filters to get candidate set
	candidates := make(map[string]*Snippet)
	for id, snippet := range m.snippets {
		if matchesTags(snippet, opts.Tags) && matchesLanguage(snippet, opts.Language) {
			candidates[id] = snippet
		}
	}

	// Step 2: If no query, return all filtered candidates with score 0
	if opts.Query == "" {
		results := make([]*SearchResult, 0, len(candidates))
		for _, snippet := range candidates {
			results = append(results, &SearchResult{
				Snippet: copySnippet(snippet),
				Score:   0,
			})
		}
		return results
	}

	// Step 3: Fuzzy search on filtered candidates
	queryLower := strings.ToLower(opts.Query)
	results := make([]*SearchResult, 0)

	// Build a list of titles for fuzzy search
	titles := make([]string, 0, len(candidates))
	snippetMap := make(map[string]*Snippet)
	for _, snippet := range candidates {
		titles = append(titles, snippet.Title)
		snippetMap[snippet.Title] = snippet
	}

	// Fuzzy search on titles
	matches := fuzzy.Find(opts.Query, titles)
	titleMatches := make(map[string]bool)
	for _, match := range matches {
		snippet := snippetMap[match.Str]
		results = append(results, &SearchResult{
			Snippet: copySnippet(snippet),
			Score:   match.Score,
		})
		titleMatches[snippet.ID] = true
	}

	// Substring matching on tags and body for candidates not already matched
	for _, snippet := range candidates {
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

// Search searches snippets by query (backward compatible wrapper)
func (m *Manager) Search(query string) []*SearchResult {
	return m.SearchWithFilters(SearchOptions{Query: query})
}
