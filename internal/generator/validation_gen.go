package generator

import (
	"fmt"
	"strings"
)

// GenerateValidationFile renders core/<module>_validation.go: a
// ValidationErrors type plus a Validate() method assembled from the field
// rules declared in the blueprint / CLI field DSL.
//
// Supported rules: required, min=N, max=N, email, oneof=a|b|c.
func GenerateValidationFile(moduleName, title string, fields []BackendField) string {
	var checks []string
	needsRegexp := false
	needsUTF8 := false

	for _, f := range fields {
		if f.Relation != "" {
			continue
		}
		for _, rule := range f.Rules {
			check, imp := buildRuleCheck(f, rule)
			if check == "" {
				fmt.Printf("⚠️  Warning: skipping unsupported rule %q on field %s.%s\n", rule.Name, moduleName, f.Name)
				continue
			}
			checks = append(checks, check)
			switch imp {
			case "regexp":
				needsRegexp = true
			case "unicode/utf8":
				needsUTF8 = true
			}
		}
	}

	var b strings.Builder
	b.WriteString("// @kthulu:validation:" + moduleName + "\n")
	b.WriteString("// Generated from the field rules declared in the project blueprint.\n")
	b.WriteString("package core\n\n")

	imports := []string{`"sort"`, `"strings"`}
	if needsUTF8 {
		imports = append(imports, `"unicode/utf8"`)
	}
	if needsRegexp {
		imports = append(imports, "\n\t\"regexp\"")
	}
	b.WriteString("import (\n\t" + strings.Join(imports, "\n\t") + "\n)\n\n")

	if needsRegexp {
		b.WriteString("var emailPattern = regexp.MustCompile(`^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$`)\n\n")
	}

	b.WriteString(`// ValidationErrors maps field names (JSON tags) to human-readable messages.
// It implements error so services can return it directly; HTTP handlers
// translate it into a 422 Unprocessable Entity response.
type ValidationErrors map[string]string

func (v ValidationErrors) Error() string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+" "+v[k])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

`)

	b.WriteString(fmt.Sprintf("// Validate checks the %s against its declared field rules.\n", title))
	b.WriteString(fmt.Sprintf("func (e *%s) Validate() error {\n", title))
	if len(checks) == 0 {
		b.WriteString("\treturn nil\n}\n")
		// strings/sort are still used by ValidationErrors.Error.
		return b.String()
	}
	b.WriteString("\terrs := ValidationErrors{}\n")
	for _, c := range checks {
		b.WriteString(c)
	}
	b.WriteString("\tif len(errs) > 0 {\n\t\treturn errs\n\t}\n\treturn nil\n}\n")
	return b.String()
}

// buildRuleCheck returns the Go source for a single rule check and the extra
// import it needs ("" if none). An empty check means the rule is unsupported
// for the field type.
func buildRuleCheck(f BackendField, rule FieldRule) (check string, extraImport string) {
	field := "e." + f.Name
	tag := strings.TrimSuffix(f.JSONTag, ",omitempty")
	set := func(msg string) string {
		return fmt.Sprintf("\t\terrs[%q] = %q\n", tag, msg)
	}

	switch rule.Name {
	case "required":
		switch f.Type {
		case "string":
			return fmt.Sprintf("\tif strings.TrimSpace(%s) == \"\" {\n%s\t}\n", field, set("is required")), ""
		case "int", "uint", "float64":
			return fmt.Sprintf("\tif %s == 0 {\n%s\t}\n", field, set("is required")), ""
		case "time.Time":
			return fmt.Sprintf("\tif %s.IsZero() {\n%s\t}\n", field, set("is required")), ""
		}
	case "min":
		switch f.Type {
		case "string":
			return fmt.Sprintf("\tif strings.TrimSpace(%s) != \"\" && utf8.RuneCountInString(%s) < %s {\n%s\t}\n",
				field, field, rule.Param, set("must be at least "+rule.Param+" characters")), "unicode/utf8"
		case "int", "uint", "float64":
			return fmt.Sprintf("\tif %s < %s {\n%s\t}\n", field, rule.Param, set("must be at least "+rule.Param)), ""
		}
	case "max":
		switch f.Type {
		case "string":
			return fmt.Sprintf("\tif utf8.RuneCountInString(%s) > %s {\n%s\t}\n",
				field, rule.Param, set("must be at most "+rule.Param+" characters")), "unicode/utf8"
		case "int", "uint", "float64":
			return fmt.Sprintf("\tif %s > %s {\n%s\t}\n", field, rule.Param, set("must be at most "+rule.Param)), ""
		}
	case "email":
		if f.Type == "string" {
			return fmt.Sprintf("\tif %s != \"\" && !emailPattern.MatchString(%s) {\n%s\t}\n",
				field, field, set("must be a valid email address")), "regexp"
		}
	case "oneof":
		if f.Type == "string" && rule.Param != "" {
			values := strings.Split(rule.Param, "|")
			quoted := make([]string, len(values))
			for i, v := range values {
				quoted[i] = fmt.Sprintf("%q", v)
			}
			cond := make([]string, len(values))
			for i, v := range values {
				cond[i] = fmt.Sprintf("%s != %q", field, v)
			}
			return fmt.Sprintf("\tif %s != \"\" && %s {\n%s\t}\n",
				field, strings.Join(cond, " && "),
				set("must be one of: "+strings.Join(values, ", "))), ""
		}
	}
	return "", ""
}
