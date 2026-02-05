package main

import (
	"strings"
	"testing"
)

func TestScanMultiLineFromReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line with two blank lines",
			input:    "echo hello\n\n\n",
			expected: "echo hello",
		},
		{
			name:     "multiple lines with two blank lines to finish",
			input:    "docker system prune -af\ndocker volume prune -f\ndocker network prune -f\n\n\n",
			expected: "docker system prune -af\ndocker volume prune -f\ndocker network prune -f",
		},
		{
			name:     "lines with one blank line in between",
			input:    "line1\n\nline2\n\n\n",
			expected: "line1\n\nline2",
		},
		{
			name:     "empty input (just two blank lines)",
			input:    "\n\n",
			expected: "",
		},
		{
			name:     "input with trailing content after two blanks",
			input:    "first\nsecond\n\n\nignored",
			expected: "first\nsecond",
		},
		{
			name:     "three lines no intermediate blanks",
			input:    "line1\nline2\nline3\n\n\n",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "single blank line does not terminate",
			input:    "before\n\nafter\n\n\n",
			expected: "before\n\nafter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := scanMultiLineFromReader(reader, "")
			if err != nil {
				t.Errorf("scanMultiLineFromReader() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("scanMultiLineFromReader() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestScanMultiLineFromReader_EOF(t *testing.T) {
	// Test EOF without two blank lines (user sends EOF with Ctrl+D)
	input := "line1\nline2"
	reader := strings.NewReader(input)
	result, err := scanMultiLineFromReader(reader, "")
	if err != nil {
		t.Errorf("scanMultiLineFromReader() error = %v", err)
		return
	}
	if result != "line1\nline2" {
		t.Errorf("scanMultiLineFromReader() = %q, want %q", result, "line1\nline2")
	}
}
