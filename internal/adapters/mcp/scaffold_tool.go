package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"gopkg.in/yaml.v3"

	"github.com/pmaojo/kthulu-go/internal/blueprint"
)

// ScaffoldModule describes one domain entity to generate.
type ScaffoldModule struct {
	Name   string   `json:"name" jsonschema:"required,description=Entity/module name in singular (e.g. tournament, team, player)."`
	Fields []string `json:"fields" jsonschema:"required,description=REQUIRED field definitions using name:type[:rules] syntax. Types: string int float bool time. Validation rules (comma-separated): required min=N max=N email oneof=a|b|c. Relations: name:belongs_to:module (generates the foreign key and typed association). Example: [\"title:string:required,min=2\" \"max_teams:int:min=2\" \"starts_at:time\" \"winner:belongs_to:team\"]."`
}

// ScaffoldProjectArgs are the arguments for the scaffold_project tool.
type ScaffoldProjectArgs struct {
	Name       string           `json:"name" jsonschema:"required,description=Project name. A directory with this name is created under the working directory."`
	Modules    []ScaffoldModule `json:"modules" jsonschema:"required,description=The domain entities of the application WITH their fields. Model the real domain: every entity needs its actual fields and relations — an empty or missing field list is rejected."`
	Database   string           `json:"database,omitempty" jsonschema:"description=Database: sqlite (default), postgres or mysql."`
	Features   []string         `json:"features,omitempty" jsonschema:"description=Extra features: auth and user (recommended) plus infrastructure runtimes: queues (background jobs + scheduler), mail, storage, cache, events, session, policy, rate, i18n, seeder, validate."`
	ModulePath string           `json:"module_path,omitempty" jsonschema:"description=Go module path for go.mod (e.g. github.com/acme/tournaments). Defaults to the project name."`
	Auth       string           `json:"auth,omitempty" jsonschema:"description=Auth type: jwt (default), oauth or both."`
}

// ScaffoldProjectTool returns a structured project scaffolding tool. Unlike
// the raw 'create' command it takes the domain model as structured data, so
// agents declare every entity's fields instead of falling back to defaults.
func ScaffoldProjectTool(executor CommandExecutor, workingDir string) RegisteredTool {
	return RegisteredTool{
		Name: "scaffold_project",
		Description: "PREFERRED way to create a new application. Generates a full project (backend modules, validation, admin UI, REST API) from a structured domain model. " +
			"Declare EVERY entity with its real fields and relations in 'modules' — this is what makes the generated app complete. " +
			"After it succeeds, call workdir_set with the new project directory, then finish setup with the printed go commands.",
		Handler: func(ctx context.Context, args ScaffoldProjectArgs) (*mcp_golang.ToolResponse, error) {
			if err := validateScaffoldArgs(args); err != nil {
				return nil, err
			}

			dir := resolveWorkdir(workingDir)
			planPath, err := writeScaffoldPlan(dir, args)
			if err != nil {
				return nil, err
			}

			cmdArgs := []string{"create", args.Name, "--from-plan", planPath}
			if args.ModulePath != "" {
				cmdArgs = append(cmdArgs, "--module-path", args.ModulePath)
			}

			response, err := runCreateCLI(ctx, executor, dir, workingDir, args.Name, "scaffold", cmdArgs)
			if err != nil {
				return nil, err
			}
			response += "\n\nNEXT STEPS:\n1. Finish setup (run via shell_execute):\n   go run github.com/a-h/templ/cmd/templ@v0.3.977 generate ./...\n   go mod tidy\n2. Verify with go_build, then go_test"
			if findings := reviewDomainModel(args.Modules); len(findings) > 0 {
				response += "\n\nDOMAIN MODEL REVIEW (optional improvements — apply with add_module or by editing entities):\n- " + strings.Join(findings, "\n- ")
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(response)), nil
		},
	}
}

func validateScaffoldArgs(args ScaffoldProjectArgs) error {
	if strings.TrimSpace(args.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(args.Modules) == 0 {
		return fmt.Errorf("modules is required: declare every domain entity with its fields, e.g. {\"name\":\"tournament\",\"fields\":[\"title:string:required\",\"starts_at:time\"]}")
	}
	for _, m := range args.Modules {
		if strings.TrimSpace(m.Name) == "" {
			return fmt.Errorf("every module needs a name")
		}
		if len(m.Fields) == 0 {
			return fmt.Errorf("module %q has no fields — declare its real domain fields (name:type[:rules]), e.g. [\"title:string:required\", \"starts_at:time\", \"winner:belongs_to:team\"]", m.Name)
		}
		for _, f := range m.Fields {
			if !strings.Contains(f, ":") {
				return fmt.Errorf("module %q field %q is missing a type — use name:type[:rules], e.g. \"score:int:min=0\"", m.Name, f)
			}
		}
	}
	return nil
}

// writeScaffoldPlan renders the blueprint YAML used by create --from-plan.
func writeScaffoldPlan(dir string, args ScaffoldProjectArgs) (string, error) {
	database := args.Database
	if database == "" {
		database = "sqlite"
	}
	auth := args.Auth
	if auth == "" {
		auth = "jwt"
	}

	bp := blueprint.ProjectBlueprint{
		Name:     args.Name,
		Template: "microservice",
		Database: database,
		Frontend: "templ",
		Auth:     auth,
		Features: args.Features,
		Modules:  map[string]blueprint.ModuleConfig{},
	}
	for _, m := range args.Modules {
		bp.Modules[m.Name] = blueprint.ModuleConfig{Fields: m.Fields}
	}

	data, err := yaml.Marshal(&bp)
	if err != nil {
		return "", fmt.Errorf("encode plan: %w", err)
	}
	planPath := filepath.Join(dir, args.Name+"-plan.yaml")
	if err := os.WriteFile(planPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write plan: %w", err)
	}
	return planPath, nil
}
