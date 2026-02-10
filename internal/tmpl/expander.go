package tmpl

import (
	"fmt"
	"regexp"
	"strings"
)

// Expand replaces template variables with provided values.
// Returns error if any variable is missing a value.
// Converts escaped variables ($${VAR}) to literal ${VAR}.
func Expand(body string, values map[string]string) (*TemplateResult, error) {
	// First, extract all variables to check for missing values
	variables := ExtractVariables(body)
	for _, varName := range variables {
		if _, ok := values[varName]; !ok {
			return nil, fmt.Errorf("missing value for variable: %s", varName)
		}
	}

	// Replace unescaped variables
	pattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	expanded := body

	// Track positions to avoid replacing escaped variables
	matches := pattern.FindAllStringSubmatchIndex(expanded, -1)
	if matches != nil {
		// Process matches in reverse order to maintain string positions
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			start := match[0]
			end := match[1]
			varName := expanded[match[2]:match[3]]

			// Skip if escaped (preceded by $)
			if start > 0 && expanded[start-1] == '$' {
				continue
			}

			// Replace with value
			if value, ok := values[varName]; ok {
				expanded = expanded[:start] + value + expanded[end:]
			}
		}
	}

	// Convert escaped variables ($${VAR}) to literal ${VAR}
	expanded = strings.ReplaceAll(expanded, "$${", "${")

	return &TemplateResult{
		Expanded:  expanded,
		Variables: values,
	}, nil
}
