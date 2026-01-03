package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"snipgo/internal/core"
)

// serializeSnippetForEdit creates a markdown file with frontmatter for editing
func serializeSnippetForEdit(snippet *core.Snippet) ([]byte, error) {
	// Use the core package's SerializeFrontmatter function
	return core.SerializeFrontmatter(snippet)
}

// parseSnippetFromEdit parses a markdown file edited by the user
func parseSnippetFromEdit(content []byte) (*core.Snippet, error) {
	// Use the core package to parse frontmatter
	parsedSnippet, err := core.ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return parsedSnippet, nil
}

// formatSnippetForFzf formats a snippet for fzf display in pet CLI style: [Title] Body #tag1 #tag2
func formatSnippetForFzf(snippet *core.Snippet) string {
	// Get first line of body for display
	bodyFirstLine := ""
	if snippet.Body != "" {
		bodyLines := strings.Split(snippet.Body, "\n")
		bodyFirstLine = bodyLines[0]
		if len(bodyFirstLine) > 100 {
			bodyFirstLine = bodyFirstLine[:100] + "..."
		}
	}

	// Format: [Title] Body
	line := fmt.Sprintf("[%s] %s", snippet.Title, bodyFirstLine)

	// Add tags if present
	if len(snippet.Tags) > 0 {
		tagStr := strings.Join(snippet.Tags, " #")
		line += " #" + tagStr
	}

	return line
}

// selectSnippetWithFzf uses fzf to interactively select a snippet from the given list
func selectSnippetWithFzf(snippets []*core.Snippet) (*core.Snippet, error) {
	if len(snippets) == 0 {
		return nil, fmt.Errorf("no snippets to select from")
	}

	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, fmt.Errorf("fzf is not installed. Please install fzf first: https://github.com/junegunn/fzf")
	}

	// Build fzf input: format each snippet
	var fzfInput strings.Builder
	snippetMap := make(map[string]*core.Snippet) // Map from formatted line to snippet
	for _, snippet := range snippets {
		formatted := formatSnippetForFzf(snippet)
		fzfInput.WriteString(formatted)
		fzfInput.WriteString("\n")
		snippetMap[formatted] = snippet
	}

	// Run fzf
	cmd := exec.Command("fzf", "--ansi", "--height", "40%")
	cmd.Stdin = strings.NewReader(fzfInput.String())
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		// fzf returns exit code 1 when user cancels (ESC)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, fmt.Errorf("fzf error: %w", err)
	}

	// Parse selected line
	selectedLine := strings.TrimSpace(string(output))
	if selectedLine == "" {
		return nil, fmt.Errorf("no snippet selected")
	}

	// Find snippet by matching the selected line
	selectedSnippet, found := snippetMap[selectedLine]
	if !found {
		// If exact match not found, try to extract title from the line
		// Format: [Title] Body #tags
		if strings.HasPrefix(selectedLine, "[") {
			endIdx := strings.Index(selectedLine, "]")
			if endIdx > 0 {
				title := selectedLine[1:endIdx]
				// Find snippet by title
				for _, snippet := range snippets {
					if snippet.Title == title {
						return snippet, nil
					}
				}
			}
		}
		return nil, fmt.Errorf("could not find selected snippet")
	}

	return selectedSnippet, nil
}


