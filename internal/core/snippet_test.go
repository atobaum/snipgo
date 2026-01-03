package core

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSnippet(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		wantTitle string
	}{
		{
			name:      "valid title",
			title:     "Test Snippet",
			wantTitle: "Test Snippet",
		},
		{
			name:      "empty title",
			title:     "",
			wantTitle: "",
		},
		{
			name:      "title with special characters",
			title:     "Test & Snippet < >",
			wantTitle: "Test & Snippet < >",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snippet := NewSnippet(tt.title)

			if snippet.Title != tt.wantTitle {
				t.Errorf("NewSnippet() Title = %v, want %v", snippet.Title, tt.wantTitle)
			}

			// Validate UUID format
			if _, err := uuid.Parse(snippet.ID); err != nil {
				t.Errorf("NewSnippet() ID = %v, want valid UUID, got error: %v", snippet.ID, err)
			}

			// Validate initial values
			if len(snippet.Tags) != 0 {
				t.Errorf("NewSnippet() Tags = %v, want empty slice", snippet.Tags)
			}

			if snippet.Language != "" {
				t.Errorf("NewSnippet() Language = %v, want empty string", snippet.Language)
			}

			if snippet.IsFavorite != false {
				t.Errorf("NewSnippet() IsFavorite = %v, want false", snippet.IsFavorite)
			}

			if snippet.Body != "" {
				t.Errorf("NewSnippet() Body = %v, want empty string", snippet.Body)
			}

			// Validate timestamps are set and close to current time
			now := time.Now()
			if snippet.CreatedAt.IsZero() {
				t.Error("NewSnippet() CreatedAt is zero")
			}

			if snippet.UpdatedAt.IsZero() {
				t.Error("NewSnippet() UpdatedAt is zero")
			}

			// Timestamps should be within 1 second of now
			if diff := now.Sub(snippet.CreatedAt); diff > time.Second || diff < -time.Second {
				t.Errorf("NewSnippet() CreatedAt = %v, want close to %v", snippet.CreatedAt, now)
			}

			if diff := now.Sub(snippet.UpdatedAt); diff > time.Second || diff < -time.Second {
				t.Errorf("NewSnippet() UpdatedAt = %v, want close to %v", snippet.UpdatedAt, now)
			}

			// CreatedAt and UpdatedAt should be equal initially
			if !snippet.CreatedAt.Equal(snippet.UpdatedAt) {
				t.Errorf("NewSnippet() CreatedAt = %v, UpdatedAt = %v, want equal", snippet.CreatedAt, snippet.UpdatedAt)
			}
		})
	}
}

func TestSnippet_Validate(t *testing.T) {
	tests := []struct {
		name    string
		snippet *Snippet
		wantErr bool
		errField string
	}{
		{
			name: "valid snippet",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "Test Title",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			snippet: &Snippet{
				ID:    "",
				Title: "Test Title",
			},
			wantErr:  true,
			errField: "id",
		},
		{
			name: "empty Title",
			snippet: &Snippet{
				ID:    "test-id",
				Title: "",
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "both empty",
			snippet: &Snippet{
				ID:    "",
				Title: "",
			},
			wantErr:  true,
			errField: "id", // First error is ID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.snippet.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Snippet.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("Snippet.Validate() expected error, got nil")
					return
				}

				// Check error type
				invalidErr, ok := err.(ErrInvalidSnippet)
				if !ok {
					t.Errorf("Snippet.Validate() error type = %T, want ErrInvalidSnippet", err)
					return
				}

				// Check error field
				if invalidErr.Field != tt.errField {
					t.Errorf("Snippet.Validate() error field = %v, want %v", invalidErr.Field, tt.errField)
				}

				// Check error message format
				expectedMsg := "invalid snippet: " + tt.errField + ": "
				if tt.errField == "id" {
					expectedMsg += "ID cannot be empty"
				} else {
					expectedMsg += "Title cannot be empty"
				}

				if err.Error() != expectedMsg {
					t.Errorf("Snippet.Validate() error message = %v, want %v", err.Error(), expectedMsg)
				}
			}
		})
	}
}

func TestSnippet_UpdateTimestamp(t *testing.T) {
	// Create a snippet with old timestamp
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	snippet := &Snippet{
		ID:        "test-id",
		Title:     "Test Title",
		UpdatedAt: oldTime,
		CreatedAt: oldTime,
	}

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update timestamp
	snippet.UpdateTimestamp()

	// Verify UpdatedAt changed
	if snippet.UpdatedAt.Equal(oldTime) {
		t.Error("Snippet.UpdateTimestamp() UpdatedAt did not change")
	}

	// Verify UpdatedAt is close to current time
	now := time.Now()
	if diff := now.Sub(snippet.UpdatedAt); diff > time.Second || diff < -time.Second {
		t.Errorf("Snippet.UpdateTimestamp() UpdatedAt = %v, want close to %v", snippet.UpdatedAt, now)
	}

	// Verify CreatedAt did not change
	if !snippet.CreatedAt.Equal(oldTime) {
		t.Errorf("Snippet.UpdateTimestamp() CreatedAt = %v, want %v (should not change)", snippet.CreatedAt, oldTime)
	}
}

func TestErrInvalidSnippet_Error(t *testing.T) {
	tests := []struct {
		name   string
		err    ErrInvalidSnippet
		want   string
	}{
		{
			name: "ID error",
			err: ErrInvalidSnippet{
				Field:  "id",
				Reason: "ID cannot be empty",
			},
			want: "invalid snippet: id: ID cannot be empty",
		},
		{
			name: "Title error",
			err: ErrInvalidSnippet{
				Field:  "title",
				Reason: "Title cannot be empty",
			},
			want: "invalid snippet: title: Title cannot be empty",
		},
		{
			name: "custom error",
			err: ErrInvalidSnippet{
				Field:  "tags",
				Reason: "Invalid tag format",
			},
			want: "invalid snippet: tags: Invalid tag format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ErrInvalidSnippet.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

