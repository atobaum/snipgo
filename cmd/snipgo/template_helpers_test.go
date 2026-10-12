package main

import (
	"reflect"
	"testing"
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
