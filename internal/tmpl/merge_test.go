package tmpl

import (
	"testing"
)

func TestMergeWithMetadata(t *testing.T) {
	tests := []struct {
		name         string
		bodyVarNames []string
		metadata     map[string]*Variable
		want         []*Variable
	}{
		{
			name:         "body vars with matching metadata",
			bodyVarNames: []string{"host", "port"},
			metadata: map[string]*Variable{
				"host": {Description: "Server hostname", Default: "localhost"},
				"port": {Description: "Server port", Default: "8080"},
			},
			want: []*Variable{
				{Name: "host", Description: "Server hostname", Default: "localhost"},
				{Name: "port", Description: "Server port", Default: "8080"},
			},
		},
		{
			name:         "body vars with no metadata",
			bodyVarNames: []string{"user", "pass"},
			metadata:     map[string]*Variable{},
			want: []*Variable{
				{Name: "user"},
				{Name: "pass"},
			},
		},
		{
			name:         "body vars with partial metadata",
			bodyVarNames: []string{"db", "table", "limit"},
			metadata: map[string]*Variable{
				"db":    {Description: "Database name", Default: "prod"},
				"limit": {Description: "Row limit", Default: "100"},
			},
			want: []*Variable{
				{Name: "db", Description: "Database name", Default: "prod"},
				{Name: "table"},
				{Name: "limit", Description: "Row limit", Default: "100"},
			},
		},
		{
			name:         "empty body vars",
			bodyVarNames: []string{},
			metadata: map[string]*Variable{
				"unused": {Description: "Not in body"},
			},
			want: []*Variable{},
		},
		{
			name:         "order follows body occurrence",
			bodyVarNames: []string{"z", "a", "m"},
			metadata: map[string]*Variable{
				"a": {Description: "First alphabetically"},
				"m": {Description: "Middle"},
				"z": {Description: "Last alphabetically"},
			},
			want: []*Variable{
				{Name: "z", Description: "Last alphabetically"},
				{Name: "a", Description: "First alphabetically"},
				{Name: "m", Description: "Middle"},
			},
		},
		{
			name:         "metadata not in body is excluded",
			bodyVarNames: []string{"used"},
			metadata: map[string]*Variable{
				"used":   {Description: "In body"},
				"unused": {Description: "Not in body"},
			},
			want: []*Variable{
				{Name: "used", Description: "In body"},
			},
		},
		{
			name:         "with choices field",
			bodyVarNames: []string{"env"},
			metadata: map[string]*Variable{
				"env": {
					Description: "Environment",
					Default:     "dev",
					Choices:     []string{"dev", "staging", "prod"},
				},
			},
			want: []*Variable{
				{
					Name:        "env",
					Description: "Environment",
					Default:     "dev",
					Choices:     []string{"dev", "staging", "prod"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeWithMetadata(tt.bodyVarNames, tt.metadata)

			if len(got) != len(tt.want) {
				t.Errorf("MergeWithMetadata() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("MergeWithMetadata()[%d].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Description != tt.want[i].Description {
					t.Errorf("MergeWithMetadata()[%d].Description = %q, want %q", i, got[i].Description, tt.want[i].Description)
				}
				if got[i].Default != tt.want[i].Default {
					t.Errorf("MergeWithMetadata()[%d].Default = %q, want %q", i, got[i].Default, tt.want[i].Default)
				}
				if len(got[i].Choices) != len(tt.want[i].Choices) {
					t.Errorf("MergeWithMetadata()[%d].Choices length = %d, want %d", i, len(got[i].Choices), len(tt.want[i].Choices))
					continue
				}
				for j := range got[i].Choices {
					if got[i].Choices[j] != tt.want[i].Choices[j] {
						t.Errorf("MergeWithMetadata()[%d].Choices[%d] = %q, want %q", i, j, got[i].Choices[j], tt.want[i].Choices[j])
					}
				}
			}
		})
	}
}
