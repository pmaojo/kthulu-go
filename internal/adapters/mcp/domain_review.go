package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/jinzhu/inflection"
	mcp_golang "github.com/metoro-io/mcp-golang"
)

// ReviewDomainModelArgs are the arguments for the review_domain_model tool.
type ReviewDomainModelArgs struct {
	Modules []ScaffoldModule `json:"modules" jsonschema:"required,description=The proposed domain model: entities with their fields in name:type[:rules] syntax."`
}

// ReviewDomainModelTool returns a deterministic domain-model reviewer.
// Agents call it before scaffold_project to harden their model; the same
// review runs automatically inside scaffold_project.
func ReviewDomainModelTool() RegisteredTool {
	return RegisteredTool{
		Name: "review_domain_model",
		Description: "Review a proposed domain model BEFORE scaffolding it. Returns concrete improvement suggestions: " +
			"missing relations between entities, enum-like fields without oneof rules, timestamps not typed as time, " +
			"emails without validation, plural entity names, and missing required rules. Iterate until the review is clean, then call scaffold_project.",
		Handler: func(ctx context.Context, args ReviewDomainModelArgs) (*mcp_golang.ToolResponse, error) {
			if len(args.Modules) == 0 {
				return nil, fmt.Errorf("modules is required")
			}
			findings := reviewDomainModel(args.Modules)
			text := "✅ Domain model looks solid. Proceed with scaffold_project."
			if len(findings) > 0 {
				text = fmt.Sprintf("Domain model review (%d suggestion(s)):\n- %s", len(findings), strings.Join(findings, "\n- "))
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(text)), nil
		},
	}
}

type parsedField struct {
	name  string
	typ   string
	rules string
}

func parseFieldSpec(spec string) parsedField {
	parts := strings.SplitN(spec, ":", 3)
	f := parsedField{name: strings.ToLower(parts[0])}
	if len(parts) > 1 {
		f.typ = parts[1]
	}
	if len(parts) > 2 {
		f.rules = strings.ToLower(parts[2])
	}
	return f
}

// reviewDomainModel applies heuristics that catch the most common weak-agent
// modeling mistakes. Findings are suggestions, never blockers.
func reviewDomainModel(modules []ScaffoldModule) []string {
	var findings []string

	moduleNames := map[string]bool{}
	for _, m := range modules {
		moduleNames[strings.ToLower(m.Name)] = true
	}

	hasAnyRelation := false
	for _, m := range modules {
		name := strings.ToLower(m.Name)

		if inflection.Singular(name) != name {
			findings = append(findings, fmt.Sprintf("module %q: use the singular form %q — tables and pluralized names are derived automatically", m.Name, inflection.Singular(name)))
		}

		hasRequired := false
		for _, spec := range m.Fields {
			f := parseFieldSpec(spec)
			if f.typ == "belongs_to" {
				hasAnyRelation = true
				continue
			}
			if strings.Contains(f.rules, "required") {
				hasRequired = true
			}
			findings = append(findings, reviewField(m.Name, f, moduleNames)...)
		}
		if len(m.Fields) > 0 && !hasRequired {
			findings = append(findings, fmt.Sprintf("module %q: no field is marked required — mark the identifying field (e.g. its title or name) as required", m.Name))
		}
	}

	if len(modules) > 1 && !hasAnyRelation {
		findings = append(findings, "no relations between entities — real domains are connected; add belongs_to fields (e.g. player has squad:belongs_to:team)")
	}

	return findings
}

func reviewField(module string, f parsedField, moduleNames map[string]bool) []string {
	var findings []string

	switch {
	case (f.name == "status" || f.name == "state" || f.name == "kind" || f.name == "type") && !strings.Contains(f.rules, "oneof"):
		findings = append(findings, fmt.Sprintf("module %q field %q: looks like an enum — constrain it with oneof=a|b|c so invalid values are rejected", module, f.name))
	case strings.Contains(f.name, "email") && !strings.Contains(f.rules, "email"):
		findings = append(findings, fmt.Sprintf("module %q field %q: add the email validation rule", module, f.name))
	case (strings.HasSuffix(f.name, "_at") || strings.HasSuffix(f.name, "_date") || f.name == "date") && f.typ != "time":
		findings = append(findings, fmt.Sprintf("module %q field %q: timestamps should use the time type, not %s", module, f.name, f.typ))
	case strings.HasSuffix(f.name, "_id") && f.typ != "belongs_to":
		ref := strings.TrimSuffix(f.name, "_id")
		if moduleNames[ref] || moduleNames[inflection.Singular(ref)] || moduleNames[inflection.Plural(ref)] {
			findings = append(findings, fmt.Sprintf("module %q field %q: model this as a relation (%s:belongs_to:%s) to get the typed association and foreign key", module, f.name, ref, ref))
		}
	}

	return findings
}
