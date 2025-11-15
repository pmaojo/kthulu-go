// @kthulu:module:auth
package db

import (
	"context"
	"errors"
	"time"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"

	"gorm.io/gorm"
)

// UserModel represents the database model for users
type UserModel struct {
	ID               uint   `gorm:"primaryKey"`
	Email            string `gorm:"uniqueIndex;not null"`
	PasswordHash     string `gorm:"not null"`
	ConfirmedAt      *time.Time
	ConfirmationCode string
	RoleID           uint `gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// Associations
	Role *RoleModel `gorm:"foreignKey:RoleID"`
}

// TableName specifies the table name for UserModel
func (UserModel) TableName() string {
	return "users"
}

// ToDomain converts UserModel to domain.User
func (u *UserModel) ToDomain() (*domain.User, error) {
	email, err := domain.NewEmail(u.Email)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:               u.ID,
		Email:            email,
		PasswordHash:     u.PasswordHash,
		ConfirmedAt:      u.ConfirmedAt,
		ConfirmationCode: u.ConfirmationCode,
		RoleID:           u.RoleID,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}

	if u.Role != nil {
		role, err := u.Role.ToDomain()
		if err != nil {
			return nil, err
		}
		user.Role = role
	}

	return user, nil
}

// FromDomain converts domain.User to UserModel
func (u *UserModel) FromDomain(user *domain.User) {
	u.ID = user.ID
	u.Email = user.Email.String()
	u.PasswordHash = user.PasswordHash
	u.ConfirmedAt = user.ConfirmedAt
	u.ConfirmationCode = user.ConfirmationCode
	u.RoleID = user.RoleID
	u.CreatedAt = user.CreatedAt
	u.UpdatedAt = user.UpdatedAt
}

// UserRepository provides a database-backed implementation of repository.UserRepository.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance bound to a Gorm database.
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

// Create persists a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	model := &UserModel{}
	model.FromDomain(user)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	// Update the domain object with the generated ID
	user.ID = model.ID
	return nil
}

// FindByID retrieves a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// FindByEmail retrieves a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return model.ToDomain()
}

// Update saves user changes.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	model := &UserModel{}
	model.FromDomain(user)

	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// List retrieves users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var models []UserModel
	err := r.db.WithContext(ctx).Preload("Role").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		user, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&count).Error
	return count, err
}

// FindByRole retrieves users by role ID.
func (r *UserRepository) FindByRole(ctx context.Context, roleID uint) ([]*domain.User, error) {
	var models []UserModel
	err := r.db.WithContext(ctx).Preload("Role").Where("role_id = ?", roleID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		user, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

// FindUnconfirmed retrieves unconfirmed users older than the specified time.
func (r *UserRepository) FindUnconfirmed(ctx context.Context, olderThan *time.Time) ([]*domain.User, error) {
	query := r.db.WithContext(ctx).Preload("Role").Where("confirmed_at IS NULL")

	if olderThan != nil {
		query = query.Where("created_at < ?", *olderThan)
	}

	var models []UserModel
	err := query.Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		user, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

// ExistsByEmail checks if a user exists with the given email.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByID checks if a user exists with the given ID.
func (r *UserRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// FindPaginated returns paginated users
func (r *UserRepository) FindPaginated(ctx context.Context, params repository.PaginationParams) (repository.PaginationResult[*domain.User], error) {
	var models []UserModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.User]{}, err
	}

	// Build query with pagination
	query := r.db.WithContext(ctx).Preload("Role")

	// Add sorting with validation
	allowedSortFields := []string{"id", "email", "created_at", "updated_at"}
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
		return repository.PaginationResult[*domain.User]{}, err
	}

	// Convert to domain objects
	users := make([]*domain.User, len(models))
	for i, model := range models {
		user, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.User]{}, err
		}
		users[i] = user
	}

	return repository.NewPaginationResult(users, total, params), nil
}

// SearchPaginated returns paginated users matching search query
func (r *UserRepository) SearchPaginated(ctx context.Context, query string, params repository.PaginationParams) (repository.PaginationResult[*domain.User], error) {
	var models []UserModel
	var total int64

	// Build search query
	searchQuery := r.db.WithContext(ctx).Model(&UserModel{}).Where("email ILIKE ?", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return repository.PaginationResult[*domain.User]{}, err
	}

	// Add preloading and sorting with validation
	searchQuery = searchQuery.Preload("Role")
	allowedSortFields := []string{"id", "email", "created_at", "updated_at"}
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
		return repository.PaginationResult[*domain.User]{}, err
	}

	// Convert to domain objects
	users := make([]*domain.User, len(models))
	for i, model := range models {
		user, err := model.ToDomain()
		if err != nil {
			return repository.PaginationResult[*domain.User]{}, err
		}
		users[i] = user
	}

	return repository.NewPaginationResult(users, total, params), nil
}

// Ensure UserRepository implements the interface
var _ repository.UserRepository = (*UserRepository)(nil)
