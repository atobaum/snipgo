package tmpl

// Variable represents a template variable definition.
// Name is populated via post-unmarshal hydration (see D9), not from YAML.
type Variable struct {
	Name        string   `yaml:"-" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Choices     []string `yaml:"choices,omitempty" json:"choices,omitempty"`
}

// TemplateResult holds the expansion output.
type TemplateResult struct {
	Expanded  string            // The expanded body text
	Variables map[string]string // Variable name -> resolved value
}
