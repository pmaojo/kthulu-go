package parser

// EntityDefinition represents the parsed structure of a domain entity
type EntityDefinition struct {
	Name        string            `json:"name"`
	Module      string            `json:"module"` // To be filled by the caller context
	Description string            `json:"description"`
	Fields      []FieldDefinition `json:"fields"`
}

// FieldDefinition represents a single field in the entity
type FieldDefinition struct {
	Name            string   `json:"name"`            // Go Struct Field Name
	JSONName        string   `json:"jsonName"`        // API JSON Name
	Type            string   `json:"type"`            // Go Type
	Label           string   `json:"label"`           // UI Label (derived)
	Required        bool     `json:"required"`        // From validation
	ValidationRules []string `json:"validationRules"` // Raw validation tags
	IsRelation      bool     `json:"isRelation"`
	RelationType    string   `json:"relationType"` // HasMany, BelongsTo
}
