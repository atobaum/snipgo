package main

import (
	"fmt"
	"os"
	"strings"

	"snipgo/internal/core"
	"snipgo/internal/history"
	"snipgo/internal/tmpl"
)

// parseVarFlags parses -v KEY=VALUE flags into a map.
func parseVarFlags(flags []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, f := range flags {
		idx := strings.Index(f, "=")
		if idx < 1 {
			return nil, fmt.Errorf("invalid variable format %q: expected KEY=VALUE", f)
		}
		key := f[:idx]
		value := f[idx+1:]
		result[key] = value
	}
	return result, nil
}

// mergeVariables merges body-detected variables with frontmatter metadata (D10).
// Returns variables in body occurrence order. Frontmatter-only variables are excluded.
// This is a thin wrapper around tmpl.MergeWithMetadata for CLI usage.
func mergeVariables(bodyVarNames []string, frontmatterVars map[string]*tmpl.Variable) []*tmpl.Variable {
	return tmpl.MergeWithMetadata(bodyVarNames, frontmatterVars)
}

// expandSnippetBody extracts variables, merges with provided values and defaults,
// and returns the expanded body. If interactive and missing values, prompts user.
// After successful expansion, saves all values to history.
func expandSnippetBody(snippet *core.Snippet, providedVars map[string]string, raw bool, interactive bool, hist *history.VarHistory) (string, error) {
	if raw {
		return snippet.Body, nil
	}

	bodyVarNames := tmpl.ExtractVariables(snippet.Body)
	if len(bodyVarNames) == 0 {
		return snippet.Body, nil
	}

	// Merge with frontmatter metadata
	vars := mergeVariables(bodyVarNames, snippet.Variables)

	// Collect values
	values := make(map[string]string)
	for k, v := range providedVars {
		values[k] = v
	}

	// Fill defaults for missing
	for _, v := range vars {
		if _, ok := values[v.Name]; !ok && v.Default != "" {
			values[v.Name] = v.Default
		}
	}

	// Check for missing values
	var missing []*tmpl.Variable
	for _, v := range vars {
		if _, ok := values[v.Name]; !ok {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		if !interactive {
			names := make([]string, len(missing))
			for i, v := range missing {
				names[i] = v.Name
			}
			return "", fmt.Errorf("missing values for variables: %s", strings.Join(names, ", "))
		}
		// Prompt for missing
		prompted, err := promptForVariables(vars, values, hist)
		if err != nil {
			return "", err
		}
		values = prompted
	}

	result, err := tmpl.Expand(snippet.Body, values)
	if err != nil {
		return "", err
	}

	// Save all values to history (graceful degradation if hist is nil)
	if hist != nil {
		for k, v := range values {
			if err := hist.Add(k, v); err != nil {
				// Log but don't fail - history is best-effort
				fmt.Fprintf(os.Stderr, "warning: failed to save variable to history: %v\n", err)
			}
		}
	}

	return result.Expanded, nil
}
