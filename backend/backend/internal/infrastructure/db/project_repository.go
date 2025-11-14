// @kthulu:module:projects
package db

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

// ProjectModel represents the database model for projects
type ProjectModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;uniqueIndex"`
	Modules     string `gorm:"type:text"` // JSON
	Template    string
	Database    string
	Frontend    string
	SkipGit     bool `gorm:"default:false"`
	SkipDocker  bool `gorm:"default:false"`
	Author      string
	License     string
	Description string
	Path        string
	DryRun      bool   `gorm:"default:false"`
	CreatedAt   string `gorm:"type:text"`
	UpdatedAt   string `gorm:"type:text"`
}

// TableName specifies the table name for ProjectModel
func (ProjectModel) TableName() string {
	return "projects"
}

// ToDomain converts ProjectModel to domain.Project
func (p *ProjectModel) ToDomain() (*domain.Project, error) {
	project := &domain.Project{
		ID:          p.ID,
		Name:        p.Name,
		Modules:     []string{}, // Will be parsed from JSON
		Template:    p.Template,
		Database:    p.Database,
		Frontend:    p.Frontend,
		SkipGit:     p.SkipGit,
		SkipDocker:  p.SkipDocker,
		Author:      p.Author,
		License:     p.License,
		Description: p.Description,
		Path:        p.Path,
		DryRun:      p.DryRun,
	}

	// Parse JSON fields if needed
	if p.Modules != "" {
		// Simple parsing - in real implementation, use json.Unmarshal
		project.Modules = []string{p.Modules} // Placeholder
	}

	// Parse timestamps
	if p.CreatedAt != "" {
		// Parse timestamp
	}
	if p.UpdatedAt != "" {
		// Parse timestamp
	}

	return project, nil
}

// FromDomain converts domain.Project to ProjectModel
func (p *ProjectModel) FromDomain(project *domain.Project) {
	p.ID = project.ID
	p.Name = project.Name
	// Convert modules to JSON string
	p.Modules = "" // Placeholder - should serialize to JSON
	p.Template = project.Template
	p.Database = project.Database
	p.Frontend = project.Frontend
	p.SkipGit = project.SkipGit
	p.SkipDocker = project.SkipDocker
	p.Author = project.Author
	p.License = project.License
	p.Description = project.Description
	p.Path = project.Path
	p.DryRun = project.DryRun
	// Convert timestamps to strings
	p.CreatedAt = project.CreatedAt.Format("2006-01-02 15:04:05")
	p.UpdatedAt = project.UpdatedAt.Format("2006-01-02 15:04:05")
}

// ProjectRepository provides a database-backed implementation of repository.ProjectRepository.
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new instance bound to a Gorm database.
func NewProjectRepository(db *gorm.DB) repository.ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create persists a new project.
func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	model := &ProjectModel{}
	model.FromDomain(project)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// Update the domain object with the generated ID
	project.ID = model.ID
	return nil
}

// FindByID retrieves a project by ID.
func (r *ProjectRepository) FindByID(ctx context.Context, id uint) (*domain.Project, error) {
	var model ProjectModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByName retrieves a project by name.
func (r *ProjectRepository) FindByName(ctx context.Context, name string) (*domain.Project, error) {
	var model ProjectModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves project changes.
func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	model := &ProjectModel{}
	model.FromDomain(project)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a project by ID.
func (r *ProjectRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&ProjectModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

// List retrieves projects with pagination.
func (r *ProjectRepository) List(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		project, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		projects[i] = project
	}

	return projects, nil
}

// Count returns the total number of projects.
func (r *ProjectRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProjectModel{}).Count(&count).Error
	return count, err
}

// FindPaginated returns paginated projects
func (r *ProjectRepository) FindPaginated(ctx context.Context, params repository.PaginationParams) (repository.PaginationResult[*domain.Project], error) {
	var models []ProjectModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ProjectModel{}).Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.Project]{}, err
	}

	// Build query with pagination
	query := r.db.WithContext(ctx)

	// Add sorting with validation
	allowedSortFields := []string{"id", "name", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "created_at"
		sortDir = "desc"
	}
	query = query.Order(sortBy + " " + sortDir)

	// Add pagination
	if err := query.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.Project]{}, err
	}

	// Convert to domain objects
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		project, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.Project]{}, err
		}
		projects[i] = project
	}

	return repository.NewPaginationResult(projects, total, params), nil
}

// SearchPaginated returns paginated projects matching search query
func (r *ProjectRepository) SearchPaginated(ctx context.Context, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.Project], error) {
	var models []ProjectModel
	var total int64

	// Build search query
	searchQuery := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("name ILIKE ?", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.Project]{}, err
	}

	// Add sorting with validation
	searchQuery = searchQuery.Select("*")
	allowedSortFields := []string{"id", "name", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "created_at"
		sortDir = "desc"
	}
	searchQuery = searchQuery.Order(sortBy + " " + sortDir)

	// Add pagination
	if err := searchQuery.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.Project]{}, err
	}

	// Convert to domain objects
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		project, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.Project]{}, err
		}
		projects[i] = project
	}

	return repository.NewPaginationResult(projects, total, params), nil
}

// ExistsByName checks if a project exists with the given name.
func (r *ProjectRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a project exists with the given ID.
func (r *ProjectRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Ensure ProjectRepository implements the interface
var _ repository.ProjectRepository = (*ProjectRepository)(nil)
