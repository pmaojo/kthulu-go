package generator

import "github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"

// UIEntityDefinition bridges the parsed entity to the UI templates
type UIEntityDefinition struct {
	Name       string             `json:"name"`
	PluralName string             `json:"pluralName"`
	Module     string             `json:"module"`
	Fields     []UIFieldDefinition `json:"fields"`
	Imports    []string           `json:"imports"` // Extra imports needed for the UI
}

type UIFieldDefinition struct {
	Name          string `json:"name"`          // Field Name (e.g. "sku")
	Label         string `json:"label"`         // Display Label (e.g. "SKU")
	ComponentType string `json:"componentType"` // Input, Select, DatePicker, Switch, etc.
	InputType     string `json:"inputType"`     // text, number, email (html attributes)
	Required      bool   `json:"required"`
	IsRelation    bool   `json:"isRelation"`
	TargetEntity  string `json:"targetEntity"`  // If relation, which entity?
	Sortable      bool   `json:"sortable"`
	Filterable    bool   `json:"filterable"`
	ZodType       string `json:"zodType"`       // explicit zod type
}

// MapEntityToUI converts a parser.EntityDefinition to a UIEntityDefinition
func MapEntityToUI(def parser.EntityDefinition) UIEntityDefinition {
	uiDef := UIEntityDefinition{
		Name:       def.Name,
		PluralName: def.Name + "s", // Simplistic pluralization, should use inflection lib
		Module:     def.Module,
	}

	for _, field := range def.Fields {
		uiField := UIFieldDefinition{
			Name:       field.JSONName,
			Label:      field.Name, // Default to Go name, maybe humanize it later
			Required:   contains(field.ValidationRules, "required"),
			IsRelation: field.IsRelation,
			Sortable:   true, // Default to true for now
			Filterable: true,
		}

		if uiField.Name == "" {
			uiField.Name = field.Name // Fallback
		}

		// Map Go Types to UI Components
		switch field.Type {
		case "string":
			uiField.ComponentType = "Input"
			uiField.InputType = "text"
			uiField.ZodType = "string()"
		case "int", "int64", "uint", "float64":
			uiField.ComponentType = "Input"
			uiField.InputType = "number"
			uiField.ZodType = "number()"
		case "bool":
			uiField.ComponentType = "Switch"
			uiField.InputType = "checkbox"
			uiField.ZodType = "boolean()"
		case "time.Time", "*time.Time":
			uiField.ComponentType = "DatePicker"
			uiField.ZodType = "string()" // Fallback to string for Input
		default:
			if field.IsRelation {
				uiField.ComponentType = "Select" // Default for relations
				// Extract target entity from type (e.g. "[]ProductVariant" -> "ProductVariant")
				if len(field.Type) > 2 && field.Type[:2] == "[]" {
					uiField.TargetEntity = field.Type[2:]
					uiField.InputType = "multiselect"
					uiField.ZodType = "array(z.string())" // Usually IDs
				} else if len(field.Type) > 1 && field.Type[0] == '*' {
					uiField.TargetEntity = field.Type[1:]
					uiField.InputType = "select"
					uiField.ZodType = "number()" // Usually ID
				} else {
					uiField.TargetEntity = field.Type
					uiField.InputType = "select"
					uiField.ZodType = "number()"
				}
			} else {
				uiField.ComponentType = "Input"
				uiField.InputType = "text"
				uiField.ZodType = "string()"
			}
		}

		uiDef.Fields = append(uiDef.Fields, uiField)
	}

	return uiDef
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
