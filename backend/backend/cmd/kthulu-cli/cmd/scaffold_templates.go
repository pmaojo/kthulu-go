package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
)

type Field struct {
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

type moduleTemplateData struct {
	Name          string
	Title         string
	PluralTitle   string
	Fields        []Field
	Imports       []string
	Database      string
	ProjectModule string
	ModuleRelPath string
	RoutePrefix   string
	Protected     bool
}

func newModuleTemplateData(name string, fields []string, database, projectModule, moduleRelPath string) moduleTemplateData {
	parsedFields := parseFields(fields)
	imports := make([]string, 0)
	seenImports := make(map[string]bool)

	for _, f := range parsedFields {
		if f.Relation != "" && f.RelModule != "" {
			// Assuming RelModule is the module name (e.g. "cars")
			// Import path: project/modulePath/relModule/domain
			// Alias: module + Domain
			alias := toSnakeCase(f.RelModule) + "Domain"
			impPath := fmt.Sprintf("%s/%s/%s/domain", projectModule, moduleRelPath, f.RelModule)

			// Formatted import: alias "path"
			imp := fmt.Sprintf(`%s "%s"`, alias, impPath)

			if !seenImports[imp] {
				imports = append(imports, imp)
				seenImports[imp] = true
			}
		}
	}

	return moduleTemplateData{
		Name:          name,
		Title:         exportName(name),
		PluralTitle:   Pluralize(exportName(name)),
		Fields:        parsedFields,
		Imports:       imports,
		Database:      database,
		ProjectModule: projectModule,
		ModuleRelPath: moduleRelPath,
	}
}

// Pluralize using inflection library
func Pluralize(s string) string {
	if s == "" {
		return ""
	}
	return inflection.Plural(s)
}

func parseFields(rawFields []string) []Field {
	fields := make([]Field, 0, len(rawFields))
	for _, f := range rawFields {
		parts := strings.Split(f, ":")
		if len(parts) < 2 {
			continue
		}
		name := exportName(parts[0])
		typ := parts[1]
		sqlType := "TEXT"
		goType := "string"
		if len(parts) >= 3 {
			// Format: name:type:relation (e.g., car:belongs_to:cars)
			// Actually type is implicit or explicit?
			// User said: car:belongs_to:cars
			// Here "car" is field name. "belongs_to" is relation type. "cars" is target?
			// Standard parser expects name:type.
			// If we support relation, we need to handle it.
			// Let's assume if type is 'belongs_to', then it's a relation.
			// Or if parts[1] is a standard type, use it.

			// Let's support: field:type:relation
			// But for relationships like belongs_to, the type is usually uint (FK) + struct.

			if parts[1] == "belongs_to" {
				// Special handling
				// parts[0] = car
				// parts[1] = belongs_to
				// parts[2] = cars (table/module)

				// We need to generate:
				// CarID uint `json:"car_id"`
				// Car *Car `json:"car,omitempty"`

				// So we add two fields? Or one Field struct with extra info?
				// The template ranges over fields.
				// Let's handle it by adding the FK field now, and maybe the relation field too.

				relTarget := parts[2]
				// We expect relTarget to be the module name (e.g. "cars" or "person")
				// We need to derive the Go type (Singular Title) and SQL table (Plural)

				// 1. Go Type: Singularize and TitleCase
				// E.g. "cars" -> "Car", "person" -> "Person"
				singularTarget := inflection.Singular(relTarget)
				relTargetTitle := exportName(singularTarget)

				// 2. SQL Table: Pluralize
				// E.g. "cars" -> "cars", "person" -> "people"
				// Note: TableName logic usually involves TitleCase then Pluralize?
				// The generator uses Pluralize(Capitalize(name)).
				// But SQL table names in migration are usually lowercase?
				// Wait, migration template uses {{.PluralTitle}} which is Pluralize(TitleCase(name)).
				// But typically tables are lowercase in SQL (users, people).
				// The migration template: CREATE TABLE IF NOT EXISTS {{.PluralTitle}}
				// If PluralTitle is "Users", table is "Users".
				// Postgres is case-insensitive (folds to lower), but if quoted it matters.
				// The template does NOT quote.
				// So "Users" -> users.
				// So for FK, we need Pluralize(TitleCase(singularTarget)) to match exactly if we want to be safe,
				// OR just Pluralize(singularTarget) if we assume lowercase.
				// Let's match the generator logic: Pluralize(exportName(singularTarget)).
				relTable := Pluralize(relTargetTitle)

				// 1. Foreign Key
				fields = append(fields, Field{
					Name:    name + "ID",
					Type:    "uint",
					JSONTag: toSnakeCase(name) + "_id",
					GormTag: toSnakeCase(name) + "_id",
					SQLType: "INTEGER",
				})

				// 2. Relation Field
				// Use alias for import to avoid conflict with local domain package
				alias := toSnakeCase(relTarget) + "Domain"

				fields = append(fields, Field{
					Name:         name,
					Type:         "*" + alias + "." + relTargetTitle,
					JSONTag:      toSnakeCase(name) + ",omitempty",
					GormTag:      "foreignKey:" + name + "ID",
					SQLType:      "-",
					Relation:     "belongs_to",
					RelModule:    relTarget,
					RelTable:     relTable,
					FKColumnName: toSnakeCase(name) + "_id",
				})
				continue
			}
		}

		switch typ {
		case "string":
			goType = "string"
			sqlType = "TEXT"
		case "int":
			goType = "int"
			sqlType = "INTEGER"
		case "bool":
			goType = "bool"
			sqlType = "BOOLEAN"
		case "float":
			goType = "float64"
			sqlType = "REAL"
		case "time":
			goType = "time.Time"
			sqlType = "TIMESTAMP"
		}

		fields = append(fields, Field{
			Name:    name,
			Type:    goType,
			JSONTag: toSnakeCase(name),
			GormTag: toSnakeCase(name),
			SQLType: sqlType,
		})
	}
	return fields
}

func renderModuleTemplate(t *template.Template, data moduleTemplateData) string {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

var (
	moduleFileTemplate = template.Must(template.New("moduleFile").Parse(`// @kthulu:module:{{.Name}}
// @kthulu:generated:true
package {{.Name}}

import (
	"go.uber.org/fx"
	"github.com/gorilla/mux"

	"{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/handlers"
	"{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/repository"
	"{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/service"
)

// Providers returns the Fx providers for the {{.Name}} module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        repository.New{{.Title}}Repository,
                        service.New{{.Title}}Service,
                        handlers.New{{.Title}}Handler,
                ),
                fx.Invoke(func(r *mux.Router, h *handlers.{{.Title}}Handler) {
                        h.RegisterRoutes(r)
                }),
        )
}
`))

	domainFileTemplate = template.Must(template.New("domainFile").Parse(`// @kthulu:domain:{{.Name}}
package domain

import (
	"time"
	{{range .Imports}}
	"{{.}}"
	{{end}}
)

// SearchFilter represents search criteria
type SearchFilter struct {
	Query string
	Limit int
	Offset int
}

// {{.Title}} represents a {{.Name}} entity
type {{.Title}} struct {
        ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
        CreatedAt time.Time ` + "`json:\"created_at\"`" + `
        UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `

{{range .Fields}}        {{.Name}} {{.Type}} ` + "`json:\"{{.JSONTag}}\" gorm:\"{{.GormTag}}\"`" + `
{{end}}
}

// TableName overrides the table name used by User to ` + "`{{.PluralTitle}}`" + `
func ({{.Title}}) TableName() string {
	return "{{.PluralTitle}}"
}

// {{.Title}}Repository defines the repository interface
type {{.Title}}Repository interface {
        Create(entity *{{.Title}}) error
        GetByID(id uint) (*{{.Title}}, error)
        Update(entity *{{.Title}}) error
        Delete(id uint) error
        List(filter SearchFilter) ([]*{{.Title}}, error)
}

// {{.Title}}Service defines the service interface
type {{.Title}}Service interface {
        Create{{.Title}}(entity *{{.Title}}) error
        Get{{.Title}}ByID(id uint) (*{{.Title}}, error)
        Update{{.Title}}(entity *{{.Title}}) error
        Delete{{.Title}}(id uint) error
        List{{.Title}}s(filter SearchFilter) ([]*{{.Title}}, error)
}
`))

	repositoryFileTemplate = template.Must(template.New("repositoryFile").Parse(`// @kthulu:repository:{{.Name}}
package repository

import (
        "gorm.io/gorm"

        "{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/domain"
)

type {{.Title}}Repository struct {
        db *gorm.DB
}

func New{{.Title}}Repository(db *gorm.DB) domain.{{.Title}}Repository {
        return &{{.Title}}Repository{db: db}
}

func (r *{{.Title}}Repository) Create(entity *domain.{{.Title}}) error {
        return r.db.Create(entity).Error
}

func (r *{{.Title}}Repository) GetByID(id uint) (*domain.{{.Title}}, error) {
        var entity domain.{{.Title}}
        err := r.db.First(&entity, id).Error
        return &entity, err
}

func (r *{{.Title}}Repository) Update(entity *domain.{{.Title}}) error {
        return r.db.Save(entity).Error
}

func (r *{{.Title}}Repository) Delete(id uint) error {
        return r.db.Delete(&domain.{{.Title}}{}, id).Error
}

func (r *{{.Title}}Repository) List(filter domain.SearchFilter) ([]*domain.{{.Title}}, error) {
        var entities []*domain.{{.Title}}
        query := r.db.Model(&domain.{{.Title}}{})

		if filter.Query != "" {
			// Basic search implementation
			// Note: Adjust fields based on your actual model
			// query = query.Where("name LIKE ?", "%"+filter.Query+"%")
		}

		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}

        err := query.Find(&entities).Error
        return entities, err
}
`))

	serviceFileTemplate = template.Must(template.New("serviceFile").Parse(`// @kthulu:service:{{.Name}}
package service

import (
        "{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/domain"
)

type {{.Title}}Service struct {
        repo domain.{{.Title}}Repository
}

func New{{.Title}}Service(repo domain.{{.Title}}Repository) domain.{{.Title}}Service {
        return &{{.Title}}Service{repo: repo}
}

func (s *{{.Title}}Service) Create{{.Title}}(entity *domain.{{.Title}}) error {
        // Add business logic here
        return s.repo.Create(entity)
}

func (s *{{.Title}}Service) Get{{.Title}}ByID(id uint) (*domain.{{.Title}}, error) {
        return s.repo.GetByID(id)
}

func (s *{{.Title}}Service) Update{{.Title}}(entity *domain.{{.Title}}) error {
        // Add business logic here
        return s.repo.Update(entity)
}

func (s *{{.Title}}Service) Delete{{.Title}}(id uint) error {
        // Add business logic here
        return s.repo.Delete(id)
}

func (s *{{.Title}}Service) List{{.Title}}s(filter domain.SearchFilter) ([]*domain.{{.Title}}, error) {
        return s.repo.List(filter)
}
`))

	handlerFileTemplate = template.Must(template.New("handlerFile").Parse(`// @kthulu:handler:{{.Name}}
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "{{.ProjectModule}}/{{.ModuleRelPath}}/{{.Name}}/domain"
		{{if .Protected}}"{{.ProjectModule}}/internal/infrastructure/middleware"{{end}}
)

type {{.Title}}Handler struct {
        service domain.{{.Title}}Service
}

func New{{.Title}}Handler(service domain.{{.Title}}Service) *{{.Title}}Handler {
        return &{{.Title}}Handler{service: service}
}

func (h *{{.Title}}Handler) RegisterRoutes(r *mux.Router) {
		{{if .RoutePrefix}}sub := r.PathPrefix("{{.RoutePrefix}}").Subrouter(){{else}}sub := r{{end}}
		{{if .Protected}}sub.Use(middleware.AuthMiddleware){{end}}

        sub.HandleFunc("/{{.Name}}s", h.List).Methods("GET")
        sub.HandleFunc("/{{.Name}}s", h.Create).Methods("POST")
        sub.HandleFunc("/{{.Name}}s/{id}", h.GetByID).Methods("GET")
}

// Create handles the creation of a new {{.Name}}
// @Summary      Create a new {{.Name}}
// @Description  Creates a new {{.Name}} with the provided data
// @Tags         {{.Name}}s
// @Accept       json
// @Produce      json
// @Param        input body domain.{{.Title}} true "{{.Title}} object"
// @Success      200  {object}  domain.{{.Title}}
// @Failure      400  {object}  map[string]string "Invalid input"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /{{.Name}}s [post]
func (h *{{.Title}}Handler) Create(w http.ResponseWriter, r *http.Request) {
        var entity domain.{{.Title}}
        if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }

        if err := h.service.Create{{.Title}}(&entity); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// GetByID retrieves a {{.Name}} by its ID
// @Summary      Get {{.Name}}
// @Description  Get a {{.Name}} by its ID
// @Tags         {{.Name}}s
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "{{.Title}} ID"
// @Success      200  {object}  domain.{{.Title}}
// @Failure      400  {object}  map[string]string "Invalid ID"
// @Failure      404  {object}  map[string]string "{{.Title}} not found"
// @Router       /{{.Name}}s/{id} [get]
func (h *{{.Title}}Handler) GetByID(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id, err := strconv.ParseUint(vars["id"], 10, 32)
        if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
        }

        entity, err := h.service.Get{{.Title}}ByID(uint(id))
        if err != nil {
                http.Error(w, err.Error(), http.StatusNotFound)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entity)
}

// List retrieves all {{.Name}}s
// @Summary      List {{.Name}}s
// @Description  Get a list of all {{.Name}}s
// @Tags         {{.Name}}s
// @Accept       json
// @Produce      json
// @Success      200  {array}   domain.{{.Title}}
// @Failure      500  {object}  map[string]string "Internal server error"
// @Router       /{{.Name}}s [get]
func (h *{{.Title}}Handler) List(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		filter := domain.SearchFilter{
			Query: query,
			Limit: limit,
			Offset: offset,
		}

        entities, err := h.service.List{{.Title}}s(filter)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
`))

	migrationFileTemplate = template.Must(template.New("migrationFile").Parse(`-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS {{.PluralTitle}} (
    {{if eq .Database "postgres"}}id SERIAL PRIMARY KEY{{else}}id INTEGER PRIMARY KEY AUTOINCREMENT{{end}},
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP{{range .Fields}}{{if ne .SQLType "-"}},
    {{.JSONTag}} {{.SQLType}}{{end}}{{end}}{{range .Fields}}{{if eq .Relation "belongs_to"}},
    FOREIGN KEY ({{.FKColumnName}}) REFERENCES {{.RelTable}}(id){{end}}{{end}}
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS {{.PluralTitle}};
`))
)

func generateModuleFile(name, projectModule, moduleRelPath string) string {
	data := newModuleTemplateData(name, nil, "", projectModule, moduleRelPath)
	return renderModuleTemplate(moduleFileTemplate, data)
}

func generateDomainFile(name string, fields []string) string {
	data := newModuleTemplateData(name, fields, "", "", "")
	return renderModuleTemplate(domainFileTemplate, data)
}

func generateRepositoryFile(name, projectModule, moduleRelPath string) string {
	data := newModuleTemplateData(name, nil, "", projectModule, moduleRelPath)
	return renderModuleTemplate(repositoryFileTemplate, data)
}

func generateServiceFile(name, projectModule, moduleRelPath string) string {
	data := newModuleTemplateData(name, nil, "", projectModule, moduleRelPath)
	return renderModuleTemplate(serviceFileTemplate, data)
}

func generateHandlerFile(name, projectModule, moduleRelPath string) string {
	data := newModuleTemplateData(name, nil, "", projectModule, moduleRelPath)
	return renderModuleTemplate(handlerFileTemplate, data)
}

func generateHandlerFileWithConfig(name, projectModule, moduleRelPath string, prefix string, protected bool) string {
	data := newModuleTemplateData(name, nil, "", projectModule, moduleRelPath)
	data.RoutePrefix = prefix
	data.Protected = protected
	return renderModuleTemplate(handlerFileTemplate, data)
}

func generateMigrationContent(name string, fields []string, database string) string {
	data := newModuleTemplateData(name, fields, database, "", "")
	return renderModuleTemplate(migrationFileTemplate, data)
}
