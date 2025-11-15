// @kthulu:module:templates
package db

import (
	"context"
	"errors"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"

	"gorm.io/gorm"
)

// TemplateModel represents the database model for templates
type TemplateModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;uniqueIndex"`
	Version     string
	Description string
	Author      string
	Category    string
	Tags        string `gorm:"type:text"` // JSON
	Content     string `gorm:"type:text"` // JSON
	Remote      bool   `gorm:"default:false"`
	URL         string
	CreatedAt   string `gorm:"type:text"`
	UpdatedAt   string `gorm:"type:text"`
}

// TableName specifies the table name for TemplateModel
func (TemplateModel) TableName() string {
	return "templates"
}

// ToDomain converts TemplateModel to domain.Template
func (t *TemplateModel) ToDomain() (*domain.Template, error) {
	template := &domain.Template{
		ID:          t.ID,
		Name:        t.Name,
		Version:     t.Version,
		Description: t.Description,
		Author:      t.Author,
		Category:    t.Category,
		Remote:      t.Remote,
		URL:         t.URL,
	}

	// Parse timestamps
	// Similar to project repository

	return template, nil
}

// FromDomain converts domain.Template to TemplateModel
func (t *TemplateModel) FromDomain(template *domain.Template) {
	t.ID = template.ID
	t.Name = template.Name
	t.Version = template.Version
	t.Description = template.Description
	t.Author = template.Author
	t.Category = template.Category
	// Convert Tags and Content to JSON strings
	t.Remote = template.Remote
	t.URL = template.URL
	// Convert timestamps
}

// TemplateRepository provides a database-backed implementation of repository.TemplateRepository.
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new instance bound to a Gorm database.
func NewTemplateRepository(db *gorm.DB) repository.TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create persists a new template.
func (r *TemplateRepository) Create(ctx context.Context, template *domain.Template) error {
	model := &TemplateModel{}
	model.FromDomain(template)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	template.ID = model.ID
	return nil
}

// FindByID retrieves a template by ID.
func (r *TemplateRepository) FindByID(ctx context.Context, id uint) (*domain.Template, error) {
	var model TemplateModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByName retrieves a template by name.
func (r *TemplateRepository) FindByName(ctx context.Context, name string) (*domain.Template, error) {
	var model TemplateModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves template changes.
func (r *TemplateRepository) Update(ctx context.Context, template *domain.Template) error {
	model := &TemplateModel{}
	model.FromDomain(template)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a template by ID.
func (r *TemplateRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&TemplateModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("template not found")
	}
	return nil
}

// List retrieves templates with pagination.
func (r *TemplateRepository) List(ctx context.Context, limit, offset int) ([]*domain.Template, error) {
	var models []TemplateModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	templates := make([]*domain.Template, len(models))
	for i, model := range models {
		template, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}

	return templates, nil
}

// ListByCategory retrieves templates by category.
func (r *TemplateRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Template, error) {
	var models []TemplateModel
	err := r.db.WithContext(ctx).Where("category = ?", category).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	templates := make([]*domain.Template, len(models))
	for i, model := range models {
		template, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}

	return templates, nil
}

// Count returns the total number of templates.
func (r *TemplateRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateModel{}).Count(&count).Error
	return count, err
}

// FindPaginated returns paginated templates
func (r *TemplateRepository) FindPaginated(ctx context.Context, params repository.PaginationParams) (repository.PaginationResult[*domain.Template], error) {
	var models []TemplateModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&TemplateModel{}).Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.Template]{}, err
	}

	query := r.db.WithContext(ctx)

	allowedSortFields := []string{"id", "name", "category", "author", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "name"
		sortDir = "asc"
	}
	query = query.Order(sortBy + " " + sortDir)

	if err := query.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.Template]{}, err
	}

	templates := make([]*domain.Template, len(models))
	for i, model := range models {
		template, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.Template]{}, err
		}
		templates[i] = template
	}

	return repository.NewPaginationResult(templates, total, params), nil
}

// SearchPaginated returns paginated templates matching search query
func (r *TemplateRepository) SearchPaginated(ctx context.Context, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.Template], error) {
	var models []TemplateModel
	var total int64

	searchQuery := r.db.WithContext(ctx).Model(&TemplateModel{}).Where("name ILIKE ? OR description ILIKE ? OR author ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.Template]{}, err
	}

	searchQuery = searchQuery.Select("*")
	allowedSortFields := []string{"id", "name", "category", "author", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "name"
		sortDir = "asc"
	}
	searchQuery = searchQuery.Order(sortBy + " " + sortDir)

	if err := searchQuery.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.Template]{}, err
	}

	templates := make([]*domain.Template, len(models))
	for i, model := range models {
		template, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.Template]{}, err
		}
		templates[i] = template
	}

	return repository.NewPaginationResult(templates, total, params), nil
}

// ExistsByName checks if a template exists with the given name.
func (r *TemplateRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a template exists with the given ID.
func (r *TemplateRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TemplateModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Ensure TemplateRepository implements the interface
var _ repository.TemplateRepository = (*TemplateRepository)(nil)
