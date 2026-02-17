package main

import (
	"reflect"
	"testing"

	"snipgo/internal/tmpl"
)

func TestParseVarFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "simple key-value pairs",
			flags: []string{"A=1", "B=2"},
			want:  map[string]string{"A": "1", "B": "2"},
		},
		{
			name:    "invalid format missing equals",
			flags:   []string{"INVALID"},
			wantErr: true,
		},
		{
			name:  "value with equals sign",
			flags: []string{"A=1=2"},
			want:  map[string]string{"A": "1=2"},
		},
		{
			name:  "empty flags",
			flags: []string{},
			want:  map[string]string{},
		},
		{
			name:  "value with spaces",
			flags: []string{"MSG=hello world", "PATH=/usr/bin"},
			want:  map[string]string{"MSG": "hello world", "PATH": "/usr/bin"},
		},
		{
			name:    "key only with equals",
			flags:   []string{"KEY="},
			want:    map[string]string{"KEY": ""},
			wantErr: false,
		},
		{
			name:    "equals at start",
			flags:   []string{"=value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVarFlags(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVarFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVarFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeVariables(t *testing.T) {
	tests := []struct {
		name             string
		bodyVarNames     []string
		frontmatterVars  map[string]*tmpl.Variable
		wantNames        []string
		wantDescriptions []string
		wantDefaults     []string
	}{
		{
			name:         "enrich body vars with frontmatter metadata",
			bodyVarNames: []string{"A", "B"},
			frontmatterVars: map[string]*tmpl.Variable{
				"A": {Description: "Variable A", Default: "default_a"},
			},
			wantNames:        []string{"A", "B"},
			wantDescriptions: []string{"Variable A", ""},
			wantDefaults:     []string{"default_a", ""},
		},
		{
			name:         "exclude stale frontmatter variables (D10)",
			bodyVarNames: []string{"A"},
			frontmatterVars: map[string]*tmpl.Variable{
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
			frontmatterVars:  map[string]*tmpl.Variable{"A": {Description: "Var A"}},
			wantNames:        []string{},
			wantDescriptions: []string{},
			wantDefaults:     []string{},
		},
		{
			name:             "body vars only no frontmatter",
			bodyVarNames:     []string{"A", "B", "C"},
			frontmatterVars:  map[string]*tmpl.Variable{},
			wantNames:        []string{"A", "B", "C"},
			wantDescriptions: []string{"", "", ""},
			wantDefaults:     []string{"", "", ""},
		},
		{
			name:         "preserve body occurrence order",
			bodyVarNames: []string{"Z", "A", "M"},
			frontmatterVars: map[string]*tmpl.Variable{
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
			frontmatterVars: map[string]*tmpl.Variable{
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
			got := mergeVariables(tt.bodyVarNames, tt.frontmatterVars)

			// Check length
			if len(got) != len(tt.wantNames) {
				t.Errorf("mergeVariables() returned %d variables, want %d", len(got), len(tt.wantNames))
				return
			}

			// Check each variable
			for i, v := range got {
				if v.Name != tt.wantNames[i] {
					t.Errorf("mergeVariables()[%d].Name = %q, want %q", i, v.Name, tt.wantNames[i])
				}
				if v.Description != tt.wantDescriptions[i] {
					t.Errorf("mergeVariables()[%d].Description = %q, want %q", i, v.Description, tt.wantDescriptions[i])
				}
				if v.Default != tt.wantDefaults[i] {
					t.Errorf("mergeVariables()[%d].Default = %q, want %q", i, v.Default, tt.wantDefaults[i])
				}
			}

			// Special check for choices in the last test
			if tt.name == "enrich with choices" && len(got) > 0 {
				wantChoices := []string{"dev", "staging", "prod"}
				if !reflect.DeepEqual(got[0].Choices, wantChoices) {
					t.Errorf("mergeVariables()[0].Choices = %v, want %v", got[0].Choices, wantChoices)
				}
			}
		})
	}
}
