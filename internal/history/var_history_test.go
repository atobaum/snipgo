package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVarHistory_AddAndGet_SingleValue(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	err = h.Add("username", "alice")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	values := h.Get("username")
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "alice" {
		t.Errorf("expected 'alice', got '%s'", values[0])
	}
}

func TestVarHistory_AddDuplicate_MovesToFront(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	// Add three values
	h.Add("username", "alice")
	h.Add("username", "bob")
	h.Add("username", "charlie")

	values := h.Get("username")
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
	if values[0] != "charlie" || values[1] != "bob" || values[2] != "alice" {
		t.Errorf("expected [charlie, bob, alice], got %v", values)
	}

	// Add duplicate "bob" - should move to front, not duplicate
	h.Add("username", "bob")

	values = h.Get("username")
	if len(values) != 3 {
		t.Fatalf("expected 3 values after duplicate, got %d", len(values))
	}
	if values[0] != "bob" {
		t.Errorf("expected 'bob' at front, got '%s'", values[0])
	}
	if values[1] != "charlie" || values[2] != "alice" {
		t.Errorf("expected [bob, charlie, alice], got %v", values)
	}
}

func TestVarHistory_MaxEntries_DropsOldest(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	// Add 11 values (max is 10)
	for i := 0; i < 11; i++ {
		h.Add("port", string(rune('a'+i)))
	}

	values := h.Get("port")
	if len(values) != 10 {
		t.Fatalf("expected 10 values (max), got %d", len(values))
	}

	// Most recent should be 'k' (index 10)
	if values[0] != "k" {
		t.Errorf("expected most recent 'k', got '%s'", values[0])
	}

	// Oldest should be 'b' (index 1), 'a' (index 0) should be dropped
	if values[9] != "b" {
		t.Errorf("expected oldest 'b', got '%s'", values[9])
	}
}

func TestVarHistory_GetNonExistent_ReturnsEmptySlice(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	values := h.Get("nonexistent")
	if values == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(values) != 0 {
		t.Errorf("expected empty slice, got %d values", len(values))
	}
}

func TestVarHistory_Persistence_RoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	// Create first instance, add data
	h1, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	h1.Add("username", "alice")
	h1.Add("username", "bob")
	h1.Add("port", "8080")

	// Create second instance from same file
	h2, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory (second instance) failed: %v", err)
	}

	// Verify data persisted
	values := h2.Get("username")
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "bob" || values[1] != "alice" {
		t.Errorf("expected [bob, alice], got %v", values)
	}

	portValues := h2.Get("port")
	if len(portValues) != 1 {
		t.Fatalf("expected 1 port value, got %d", len(portValues))
	}
	if portValues[0] != "8080" {
		t.Errorf("expected '8080', got '%s'", portValues[0])
	}
}

func TestVarHistory_MissingFile_ReturnsEmptyHistory(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "nonexistent.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Errorf("expected no error for missing file, got: %v", err)
	}

	values := h.Get("anything")
	if len(values) != 0 {
		t.Errorf("expected empty history, got %d values", len(values))
	}
}

func TestVarHistory_CorruptFile_ReturnsEmptyHistory(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "corrupt.json")

	// Write corrupt JSON
	err := os.WriteFile(histPath, []byte("this is not valid JSON {{{"), 0644)
	if err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Errorf("expected no error for corrupt file, got: %v", err)
	}

	values := h.Get("anything")
	if len(values) != 0 {
		t.Errorf("expected empty history from corrupt file, got %d values", len(values))
	}
}

func TestVarHistory_GetAll(t *testing.T) {
	tempDir := t.TempDir()
	histPath := filepath.Join(tempDir, "history.json")

	h, err := NewVarHistory(histPath)
	if err != nil {
		t.Fatalf("NewVarHistory failed: %v", err)
	}

	h.Add("username", "alice")
	h.Add("username", "bob")
	h.Add("port", "8080")
	h.Add("host", "localhost")

	all := h.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(all))
	}

	if len(all["username"]) != 2 {
		t.Errorf("expected 2 username values, got %d", len(all["username"]))
	}
	if len(all["port"]) != 1 {
		t.Errorf("expected 1 port value, got %d", len(all["port"]))
	}
	if len(all["host"]) != 1 {
		t.Errorf("expected 1 host value, got %d", len(all["host"]))
	}
}
