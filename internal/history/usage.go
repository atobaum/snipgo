package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// UsageEntry tracks usage statistics for a single snippet.
type UsageEntry struct {
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

// UsageTracker records how often each snippet is selected.
// Data is persisted as JSON to the configured path.
type UsageTracker struct {
	mu      sync.RWMutex
	entries map[string]UsageEntry // key: snippet ID
	path    string
}

// NewUsageTracker creates a tracker backed by the given file path.
// If the file is missing or corrupt, an empty tracker is returned (graceful degradation).
func NewUsageTracker(path string) (*UsageTracker, error) {
	t := &UsageTracker{
		entries: make(map[string]UsageEntry),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return t, nil
		}
		return t, nil // other read errors: graceful degradation
	}

	if err := json.Unmarshal(data, &t.entries); err != nil {
		t.entries = make(map[string]UsageEntry) // corrupt file: graceful degradation
	}

	return t, nil
}

// Record increments the usage count for snippetID and updates LastUsed to now.
// Changes are persisted to disk immediately.
func (t *UsageTracker) Record(snippetID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry := t.entries[snippetID]
	entry.Count++
	entry.LastUsed = time.Now()
	t.entries[snippetID] = entry

	return t.save()
}

// GetCount returns the usage count for snippetID. Returns 0 if not found.
func (t *UsageTracker) GetCount(snippetID string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.entries[snippetID].Count
}

// GetLastUsed returns the last-used time for snippetID. Returns zero time if not found.
func (t *UsageTracker) GetLastUsed(snippetID string) time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.entries[snippetID].LastUsed
}

// GetAll returns a copy of all usage entries.
func (t *UsageTracker) GetAll() map[string]UsageEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]UsageEntry, len(t.entries))
	for k, v := range t.entries {
		result[k] = v
	}
	return result
}

// save persists the current state to disk using an atomic write.
// Must be called with the write lock held.
func (t *UsageTracker) save() error {
	return atomicWriteJSON(t.path, t.entries)
}
