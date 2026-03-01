package history

import (
	"encoding/json"
	"os"
	"sync"
)

const maxPerVar = 10

type VarHistory struct {
	mu        sync.RWMutex
	entries   map[string][]string // variable name -> recent values (newest first)
	path      string              // file path for persistence
	maxPerVar int
}

// NewVarHistory creates a history store at the given path. Loads existing data if present.
// Returns empty history if file doesn't exist or is corrupt (graceful degradation).
func NewVarHistory(path string) (*VarHistory, error) {
	h := &VarHistory{
		entries:   make(map[string][]string),
		path:      path,
		maxPerVar: maxPerVar,
	}

	// Try to load existing data
	data, err := os.ReadFile(path)
	if err != nil {
		// Missing file is OK - return empty history
		if os.IsNotExist(err) {
			return h, nil
		}
		// Other read errors - return empty history
		return h, nil
	}

	// Try to unmarshal - if corrupt, return empty history
	if err := json.Unmarshal(data, &h.entries); err != nil {
		// Corrupt file - return empty history (graceful degradation)
		h.entries = make(map[string][]string)
		return h, nil
	}

	return h, nil
}

// Get returns recent values for a variable. Returns empty slice if not found.
func (h *VarHistory) Get(varName string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	values, exists := h.entries[varName]
	if !exists {
		return []string{} // Return empty slice, not nil
	}

	// Return a copy to prevent external mutations
	result := make([]string, len(values))
	copy(result, values)
	return result
}

// Add adds a value to the history for a variable. If the value already exists,
// it moves to the front. Persists to disk after adding.
func (h *VarHistory) Add(varName, value string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	values, exists := h.entries[varName]
	if !exists {
		// New variable
		h.entries[varName] = []string{value}
	} else {
		// Remove duplicate if exists
		newValues := []string{value}
		for _, v := range values {
			if v != value {
				newValues = append(newValues, v)
			}
		}

		// Enforce max limit
		if len(newValues) > h.maxPerVar {
			newValues = newValues[:h.maxPerVar]
		}

		h.entries[varName] = newValues
	}

	// Persist to disk
	return h.save()
}

// GetAll returns all entries. Returns a copy to prevent external mutations.
func (h *VarHistory) GetAll() map[string][]string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a deep copy
	result := make(map[string][]string)
	for k, v := range h.entries {
		valueCopy := make([]string, len(v))
		copy(valueCopy, v)
		result[k] = valueCopy
	}
	return result
}

// save persists the current state to disk. Must be called with lock held.
func (h *VarHistory) save() error {
	return atomicWriteJSON(h.path, h.entries)
}
