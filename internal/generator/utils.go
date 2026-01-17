package generator

import (
	"strings"

	"github.com/jinzhu/inflection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English, cases.NoLower)

func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	return titleCaser.String(s)
}

func Pluralize(s string) string {
	if s == "" {
		return ""
	}
	return inflection.Plural(s)
}

func Singularize(s string) string {
	if s == "" {
		return ""
	}
	return inflection.Singular(s)
}

func ToSnakeCase(s string) string {
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

func ToKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// FrontendTemplateData holds data for rendering frontend templates
type FrontendTemplateData struct {
	Name       string
	Title      string // Capitalized name (e.g., "Product")
	PluralName string // Pluralized name (e.g., "products")
	Fields     []FrontendField
}

// FrontendField represents a field in the frontend entity
type FrontendField struct {
	Name     string
	Type     string
	Label    string
	Required bool
}

// ParseFrontendFields converts CLI field strings (field:type) to FrontendField structs
func ParseFrontendFields(rawFields []string) []FrontendField {
	fields := make([]FrontendField, 0, len(rawFields) )
	for _, f := range rawFields {
		parts := strings.Split(f, ":")
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		typ := parts[1]
		tsType := "string"

		switch typ {
		case "int", "float", "number":
			tsType = "number"
		case "bool", "boolean":
			tsType = "boolean"
		case "time", "date":
			tsType = "string"
		}

		fields = append(fields, FrontendField{
			Name:     name,
			Type:     tsType,
			Label:    Capitalize(name),
			Required: true,
		})
	}
	return fields
}

// BackendField represents a field in the backend entity
type BackendField struct {
	Name         string
	Type         string
	JSONTag      string
	GormTag      string
	SQLType      string
	Relation     string
	RelModule    string
	RelTable     string
	FKColumnName string
}

// ParseBackendFields converts CLI field strings (name:type or name:belongs_to:module) to BackendField structs
func ParseBackendFields(rawFields []string) []BackendField {
	fields := make([]BackendField, 0, len(rawFields))
	for _, f := range rawFields {
		parts := strings.Split(f, ":")
		if len(parts) < 2 {
			continue
		}
		name := Capitalize(parts[0])
		typ := parts[1]
		sqlType := "TEXT"
		goType := "string"

		if len(parts) >= 3 && parts[1] == "belongs_to" {
			relTarget := parts[2]
			singularTarget := inflection.Singular(relTarget)
			relTargetTitle := Capitalize(singularTarget)
			relTable := Pluralize(relTargetTitle)

			// 1. Foreign Key
			fields = append(fields, BackendField{
				Name:    name + "ID",
				Type:    "uint",
				JSONTag: ToSnakeCase(name) + "_id",
				GormTag: ToSnakeCase(name) + "_id",
				SQLType: "INTEGER",
			})

			// 2. Relation Field
			// Use alias for import to avoid conflict with local domain package
			alias := ToSnakeCase(relTarget) + "Domain"

			fields = append(fields, BackendField{
				Name:         name,
				Type:         "*" + alias + "." + relTargetTitle,
				JSONTag:      ToSnakeCase(name) + ",omitempty",
				GormTag:      "foreignKey:" + name + "ID",
				SQLType:      "-",
				Relation:     "belongs_to",
				RelModule:    relTarget,
				RelTable:     relTable,
				FKColumnName: ToSnakeCase(name) + "_id",
			})
			continue
		}

		switch typ {
		case "string":
			goType = "string"
			sqlType = "TEXT"
		case "int":
			goType = "int"
			sqlType = "INTEGER"
		case "bool", "boolean":
			goType = "bool"
			sqlType = "BOOLEAN"
		case "float":
			goType = "float64"
			sqlType = "REAL"
		case "time":
			goType = "time.Time"
			sqlType = "TIMESTAMP"
		}

		fields = append(fields, BackendField{
			Name:    name,
			Type:    goType,
			JSONTag: ToSnakeCase(name),
			GormTag: ToSnakeCase(name),
			SQLType: sqlType,
		})
	}
	return fields
}
