// @kthulu:module:org
package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// organizationModel represents the database model for organizations
type organizationModel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;size:100"`
	Slug        string `gorm:"uniqueIndex;not null;size:50"`
	Description string `gorm:"size:500"`
	Type        string `gorm:"not null;size:20"`
	Domain      string `gorm:"uniqueIndex;size:100"`
	LogoURL     string `gorm:"size:500"`
	Website     string `gorm:"size:500"`
	Phone       string `gorm:"size:20"`
	Address     string `gorm:"size:200"`
	City        string `gorm:"size:100"`
	State       string `gorm:"size:100"`
	Country     string `gorm:"size:100"`
	PostalCode  string `gorm:"size:20"`
	IsActive    bool   `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relationships
	Users []organizationUserModel `gorm:"foreignKey:OrganizationID"`
}

func (organizationModel) TableName() string {
	return "organizations"
}

// organizationUserModel represents the database model for organization users
type organizationUserModel struct {
	ID             uint   `gorm:"primaryKey"`
	OrganizationID uint   `gorm:"not null;index"`
	UserID         uint   `gorm:"not null;index"`
	Role           string `gorm:"not null;size:20"`
	JoinedAt       time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relationships
	Organization organizationModel `gorm:"foreignKey:OrganizationID"`
	User         UserModel         `gorm:"foreignKey:UserID"`
}

func (organizationUserModel) TableName() string {
	return "organization_users"
}

// invitationModel represents the database model for invitations
type invitationModel struct {
	ID             uint      `gorm:"primaryKey"`
	OrganizationID uint      `gorm:"not null;index"`
	InviterID      uint      `gorm:"not null;index"`
	Email          string    `gorm:"not null;size:255;index"`
	Role           string    `gorm:"not null;size:20"`
	Token          string    `gorm:"uniqueIndex;not null;size:255"`
	Status         string    `gorm:"not null;size:20;index"`
	Message        string    `gorm:"size:500"`
	ExpiresAt      time.Time `gorm:"not null;index"`
	AcceptedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relationships
	Organization organizationModel `gorm:"foreignKey:OrganizationID"`
	Inviter      UserModel         `gorm:"foreignKey:InviterID"`
}

func (invitationModel) TableName() string {
	return "invitations"
}

// OrganizationRepository implements repository.OrganizationRepository
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new OrganizationRepository
func NewOrganizationRepository(db *gorm.DB) repository.OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	model := r.domainToModel(org)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrOrganizationAlreadyExists
		}
		return err
	}

	org.ID = model.ID
	org.CreatedAt = model.CreatedAt
	org.UpdatedAt = model.UpdatedAt

	return nil
}

// FindByID finds an organization by ID
func (r *OrganizationRepository) FindByID(ctx context.Context, id uint) (*domain.Organization, error) {
	var model organizationModel

	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// FindBySlug finds an organization by slug
func (r *OrganizationRepository) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	var model organizationModel

	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// Update updates an organization
func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	model := r.domainToModel(org)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrOrganizationAlreadyExists
		}
		return err
	}

	org.UpdatedAt = model.UpdatedAt

	return nil
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&organizationModel{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrOrganizationNotFound
	}

	return nil
}

// List lists organizations with pagination
func (r *OrganizationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Organization, error) {
	var models []organizationModel

	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	organizations := make([]*domain.Organization, len(models))
	for i, model := range models {
		organizations[i] = r.modelToDomain(&model)
	}

	return organizations, nil
}

// Count counts total organizations
func (r *OrganizationRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&organizationModel{}).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// FindByDomain finds an organization by domain
func (r *OrganizationRepository) FindByDomain(ctx context.Context, domainName string) (*domain.Organization, error) {
	var model organizationModel

	if err := r.db.WithContext(ctx).Where("domain = ?", domainName).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// FindByOwner finds organizations owned by a user
func (r *OrganizationRepository) FindByOwner(ctx context.Context, userID uint) ([]*domain.Organization, error) {
	var models []organizationModel

	if err := r.db.WithContext(ctx).
		Joins("JOIN organization_users ON organizations.id = organization_users.organization_id").
		Where("organization_users.user_id = ? AND organization_users.role = ?", userID, domain.OrganizationRoleOwner).
		Find(&models).Error; err != nil {
		return nil, err
	}

	organizations := make([]*domain.Organization, len(models))
	for i, model := range models {
		organizations[i] = r.modelToDomain(&model)
	}

	return organizations, nil
}

// ExistsBySlug checks if an organization exists by slug
func (r *OrganizationRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&organizationModel{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByDomain checks if an organization exists by domain
func (r *OrganizationRepository) ExistsByDomain(ctx context.Context, domainName string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&organizationModel{}).Where("domain = ?", domainName).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByID checks if an organization exists by ID
func (r *OrganizationRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&organizationModel{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// domainToModel converts domain organization to database model
func (r *OrganizationRepository) domainToModel(org *domain.Organization) *organizationModel {
	return &organizationModel{
		ID:          org.ID,
		Name:        org.Name,
		Slug:        org.Slug,
		Description: org.Description,
		Type:        string(org.Type),
		Domain:      org.Domain,
		LogoURL:     org.LogoURL,
		Website:     org.Website,
		Phone:       org.Phone,
		Address:     org.Address,
		City:        org.City,
		State:       org.State,
		Country:     org.Country,
		PostalCode:  org.PostalCode,
		IsActive:    org.IsActive,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

// modelToDomain converts database model to domain organization
func (r *OrganizationRepository) modelToDomain(model *organizationModel) *domain.Organization {
	return &domain.Organization{
		ID:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		Type:        domain.OrganizationType(model.Type),
		Domain:      model.Domain,
		LogoURL:     model.LogoURL,
		Website:     model.Website,
		Phone:       model.Phone,
		Address:     model.Address,
		City:        model.City,
		State:       model.State,
		Country:     model.Country,
		PostalCode:  model.PostalCode,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// OrganizationUserRepository implements repository.OrganizationUserRepository
type OrganizationUserRepository struct {
	db *gorm.DB
}

// NewOrganizationUserRepository creates a new OrganizationUserRepository
func NewOrganizationUserRepository(db *gorm.DB) repository.OrganizationUserRepository {
	return &OrganizationUserRepository{db: db}
}

// Create creates a new organization user relationship
func (r *OrganizationUserRepository) Create(ctx context.Context, orgUser *domain.OrganizationUser) error {
	model := r.orgUserDomainToModel(orgUser)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("user already in organization")
		}
		return err
	}

	orgUser.ID = model.ID
	orgUser.CreatedAt = model.CreatedAt
	orgUser.UpdatedAt = model.UpdatedAt

	return nil
}

// FindByID finds an organization user relationship by ID
func (r *OrganizationUserRepository) FindByID(ctx context.Context, id uint) (*domain.OrganizationUser, error) {
	var model organizationUserModel

	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotInOrganization
		}
		return nil, err
	}

	return r.orgUserModelToDomain(&model), nil
}

// FindByOrganizationAndUser finds an organization user relationship by organization and user ID
func (r *OrganizationUserRepository) FindByOrganizationAndUser(ctx context.Context, organizationID, userID uint) (*domain.OrganizationUser, error) {
	var model organizationUserModel

	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotInOrganization
		}
		return nil, err
	}

	return r.orgUserModelToDomain(&model), nil
}

// Update updates an organization user relationship
func (r *OrganizationUserRepository) Update(ctx context.Context, orgUser *domain.OrganizationUser) error {
	model := r.orgUserDomainToModel(orgUser)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}

	orgUser.UpdatedAt = model.UpdatedAt

	return nil
}

// Delete deletes an organization user relationship
func (r *OrganizationUserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&organizationUserModel{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotInOrganization
	}

	return nil
}

// FindByOrganization finds all users in an organization
func (r *OrganizationUserRepository) FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.OrganizationUser, error) {
	var models []organizationUserModel

	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	orgUsers := make([]*domain.OrganizationUser, len(models))
	for i, model := range models {
		orgUsers[i] = r.orgUserModelToDomain(&model)
	}

	return orgUsers, nil
}

// FindByUser finds all organizations for a user
func (r *OrganizationUserRepository) FindByUser(ctx context.Context, userID uint) ([]*domain.OrganizationUser, error) {
	var models []organizationUserModel

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	orgUsers := make([]*domain.OrganizationUser, len(models))
	for i, model := range models {
		orgUsers[i] = r.orgUserModelToDomain(&model)
	}

	return orgUsers, nil
}

// FindByRole finds all users with a specific role in an organization
func (r *OrganizationUserRepository) FindByRole(ctx context.Context, organizationID uint, role domain.OrganizationRole) ([]*domain.OrganizationUser, error) {
	var models []organizationUserModel

	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND role = ?", organizationID, string(role)).
		Find(&models).Error; err != nil {
		return nil, err
	}

	orgUsers := make([]*domain.OrganizationUser, len(models))
	for i, model := range models {
		orgUsers[i] = r.orgUserModelToDomain(&model)
	}

	return orgUsers, nil
}

// CountByOrganization counts users in an organization
func (r *OrganizationUserRepository) CountByOrganization(ctx context.Context, organizationID uint) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&organizationUserModel{}).
		Where("organization_id = ?", organizationID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// IsUserInOrganization checks if a user is in an organization
func (r *OrganizationUserRepository) IsUserInOrganization(ctx context.Context, organizationID, userID uint) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&organizationUserModel{}).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserRole gets a user's role in an organization
func (r *OrganizationUserRepository) GetUserRole(ctx context.Context, organizationID, userID uint) (domain.OrganizationRole, error) {
	var model organizationUserModel

	if err := r.db.WithContext(ctx).
		Select("role").
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", domain.ErrUserNotInOrganization
		}
		return "", err
	}

	return domain.OrganizationRole(model.Role), nil
}

// HasRole checks if a user has a specific role in an organization
func (r *OrganizationUserRepository) HasRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&organizationUserModel{}).
		Where("organization_id = ? AND user_id = ? AND role = ?", organizationID, userID, string(role)).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// RemoveUserFromOrganization removes a user from an organization
func (r *OrganizationUserRepository) RemoveUserFromOrganization(ctx context.Context, organizationID, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Delete(&organizationUserModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotInOrganization
	}

	return nil
}

// UpdateUserRole updates a user's role in an organization
func (r *OrganizationUserRepository) UpdateUserRole(ctx context.Context, organizationID, userID uint, role domain.OrganizationRole) error {
	result := r.db.WithContext(ctx).
		Model(&organizationUserModel{}).
		Where("organization_id = ? AND user_id = ?", organizationID, userID).
		Update("role", string(role))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotInOrganization
	}

	return nil
}

// orgUserDomainToModel converts domain organization user to database model
func (r *OrganizationUserRepository) orgUserDomainToModel(orgUser *domain.OrganizationUser) *organizationUserModel {
	return &organizationUserModel{
		ID:             orgUser.ID,
		OrganizationID: orgUser.OrganizationID,
		UserID:         orgUser.UserID,
		Role:           string(orgUser.Role),
		JoinedAt:       orgUser.JoinedAt,
		CreatedAt:      orgUser.CreatedAt,
		UpdatedAt:      orgUser.UpdatedAt,
	}
}

// orgUserModelToDomain converts database model to domain organization user
func (r *OrganizationUserRepository) orgUserModelToDomain(model *organizationUserModel) *domain.OrganizationUser {
	return &domain.OrganizationUser{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		UserID:         model.UserID,
		Role:           domain.OrganizationRole(model.Role),
		JoinedAt:       model.JoinedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

// InvitationRepository implements repository.InvitationRepository
type InvitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository creates a new InvitationRepository
func NewInvitationRepository(db *gorm.DB) repository.InvitationRepository {
	return &InvitationRepository{db: db}
}

// Create creates a new invitation
func (r *InvitationRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	model := r.invitationDomainToModel(invitation)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("invitation token already exists")
		}
		return err
	}

	invitation.ID = model.ID
	invitation.CreatedAt = model.CreatedAt
	invitation.UpdatedAt = model.UpdatedAt

	return nil
}

// FindByID finds an invitation by ID
func (r *InvitationRepository) FindByID(ctx context.Context, id uint) (*domain.Invitation, error) {
	var model invitationModel

	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, err
	}

	return r.invitationModelToDomain(&model), nil
}

// FindByToken finds an invitation by token
func (r *InvitationRepository) FindByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	var model invitationModel

	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvitationNotFound
		}
		return nil, err
	}

	return r.invitationModelToDomain(&model), nil
}

// Update updates an invitation
func (r *InvitationRepository) Update(ctx context.Context, invitation *domain.Invitation) error {
	model := r.invitationDomainToModel(invitation)

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}

	invitation.UpdatedAt = model.UpdatedAt

	return nil
}

// Delete deletes an invitation
func (r *InvitationRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&invitationModel{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrInvitationNotFound
	}

	return nil
}

// FindByOrganization finds invitations by organization
func (r *InvitationRepository) FindByOrganization(ctx context.Context, organizationID uint) ([]*domain.Invitation, error) {
	var models []invitationModel

	if err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(models))
	for i, model := range models {
		invitations[i] = r.invitationModelToDomain(&model)
	}

	return invitations, nil
}

// FindByEmail finds invitations by email
func (r *InvitationRepository) FindByEmail(ctx context.Context, email string) ([]*domain.Invitation, error) {
	var models []invitationModel

	if err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(models))
	for i, model := range models {
		invitations[i] = r.invitationModelToDomain(&model)
	}

	return invitations, nil
}

// FindByInviter finds invitations by inviter
func (r *InvitationRepository) FindByInviter(ctx context.Context, inviterID uint) ([]*domain.Invitation, error) {
	var models []invitationModel

	if err := r.db.WithContext(ctx).
		Where("inviter_id = ?", inviterID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(models))
	for i, model := range models {
		invitations[i] = r.invitationModelToDomain(&model)
	}

	return invitations, nil
}

// FindByStatus finds invitations by status
func (r *InvitationRepository) FindByStatus(ctx context.Context, status domain.InvitationStatus) ([]*domain.Invitation, error) {
	var models []invitationModel

	if err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(models))
	for i, model := range models {
		invitations[i] = r.invitationModelToDomain(&model)
	}

	return invitations, nil
}

// FindExpired finds expired invitations
func (r *InvitationRepository) FindExpired(ctx context.Context) ([]*domain.Invitation, error) {
	var models []invitationModel

	if err := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", time.Now(), string(domain.InvitationStatusPending)).
		Find(&models).Error; err != nil {
		return nil, err
	}

	invitations := make([]*domain.Invitation, len(models))
	for i, model := range models {
		invitations[i] = r.invitationModelToDomain(&model)
	}

	return invitations, nil
}

// ExistsByToken checks if an invitation exists by token
func (r *InvitationRepository) ExistsByToken(ctx context.Context, token string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&invitationModel{}).
		Where("token = ?", token).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsPendingByEmail checks if a pending invitation exists for an email in an organization
func (r *InvitationRepository) ExistsPendingByEmail(ctx context.Context, organizationID uint, email string) (bool, error) {
	var count int64

	if err := r.db.WithContext(ctx).
		Model(&invitationModel{}).
		Where("organization_id = ? AND email = ? AND status = ?", organizationID, email, string(domain.InvitationStatusPending)).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteExpired deletes expired invitations
func (r *InvitationRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", before, string(domain.InvitationStatusPending)).
		Delete(&invitationModel{}).Error
}

// MarkExpired marks invitations as expired
func (r *InvitationRepository) MarkExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Model(&invitationModel{}).
		Where("expires_at < ? AND status = ?", before, string(domain.InvitationStatusPending)).
		Update("status", string(domain.InvitationStatusExpired)).Error
}

// invitationDomainToModel converts domain invitation to database model
func (r *InvitationRepository) invitationDomainToModel(invitation *domain.Invitation) *invitationModel {
	return &invitationModel{
		ID:             invitation.ID,
		OrganizationID: invitation.OrganizationID,
		InviterID:      invitation.InviterID,
		Email:          invitation.Email,
		Role:           string(invitation.Role),
		Token:          invitation.Token,
		Status:         string(invitation.Status),
		Message:        invitation.Message,
		ExpiresAt:      invitation.ExpiresAt,
		AcceptedAt:     invitation.AcceptedAt,
		CreatedAt:      invitation.CreatedAt,
		UpdatedAt:      invitation.UpdatedAt,
	}
}

// invitationModelToDomain converts database model to domain invitation
func (r *InvitationRepository) invitationModelToDomain(model *invitationModel) *domain.Invitation {
	return &domain.Invitation{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		InviterID:      model.InviterID,
		Email:          model.Email,
		Role:           domain.OrganizationRole(model.Role),
		Token:          model.Token,
		Status:         domain.InvitationStatus(model.Status),
		Message:        model.Message,
		ExpiresAt:      model.ExpiresAt,
		AcceptedAt:     model.AcceptedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}
