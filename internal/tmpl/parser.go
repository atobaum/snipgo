package tmpl

import "regexp"

// varPattern matches template variables in the format ${VAR_NAME}.
var varPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// ExtractVariables extracts unique variable names from a template body.
// Returns variable names in order of first appearance.
// Skips escaped variables ($$\{VAR\}).
func ExtractVariables(body string) []string {
	// Match ${VAR} but not $${VAR}
	// Negative lookbehind isn't supported in Go's RE2, so we'll filter manually

	// Find all matches with positions
	matches := varPattern.FindAllStringSubmatchIndex(body, -1)
	if matches == nil {
		return []string{}
	}

	seen := make(map[string]bool)
	var result []string

	for _, match := range matches {
		// match[0] is the start of the full match (including ${})
		// match[1] is the end of the full match
		// match[2] is the start of the captured group (variable name)
		// match[3] is the end of the captured group

		// Check if preceded by $ (escaped)
		if match[0] > 0 && body[match[0]-1] == '$' {
			continue
		}

		varName := body[match[2]:match[3]]
		if !seen[varName] {
			seen[varName] = true
			result = append(result, varName)
		}
	}

	if result == nil {
		return []string{}
	}
	return result
}
