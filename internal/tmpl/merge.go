package tmpl

// MergeWithMetadata enriches body variable names with frontmatter metadata (D10).
// Returns variables in body occurrence order. Frontmatter-only variables are excluded.
func MergeWithMetadata(bodyVarNames []string, metadata map[string]*Variable) []*Variable {
	result := make([]*Variable, 0, len(bodyVarNames))
	for _, name := range bodyVarNames {
		v := Variable{Name: name}
		if meta, ok := metadata[name]; ok {
			v = *meta
			v.Name = name
		}
		result = append(result, &v)
	}
	return result
}
