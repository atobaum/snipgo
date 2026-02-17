package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"snipgo/internal/core"
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
func promptForVariables(vars []*tmpl.Variable, provided map[string]string) (map[string]string, error) {
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

		// Build prompt
		prompt := v.Name
		if v.Description != "" {
			prompt += " (" + v.Description + ")"
		}
		if v.Default != "" {
			prompt += " [" + v.Default + "]"
		}

		var input string
		if len(v.Choices) > 0 {
			// Show numbered list (D11)
			fmt.Fprintf(os.Stderr, "%s:\n", prompt)
			for i, choice := range v.Choices {
				suffix := ""
				if choice == v.Default {
					suffix = " (default)"
				}
				fmt.Fprintf(os.Stderr, "  %d) %s%s\n", i+1, choice, suffix)
			}
			defaultNum := "1"
			if v.Default != "" {
				for i, c := range v.Choices {
					if c == v.Default {
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
				input = v.Default
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
			if input == "" && v.Default != "" {
				input = v.Default
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
func expandSnippetBody(snippet *core.Snippet, providedVars map[string]string, raw bool, interactive bool) (string, error) {
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
		prompted, err := promptForVariables(vars, values)
		if err != nil {
			return "", err
		}
		values = prompted
	}

	result, err := tmpl.Expand(snippet.Body, values)
	if err != nil {
		return "", err
	}
	return result.Expanded, nil
}
