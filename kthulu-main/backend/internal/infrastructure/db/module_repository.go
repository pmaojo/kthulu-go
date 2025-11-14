// @kthulu:module:modules
package db

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"

	"gorm.io/gorm"
)

// ModuleModel represents the database model for modules
type ModuleModel struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null;uniqueIndex"`
	Description  string
	Version      string
	Dependencies string `gorm:"type:text"` // JSON
	Optional     bool   `gorm:"default:false"`
	Category     string
	Tags         string `gorm:"type:text"` // JSON
	Entities     string `gorm:"type:text"` // JSON
	Routes       string `gorm:"type:text"` // JSON
	Migrations   string `gorm:"type:text"` // JSON
	Frontend     bool   `gorm:"default:false"`
	Backend      bool   `gorm:"default:true"`
	Config       string `gorm:"type:text"` // JSON
	Conflicts    string `gorm:"type:text"` // JSON
	MinVersion   string
	MaxVersion   string
	CreatedAt    string `gorm:"type:text"`
	UpdatedAt    string `gorm:"type:text"`
}

// TableName specifies the table name for ModuleModel
func (ModuleModel) TableName() string {
	return "modules"
}

// ToDomain converts ModuleModel to domain.ModuleInfo
func (m *ModuleModel) ToDomain() (*domain.ModuleInfo, error) {
	module := &domain.ModuleInfo{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Version:     m.Version,
		// Dependencies, Tags, etc. would need JSON parsing
		Optional:   m.Optional,
		Category:   m.Category,
		Frontend:   m.Frontend,
		Backend:    m.Backend,
		MinVersion: m.MinVersion,
		MaxVersion: m.MaxVersion,
	}

	// Parse timestamps
	// Similar to project repository

	return module, nil
}

// FromDomain converts domain.ModuleInfo to ModuleModel
func (m *ModuleModel) FromDomain(module *domain.ModuleInfo) {
	m.ID = module.ID
	m.Name = module.Name
	m.Description = module.Description
	m.Version = module.Version
	// Convert slices/maps to JSON strings
	m.Optional = module.Optional
	m.Category = module.Category
	m.Frontend = module.Frontend
	m.Backend = module.Backend
	m.MinVersion = module.MinVersion
	m.MaxVersion = module.MaxVersion
	// Convert timestamps
}

// ModuleRepository provides a database-backed implementation of repository.ModuleRepository.
type ModuleRepository struct {
	db *gorm.DB
}

// NewModuleRepository creates a new instance bound to a Gorm database.
func NewModuleRepository(db *gorm.DB) repository.ModuleRepository {
	return &ModuleRepository{db: db}
}

// Create persists a new module.
func (r *ModuleRepository) Create(ctx context.Context, module *domain.ModuleInfo) error {
	model := &ModuleModel{}
	model.FromDomain(module)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// Update the domain object with the generated ID
	module.ID = model.ID
	return nil
}

// FindByID retrieves a module by ID.
func (r *ModuleRepository) FindByID(ctx context.Context, id uint) (*domain.ModuleInfo, error) {
	var model ModuleModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrModuleNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByName retrieves a module by name.
func (r *ModuleRepository) FindByName(ctx context.Context, name string) (*domain.ModuleInfo, error) {
	var model ModuleModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrModuleNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves module changes.
func (r *ModuleRepository) Update(ctx context.Context, module *domain.ModuleInfo) error {
	model := &ModuleModel{}
	model.FromDomain(module)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a module by ID.
func (r *ModuleRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&ModuleModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrModuleNotFound
	}
	return nil
}

// List retrieves modules with pagination.
func (r *ModuleRepository) List(ctx context.Context, limit, offset int) ([]*domain.ModuleInfo, error) {
	var models []ModuleModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	modules := make([]*domain.ModuleInfo, len(models))
	for i, model := range models {
		module, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		modules[i] = module
	}

	return modules, nil
}

// ListByCategory retrieves modules by category.
func (r *ModuleRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.ModuleInfo, error) {
	var models []ModuleModel
	err := r.db.WithContext(ctx).Where("category = ?", category).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	modules := make([]*domain.ModuleInfo, len(models))
	for i, model := range models {
		module, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		modules[i] = module
	}

	return modules, nil
}

// Count returns the total number of modules.
func (r *ModuleRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ModuleModel{}).Count(&count).Error
	return count, err
}

// FindPaginated returns paginated modules
func (r *ModuleRepository) FindPaginated(ctx context.Context, params repository.PaginationParams) (repository.PaginationResult[*domain.ModuleInfo], error) {
	var models []ModuleModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ModuleModel{}).Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.ModuleInfo]{}, err
	}

	// Build query with pagination
	query := r.db.WithContext(ctx)

	// Add sorting with validation
	allowedSortFields := []string{"id", "name", "category", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "name"
		sortDir = "asc"
	}
	query = query.Order(sortBy + " " + sortDir)

	// Add pagination
	if err := query.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.ModuleInfo]{}, err
	}

	// Convert to domain objects
	modules := make([]*domain.ModuleInfo, len(models))
	for i, model := range models {
		module, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.ModuleInfo]{}, err
		}
		modules[i] = module
	}

	return repository.NewPaginationResult(modules, total, params), nil
}

// SearchPaginated returns paginated modules matching search query
func (r *ModuleRepository) SearchPaginated(ctx context.Context, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.ModuleInfo], error) {
	var models []ModuleModel
	var total int64

	// Build search query
	searchQuery := r.db.WithContext(ctx).Model(&ModuleModel{}).Where("name ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.ModuleInfo]{}, err
	}

	// Add sorting with validation
	searchQuery = searchQuery.Select("*")
	allowedSortFields := []string{"id", "name", "category", "created_at", "updated_at"}
	helper := &PaginationHelper{}
	sortBy := params.SortBy
	sortDir := params.SortDir
	if !helper.isValidSortField(sortBy, allowedSortFields) {
		sortBy = "name"
		sortDir = "asc"
	}
	searchQuery = searchQuery.Order(sortBy + " " + sortDir)

	// Add pagination
	if err := searchQuery.Limit(params.PageSize).Offset(params.CalculateOffset()).Find(&models).Error; err != nil {
		return repository.PaginationResult[*domain.ModuleInfo]{}, err
	}

	// Convert to domain objects
	modules := make([]*domain.ModuleInfo, len(models))
	for i, model := range models {
		module, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.ModuleInfo]{}, err
		}
		modules[i] = module
	}

	return repository.NewPaginationResult(modules, total, params), nil
}

// ExistsByName checks if a module exists with the given name.
func (r *ModuleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ModuleModel{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a module exists with the given ID.
func (r *ModuleRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ModuleModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Ensure ModuleRepository implements the interface
var _ repository.ModuleRepository = (*ModuleRepository)(nil)
