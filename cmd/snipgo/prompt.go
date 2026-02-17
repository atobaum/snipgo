package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"snipgo/internal/history"
	"snipgo/internal/tmpl"
)

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
