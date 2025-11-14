package generator

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// Pluralize provides simple English pluralization
func Pluralize(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)

	// Special cases
	specialCases := map[string]string{
		"child":  "children",
		"person": "people",
		"mouse":  "mice",
		"goose":  "geese",
		"foot":   "feet",
		"tooth":  "teeth",
		"man":    "men",
		"woman":  "women",
	}

	if plural, exists := specialCases[lower]; exists {
		// Preserve original case
		if isCapitalized(s) {
			return Capitalize(plural)
		}
		return plural
	}

	// Standard rules
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "ch") {
		return s + "es"
	}

	if strings.HasSuffix(lower, "y") && len(s) > 1 {
		prev := strings.ToLower(string(s[len(s)-2]))
		if !isVowel(prev) {
			return s[:len(s)-1] + "ies"
		}
	}

	if strings.HasSuffix(lower, "f") {
		return s[:len(s)-1] + "ves"
	}

	if strings.HasSuffix(lower, "fe") {
		return s[:len(s)-2] + "ves"
	}

	// Default: just add 's'
	return s + "s"
}

// isCapitalized checks if the first letter is uppercase
func isCapitalized(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsUpper(r)
}

// isVowel checks if a string represents a vowel
func isVowel(s string) bool {
	vowels := "aeiou"
	return strings.Contains(vowels, strings.ToLower(s))
}

// ToSnakeCase converts camelCase to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// ToKebabCase converts camelCase to kebab-case
func ToKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteByte('-')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
