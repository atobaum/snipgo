package core

import (
	"sort"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

// SearchResult represents a search result with a score
type SearchResult struct {
	Snippet *Snippet
	Score   int
}

// UsageData holds usage statistics for a snippet used during sorting.
type UsageData struct {
	Count    int
	LastUsed time.Time
}

// SearchOptions contains search query and filter criteria
type SearchOptions struct {
	Query      string   // Fuzzy search query (can be empty)
	Tags       []string // Filter by tags (AND logic, case-insensitive)
	Language   string   // Filter by language (case-insensitive, empty means no filter)
	SortBy     string   // "frequency" (default) | "name" | "last_used"
	SortOrder  string   // "asc" | "desc" (empty = field default: frequency→desc, last_used→desc, name→asc)
	UsageData  map[string]UsageData // snippet ID -> usage stats (provided by caller)
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

// resolveOrder returns the effective sort direction ("asc" or "desc") for a given field.
// Field defaults: frequency→desc, last_used→desc, name→asc.
func resolveOrder(sortBy, sortOrder string) string {
	if sortOrder == "asc" || sortOrder == "desc" {
		return sortOrder
	}
	if sortBy == "name" {
		return "asc"
	}
	return "desc"
}

// applySort sorts results according to SortBy/SortOrder and, when a query was
// provided, preserves the fuzzy score as the primary key.
func applySort(results []*SearchResult, opts SearchOptions, hasQuery bool) {
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "frequency"
	}
	order := resolveOrder(sortBy, opts.SortOrder)
	desc := order == "desc"

	usageFor := func(id string) UsageData {
		if opts.UsageData != nil {
			return opts.UsageData[id]
		}
		return UsageData{}
	}

	sort.SliceStable(results, func(i, j int) bool {
		si := results[i]
		sj := results[j]

		// When a query is present, score is the primary key (higher = better, always desc).
		if hasQuery && si.Score != sj.Score {
			return si.Score > sj.Score
		}

		ui := usageFor(si.Snippet.ID)
		uj := usageFor(sj.Snippet.ID)

		switch sortBy {
		case "name":
			ti := strings.ToLower(si.Snippet.Title)
			tj := strings.ToLower(sj.Snippet.Title)
			if ti != tj {
				if desc {
					return ti > tj
				}
				return ti < tj
			}

		case "last_used":
			if !ui.LastUsed.Equal(uj.LastUsed) {
				if desc {
					return ui.LastUsed.After(uj.LastUsed)
				}
				return ui.LastUsed.Before(uj.LastUsed)
			}
			// tie-break: title asc
			return strings.ToLower(si.Snippet.Title) < strings.ToLower(sj.Snippet.Title)

		default: // "frequency"
			if ui.Count != uj.Count {
				if desc {
					return ui.Count > uj.Count
				}
				return ui.Count < uj.Count
			}
			// tie-break: title asc
			return strings.ToLower(si.Snippet.Title) < strings.ToLower(sj.Snippet.Title)
		}

		// tie-break for name sort: stable (preserve input order)
		return false
	})
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
		applySort(results, opts, false)
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

	// Substring matching on tags, description, and body for candidates not already matched
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

		// Check description
		if strings.Contains(strings.ToLower(snippet.Description), queryLower) {
			score += 5
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

	applySort(results, opts, true)
	return results
}

// Search searches snippets by query (backward compatible wrapper)
func (m *Manager) Search(query string) []*SearchResult {
	return m.SearchWithFilters(SearchOptions{Query: query})
}
