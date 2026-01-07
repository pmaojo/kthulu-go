package cmd

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English, cases.NoLower)

// exportName returns the title-cased version of name using language-specific
// rules. If name is empty, an empty string is returned.
func exportName(name string) string {
	if name == "" {
		return ""
	}
	return titleCaser.String(name)
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
