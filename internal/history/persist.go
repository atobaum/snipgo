package history

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// atomicWriteJSON marshals v as indented JSON and writes it to path atomically.
// It ensures the parent directory exists, writes to a temp file, then renames.
// Must be called with the caller's write lock held.
func atomicWriteJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}
