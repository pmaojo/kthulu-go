package builder

// Plan represents the structure of the CLI plan used to generate override bindings.
type Plan struct {
	Replacements []Replacement       `json:"replace"`
	Decorations  []string            `json:"decorate"`
	Groups       map[string][]string `json:"group"`
}

// Replacement defines a replacement constructor and its contract details.
type Replacement struct {
	Interface      string `json:"interface"`
	Implementation string `json:"implementation"`
	Constructor    string `json:"constructor"`
}
