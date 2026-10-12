package tmpl

import (
	"reflect"
	"testing"
)

func TestMergeWithMetadata(t *testing.T) {
	tests := []struct {
		name             string
		bodyVarNames     []string
		frontmatterVars  map[string]*Variable
		wantNames        []string
		wantDescriptions []string
		wantDefaults     []string
	}{
		{
			name:         "enrich body vars with frontmatter metadata",
			bodyVarNames: []string{"A", "B"},
			frontmatterVars: map[string]*Variable{
				"A": {Description: "Variable A", Default: "default_a"},
			},
			wantNames:        []string{"A", "B"},
			wantDescriptions: []string{"Variable A", ""},
			wantDefaults:     []string{"default_a", ""},
		},
		{
			name:         "exclude stale frontmatter variables (D10)",
			bodyVarNames: []string{"A"},
			frontmatterVars: map[string]*Variable{
				"A":     {Description: "Variable A"},
				"STALE": {Description: "Stale var"},
			},
			wantNames:        []string{"A"},
			wantDescriptions: []string{"Variable A"},
			wantDefaults:     []string{""},
		},
		{
			name:             "no body vars returns empty",
			bodyVarNames:     []string{},
			frontmatterVars:  map[string]*Variable{"A": {Description: "Var A"}},
			wantNames:        []string{},
			wantDescriptions: []string{},
			wantDefaults:     []string{},
		},
		{
			name:             "body vars only no frontmatter",
			bodyVarNames:     []string{"A", "B", "C"},
			frontmatterVars:  map[string]*Variable{},
			wantNames:        []string{"A", "B", "C"},
			wantDescriptions: []string{"", "", ""},
			wantDefaults:     []string{"", "", ""},
		},
		{
			name:         "preserve body occurrence order",
			bodyVarNames: []string{"Z", "A", "M"},
			frontmatterVars: map[string]*Variable{
				"A": {Description: "First"},
				"M": {Description: "Second"},
				"Z": {Description: "Third"},
			},
			wantNames:        []string{"Z", "A", "M"},
			wantDescriptions: []string{"Third", "First", "Second"},
			wantDefaults:     []string{"", "", ""},
		},
		{
			name:         "enrich with choices",
			bodyVarNames: []string{"ENV"},
			frontmatterVars: map[string]*Variable{
				"ENV": {
					Description: "Environment",
					Default:     "dev",
					Choices:     []string{"dev", "staging", "prod"},
				},
			},
			wantNames:        []string{"ENV"},
			wantDescriptions: []string{"Environment"},
			wantDefaults:     []string{"dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeWithMetadata(tt.bodyVarNames, tt.frontmatterVars)

			// Check length
			if len(got) != len(tt.wantNames) {
				t.Errorf("MergeWithMetadata() returned %d variables, want %d", len(got), len(tt.wantNames))
				return
			}

			// Check each variable
			for i, v := range got {
				if v.Name != tt.wantNames[i] {
					t.Errorf("MergeWithMetadata()[%d].Name = %q, want %q", i, v.Name, tt.wantNames[i])
				}
				if v.Description != tt.wantDescriptions[i] {
					t.Errorf("MergeWithMetadata()[%d].Description = %q, want %q", i, v.Description, tt.wantDescriptions[i])
				}
				if v.Default != tt.wantDefaults[i] {
					t.Errorf("MergeWithMetadata()[%d].Default = %q, want %q", i, v.Default, tt.wantDefaults[i])
				}
			}

			// Special check for choices in the last test
			if tt.name == "enrich with choices" && len(got) > 0 {
				wantChoices := []string{"dev", "staging", "prod"}
				if !reflect.DeepEqual(got[0].Choices, wantChoices) {
					t.Errorf("MergeWithMetadata()[0].Choices = %v, want %v", got[0].Choices, wantChoices)
				}
			}
		})
	}
}

func TestMergeWithMetadata_DoesNotMutateMetadata(t *testing.T) {
	metadata := map[string]*Variable{"A": {Description: "Variable A"}}

	got := MergeWithMetadata([]string{"A"}, metadata)
	got[0].Description = "changed"

	if metadata["A"].Description != "Variable A" {
		t.Errorf("metadata mutated: Description = %q", metadata["A"].Description)
	}
	if metadata["A"].Name != "" {
		t.Errorf("metadata mutated: Name = %q", metadata["A"].Name)
	}
}
