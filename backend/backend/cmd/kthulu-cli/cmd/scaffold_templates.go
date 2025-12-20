package cmd

import (
	"bytes"
	"strings"
	"text/template"
)

type Field struct {
	Name    string
	Type    string
	JSONTag string
	GormTag string
	SQLType string
}

type moduleTemplateData struct {
	Name        string
	Title       string
	Fields      []Field
	Database    string
	ProjectModule string
}

func newModuleTemplateData(name string, fields []string, database, projectModule string) moduleTemplateData {
	return moduleTemplateData{
		Name:          name,
		Title:         exportName(name),
		Fields:        parseFields(fields),
		Database:      database,
		ProjectModule: projectModule,
	}
}

func parseFields(rawFields []string) []Field {
	fields := make([]Field, 0, len(rawFields))
	for _, f := range rawFields {
		parts := strings.Split(f, ":")
		if len(parts) != 2 {
			continue
		}
		name := exportName(parts[0])
		typ := parts[1]
		sqlType := "TEXT"
		goType := "string"

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
			JSONTag: strings.ToLower(name),
			GormTag: strings.ToLower(name),
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

import "go.uber.org/fx"

// Providers returns the Fx providers for the {{.Name}} module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        New{{.Title}}Repository,
                        New{{.Title}}Service,
                        New{{.Title}}Handler,
                ),
        )
}
`))

	domainFileTemplate = template.Must(template.New("domainFile").Parse(`// @kthulu:domain:{{.Name}}
package domain

import "time"

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

// TableName overrides the table name used by User to ` + "`{{.Name}}s`" + `
func ({{.Title}}) TableName() string {
	return "{{.Name}}s"
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

        "{{.ProjectModule}}/internal/adapters/http/modules/{{.Name}}/domain"
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
        "{{.ProjectModule}}/internal/adapters/http/modules/{{.Name}}/domain"
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
        "{{.ProjectModule}}/internal/adapters/http/modules/{{.Name}}/domain"
)

type {{.Title}}Handler struct {
        service domain.{{.Title}}Service
}

func New{{.Title}}Handler(service domain.{{.Title}}Service) *{{.Title}}Handler {
        return &{{.Title}}Handler{service: service}
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
CREATE TABLE IF NOT EXISTS {{.Name}}s (
    {{if eq .Database "postgres"}}id SERIAL PRIMARY KEY{{else}}id INTEGER PRIMARY KEY AUTOINCREMENT{{end}},
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP{{range .Fields}},
    {{.JSONTag}} {{.SQLType}}{{end}}
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS {{.Name}}s;
`))
)

func generateModuleFile(name string) string {
	data := newModuleTemplateData(name, nil, "", "")
	return renderModuleTemplate(moduleFileTemplate, data)
}

func generateDomainFile(name string, fields []string) string {
	data := newModuleTemplateData(name, fields, "", "")
	return renderModuleTemplate(domainFileTemplate, data)
}

func generateRepositoryFile(name, projectModule string) string {
	data := newModuleTemplateData(name, nil, "", projectModule)
	return renderModuleTemplate(repositoryFileTemplate, data)
}

func generateServiceFile(name, projectModule string) string {
	data := newModuleTemplateData(name, nil, "", projectModule)
	return renderModuleTemplate(serviceFileTemplate, data)
}

func generateHandlerFile(name, projectModule string) string {
	data := newModuleTemplateData(name, nil, "", projectModule)
	return renderModuleTemplate(handlerFileTemplate, data)
}

func generateMigrationContent(name string, fields []string, database string) string {
	data := newModuleTemplateData(name, fields, database, "")
	return renderModuleTemplate(migrationFileTemplate, data)
}
