package tmpl

import (
	"strings"
	"testing"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		values      map[string]string
		expected    string
		shouldError bool
		errorText   string
	}{
		{
			name:        "simple substitution",
			body:        "hi ${NAME}",
			values:      map[string]string{"NAME": "World"},
			expected:    "hi World",
			shouldError: false,
		},
		{
			name:        "multiple variables",
			body:        "${A} ${B}",
			values:      map[string]string{"A": "1", "B": "2"},
			expected:    "1 2",
			shouldError: false,
		},
		{
			name:        "missing variable",
			body:        "${MISSING}",
			values:      map[string]string{},
			shouldError: true,
			errorText:   "MISSING",
		},
		{
			name:        "no variables",
			body:        "no vars",
			values:      map[string]string{},
			expected:    "no vars",
			shouldError: false,
		},
		{
			name:        "escaped variable",
			body:        "$${ESCAPED}",
			values:      map[string]string{},
			expected:    "${ESCAPED}",
			shouldError: false,
		},
		{
			name:        "mix of real and escaped variables",
			body:        "${A} $${B} ${C}",
			values:      map[string]string{"A": "x", "C": "z"},
			expected:    "x ${B} z",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Expand(tt.body, tt.values)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expand() expected error containing %q, got nil", tt.errorText)
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expand() error = %v, want error containing %q", err, tt.errorText)
				}
				return
			}

			if err != nil {
				t.Errorf("Expand() unexpected error: %v", err)
				return
			}

			if result.Expanded != tt.expected {
				t.Errorf("Expand() expanded = %q, want %q", result.Expanded, tt.expected)
			}

			// Verify Variables map contains the provided values
			for k, v := range tt.values {
				if result.Variables[k] != v {
					t.Errorf("Expand() Variables[%q] = %q, want %q", k, result.Variables[k], v)
				}
			}
		})
	}
}
