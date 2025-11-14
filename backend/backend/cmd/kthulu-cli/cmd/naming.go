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
