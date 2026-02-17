package core

import "errors"

// ErrSnippetNotFound is returned when a snippet cannot be found by ID.
var ErrSnippetNotFound = errors.New("snippet not found")
