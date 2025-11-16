package cmd

import (
	"bytes"
	"text/template"
)

type moduleTemplateData struct {
	Name  string
	Title string
}

func newModuleTemplateData(name string) moduleTemplateData {
	return moduleTemplateData{
		Name:  name,
		Title: exportName(name),
	}
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

// {{.Title}} represents a {{.Name}} entity
type {{.Title}} struct {
        ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
        CreatedAt time.Time ` + "`json:\"created_at\"`" + `
        UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `

        // Add your fields here
}

// {{.Title}}Repository defines the repository interface
type {{.Title}}Repository interface {
        Create(entity *{{.Title}}) error
        GetByID(id uint) (*{{.Title}}, error)
        Update(entity *{{.Title}}) error
        Delete(id uint) error
        List() ([]*{{.Title}}, error)
}

// {{.Title}}Service defines the service interface
type {{.Title}}Service interface {
        Create{{.Title}}(entity *{{.Title}}) error
        Get{{.Title}}ByID(id uint) (*{{.Title}}, error)
        Update{{.Title}}(entity *{{.Title}}) error
        Delete{{.Title}}(id uint) error
        List{{.Title}}s() ([]*{{.Title}}, error)
}
`))

	repositoryFileTemplate = template.Must(template.New("repositoryFile").Parse(`// @kthulu:repository:{{.Name}}
package repository

import (
        "gorm.io/gorm"

        "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/{{.Name}}/domain"
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

func (r *{{.Title}}Repository) List() ([]*domain.{{.Title}}, error) {
        var entities []*domain.{{.Title}}
        err := r.db.Find(&entities).Error
        return entities, err
}
`))

	serviceFileTemplate = template.Must(template.New("serviceFile").Parse(`// @kthulu:service:{{.Name}}
package service

import (
        "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/{{.Name}}/domain"
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

func (s *{{.Title}}Service) List{{.Title}}s() ([]*domain.{{.Title}}, error) {
        return s.repo.List()
}
`))

	handlerFileTemplate = template.Must(template.New("handlerFile").Parse(`// @kthulu:handler:{{.Name}}
package handlers

import (
        "encoding/json"
        "net/http"
        "strconv"

        "github.com/gorilla/mux"
        "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/{{.Name}}/domain"
)

type {{.Title}}Handler struct {
        service domain.{{.Title}}Service
}

func New{{.Title}}Handler(service domain.{{.Title}}Service) *{{.Title}}Handler {
        return &{{.Title}}Handler{service: service}
}

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

func (h *{{.Title}}Handler) List(w http.ResponseWriter, r *http.Request) {
        entities, err := h.service.List{{.Title}}s()
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(entities)
}
`))
)

func generateModuleFile(name string) string {
	data := newModuleTemplateData(name)
	return renderModuleTemplate(moduleFileTemplate, data)
}

func generateDomainFile(name string) string {
	data := newModuleTemplateData(name)
	return renderModuleTemplate(domainFileTemplate, data)
}

func generateRepositoryFile(name string) string {
	data := newModuleTemplateData(name)
	return renderModuleTemplate(repositoryFileTemplate, data)
}

func generateServiceFile(name string) string {
	data := newModuleTemplateData(name)
	return renderModuleTemplate(serviceFileTemplate, data)
}

func generateHandlerFile(name string) string {
	data := newModuleTemplateData(name)
	return renderModuleTemplate(handlerFileTemplate, data)
}
