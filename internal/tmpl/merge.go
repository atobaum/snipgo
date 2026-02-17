package tmpl

// MergeWithMetadata merges body-detected variable names with frontmatter
// variable metadata. Returns variables in body occurrence order.
// Variables present only in frontmatter (not in body) are excluded.
func MergeWithMetadata(bodyVarNames []string, metadata map[string]*Variable) []*Variable {
	var result []*Variable
	for _, name := range bodyVarNames {
		if fv, ok := metadata[name]; ok {
			v := *fv
			v.Name = name
			result = append(result, &v)
		} else {
			result = append(result, &Variable{Name: name})
		}
	}
	return result
}
