package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
func mergeVariables(bodyVarNames []string, frontmatterVars map[string]*tmpl.Variable) []*tmpl.Variable {
	var result []*tmpl.Variable
	for _, name := range bodyVarNames {
		if fv, ok := frontmatterVars[name]; ok {
			// Enrich with frontmatter metadata
			v := *fv
			v.Name = name
			result = append(result, &v)
		} else {
			// Body-only variable, no metadata
			result = append(result, &tmpl.Variable{Name: name})
		}
	}
	return result
}

// promptForVariables interactively prompts the user for missing variable values.
// Uses the choice display format from D11.
// Priority: CLI -v flag > frontmatter default > history last-used > empty
func promptForVariables(vars []*tmpl.Variable, provided map[string]string, hist *history.VarHistory) (map[string]string, error) {
	values := make(map[string]string)
	// Copy provided values first
	for k, v := range provided {
		values[k] = v
	}

	scanner := bufio.NewScanner(os.Stdin)

	for _, v := range vars {
		if _, ok := values[v.Name]; ok {
			continue // Already provided via -v flag
		}

		// Determine default: frontmatter default > history last-used > empty
		defaultValue := v.Default
		if defaultValue == "" && hist != nil {
			histValues := hist.Get(v.Name)
			if len(histValues) > 0 {
				defaultValue = histValues[0] // Most recent value
			}
		}

		// Build prompt
		prompt := v.Name
		if v.Description != "" {
			prompt += " (" + v.Description + ")"
		}
		if defaultValue != "" {
			prompt += " [" + defaultValue + "]"
		}

		var input string
		if len(v.Choices) > 0 {
			// Show numbered list (D11)
			fmt.Fprintf(os.Stderr, "%s:\n", prompt)
			for i, choice := range v.Choices {
				suffix := ""
				if choice == defaultValue {
					suffix = " (default)"
				}
				fmt.Fprintf(os.Stderr, "  %d) %s%s\n", i+1, choice, suffix)
			}
			defaultNum := "1"
			if defaultValue != "" {
				for i, c := range v.Choices {
					if c == defaultValue {
						defaultNum = fmt.Sprintf("%d", i+1)
						break
					}
				}
			}
			fmt.Fprintf(os.Stderr, "Enter choice [%s] or custom value: ", defaultNum)

			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return nil, fmt.Errorf("failed to read input: %w", err)
				}
				return nil, fmt.Errorf("no input received")
			}
			input = strings.TrimSpace(scanner.Text())

			if input == "" {
				input = defaultValue
			} else {
				// Check if input is a number
				if num, err := strconv.Atoi(input); err == nil {
					if num >= 1 && num <= len(v.Choices) {
						input = v.Choices[num-1]
					}
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s: ", prompt)
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return nil, fmt.Errorf("failed to read input: %w", err)
				}
				return nil, fmt.Errorf("no input received")
			}
			input = strings.TrimSpace(scanner.Text())
			if input == "" && defaultValue != "" {
				input = defaultValue
			}
		}

		if input == "" {
			return nil, fmt.Errorf("no value provided for variable: %s", v.Name)
		}
		values[v.Name] = input
	}
	return values, nil
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
