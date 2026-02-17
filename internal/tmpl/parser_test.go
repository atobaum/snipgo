package tmpl

import (
	"reflect"
	"testing"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single variable",
			input:    "hello ${NAME}",
			expected: []string{"NAME"},
		},
		{
			name:     "multiple variables with duplicates",
			input:    "${A} and ${B} and ${A}",
			expected: []string{"A", "B"},
		},
		{
			name:     "no variables",
			input:    "no variables here",
			expected: []string{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "escaped variable",
			input:    "$${ESCAPED}",
			expected: []string{},
		},
		{
			name:     "valid variable name with underscores and numbers",
			input:    "${valid_name_123}",
			expected: []string{"valid_name_123"},
		},
		{
			name:     "invalid variable name starting with number",
			input:    "${123invalid}",
			expected: []string{},
		},
		{
			name:     "multiline with multiple variables",
			input:    "${A}\n${B}\n${C}",
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "mix of real and escaped variables",
			input:    "mix ${REAL} and $${ESCAPED} and ${ALSO_REAL}",
			expected: []string{"REAL", "ALSO_REAL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractVariables(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractVariables(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
