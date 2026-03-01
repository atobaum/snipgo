package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewUsageTracker_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.json")

	tracker, err := NewUsageTracker(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tracker.GetCount("any-id") != 0 {
		t.Error("expected count 0 for missing ID")
	}
}

func TestNewUsageTracker_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.json")

	if err := os.WriteFile(path, []byte("not json{{{"), 0644); err != nil {
		t.Fatal(err)
	}

	tracker, err := NewUsageTracker(path)
	if err != nil {
		t.Fatalf("expected graceful degradation, got error: %v", err)
	}
	if tracker.GetCount("any-id") != 0 {
		t.Error("expected count 0 after corrupt file")
	}
}

func TestRecord_IncrementsCount(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))

	before := time.Now()
	if err := tracker.Record("snippet-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now()

	if count := tracker.GetCount("snippet-1"); count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	lastUsed := tracker.GetLastUsed("snippet-1")
	if lastUsed.Before(before) || lastUsed.After(after) {
		t.Errorf("LastUsed %v is outside expected range [%v, %v]", lastUsed, before, after)
	}
}

func TestRecord_MultipleIncrements(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))

	for i := 0; i < 5; i++ {
		if err := tracker.Record("snippet-1"); err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}

	if count := tracker.GetCount("snippet-1"); count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestRecord_MultipleSnippets(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))

	tracker.Record("a") //nolint
	tracker.Record("a") //nolint
	tracker.Record("b") //nolint

	if count := tracker.GetCount("a"); count != 2 {
		t.Errorf("expected a=2, got %d", count)
	}
	if count := tracker.GetCount("b"); count != 1 {
		t.Errorf("expected b=1, got %d", count)
	}
}

func TestGetCount_UnknownID(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))

	if count := tracker.GetCount("nonexistent"); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestGetLastUsed_UnknownID(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))

	lastUsed := tracker.GetLastUsed("nonexistent")
	if !lastUsed.IsZero() {
		t.Errorf("expected zero time, got %v", lastUsed)
	}
}

func TestGetAll_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	tracker, _ := NewUsageTracker(filepath.Join(dir, "usage.json"))
	tracker.Record("snippet-1") //nolint

	all := tracker.GetAll()
	all["snippet-1"] = UsageEntry{Count: 999}

	if count := tracker.GetCount("snippet-1"); count != 1 {
		t.Errorf("expected original count 1 after external mutation, got %d", count)
	}
}

func TestPersistAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.json")

	tracker1, _ := NewUsageTracker(path)
	tracker1.Record("snippet-1") //nolint
	tracker1.Record("snippet-1") //nolint
	tracker1.Record("snippet-2") //nolint

	// Reload from disk
	tracker2, err := NewUsageTracker(path)
	if err != nil {
		t.Fatalf("reload error: %v", err)
	}

	if count := tracker2.GetCount("snippet-1"); count != 2 {
		t.Errorf("expected snippet-1 count=2 after reload, got %d", count)
	}
	if count := tracker2.GetCount("snippet-2"); count != 1 {
		t.Errorf("expected snippet-2 count=1 after reload, got %d", count)
	}
	if tracker2.GetLastUsed("snippet-1").IsZero() {
		t.Error("expected non-zero LastUsed after reload")
	}
}
