// @kthulu:module:contacts
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"backend/internal/domain"
	"backend/internal/repository"
)

// contactModel represents the database model for contacts
type contactModel struct {
	ID             uint      `gorm:"primaryKey"`
	OrganizationID uint      `gorm:"not null;index"`
	Type           string    `gorm:"not null;size:20;index"`
	CompanyName    string    `gorm:"size:200;index"`
	FirstName      string    `gorm:"size:100"`
	LastName       string    `gorm:"size:100"`
	Email          string    `gorm:"size:255;index"`
	Phone          string    `gorm:"size:20"`
	Mobile         string    `gorm:"size:20"`
	Website        string    `gorm:"size:500"`
	TaxNumber      string    `gorm:"size:50"`
	Notes          string    `gorm:"type:text"`
	IsActive       bool      `gorm:"default:true;index"`
	CreatedAt      Timestamp `gorm:"column:created_at"`
	UpdatedAt      Timestamp `gorm:"column:updated_at"`

	// Relationships
	Addresses []contactAddressModel `gorm:"foreignKey:ContactID"`
	Phones    []contactPhoneModel   `gorm:"foreignKey:ContactID"`
}

func (contactModel) TableName() string {
	return "contacts"
}

// contactAddressModel represents the database model for contact addresses
type contactAddressModel struct {
	ID           uint      `gorm:"primaryKey"`
	ContactID    uint      `gorm:"not null;index"`
	Type         string    `gorm:"not null;size:20"`
	AddressLine1 string    `gorm:"not null;size:200"`
	AddressLine2 string    `gorm:"size:200"`
	City         string    `gorm:"not null;size:100"`
	State        string    `gorm:"size:100"`
	Country      string    `gorm:"not null;size:100"`
	PostalCode   string    `gorm:"size:20"`
	IsPrimary    bool      `gorm:"default:false"`
	CreatedAt    Timestamp `gorm:"column:created_at"`
	UpdatedAt    Timestamp `gorm:"column:updated_at"`
}

func (contactAddressModel) TableName() string {
	return "contact_addresses"
}

// contactPhoneModel represents the database model for contact phones
type contactPhoneModel struct {
	ID        uint      `gorm:"primaryKey"`
	ContactID uint      `gorm:"not null;index"`
	Type      string    `gorm:"not null;size:20"`
	Number    string    `gorm:"not null;size:20"`
	Extension string    `gorm:"size:10"`
	IsPrimary bool      `gorm:"default:false"`
	CreatedAt Timestamp `gorm:"column:created_at"`
	UpdatedAt Timestamp `gorm:"column:updated_at"`
}

func (contactPhoneModel) TableName() string {
	return "contact_phones"
}

// ContactRepository implements the contact repository interface using GORM
type ContactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) repository.ContactRepository {
	return &ContactRepository{db: db}
}

// Create creates a new contact
func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	model := r.domainToModel(contact)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}

	contact.ID = model.ID
	contact.CreatedAt = model.CreatedAt.Time
	contact.UpdatedAt = model.UpdatedAt.Time

	return nil
}

// GetByID retrieves a contact by ID
func (r *ContactRepository) GetByID(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error) {
	var model contactModel

	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", contactID, organizationID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// GetByEmail retrieves a contact by email
func (r *ContactRepository) GetByEmail(ctx context.Context, organizationID uint, email string) (*domain.Contact, error) {
	var model contactModel

	err := r.db.WithContext(ctx).
		Where("email = ? AND organization_id = ?", email, organizationID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact by email: %w", err)
	}

	return r.modelToDomain(&model), nil
}

// Update updates an existing contact
func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	model := r.domainToModel(contact)

	result := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", contact.ID, contact.OrganizationID).
		Updates(&model)

	if result.Error != nil {
		return fmt.Errorf("failed to update contact: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrContactNotFound
	}

	contact.UpdatedAt = time.Now()
	return nil
}

// Delete deletes a contact
func (r *ContactRepository) Delete(ctx context.Context, organizationID, contactID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", contactID, organizationID).
		Delete(&contactModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete contact: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrContactNotFound
	}

	return nil
}

// List retrieves contacts with filtering and pagination
func (r *ContactRepository) List(ctx context.Context, organizationID uint, filters repository.ContactFilters) ([]*domain.Contact, int64, error) {
	query := r.db.WithContext(ctx).Model(&contactModel{}).
		Where("organization_id = ?", organizationID)

	// Apply filters
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}

	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}

	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where(
			"LOWER(company_name) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if filters.CreatedFrom != nil {
		if createdFrom, err := time.Parse(time.RFC3339, *filters.CreatedFrom); err == nil {
			query = query.Where("created_at >= ?", createdFrom)
		}
	}

	if filters.CreatedTo != nil {
		if createdTo, err := time.Parse(time.RFC3339, *filters.CreatedTo); err == nil {
			query = query.Where("created_at <= ?", createdTo)
		}
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count contacts: %w", err)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", filters.SortBy, strings.ToUpper(filters.SortOrder))
	query = query.Order(orderClause)

	// Apply pagination
	query = query.Offset(filters.GetOffset()).Limit(filters.PageSize)

	var models []contactModel
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list contacts: %w", err)
	}

	contacts := make([]*domain.Contact, len(models))
	for i, model := range models {
		contacts[i] = r.modelToDomain(&model)
	}

	return contacts, total, nil
}

// CreateAddress creates a new contact address
func (r *ContactRepository) CreateAddress(ctx context.Context, address *domain.ContactAddress) error {
	// If this is set as primary, unset other primary addresses of the same type
	if address.IsPrimary {
		if err := r.db.WithContext(ctx).
			Model(&contactAddressModel{}).
			Where("contact_id = ? AND type = ?", address.ContactID, address.Type).
			Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("failed to unset primary addresses: %w", err)
		}
	}

	model := r.addressDomainToModel(address)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create contact address: %w", err)
	}

	address.ID = model.ID
	address.CreatedAt = model.CreatedAt.Time
	address.UpdatedAt = model.UpdatedAt.Time

	return nil
}

// GetAddressesByContactID retrieves all addresses for a contact
func (r *ContactRepository) GetAddressesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactAddress, error) {
	var models []contactAddressModel

	if err := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("is_primary DESC, created_at ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get contact addresses: %w", err)
	}

	addresses := make([]*domain.ContactAddress, len(models))
	for i, model := range models {
		addresses[i] = r.addressModelToDomain(&model)
	}

	return addresses, nil
}

// GetAddressByID retrieves a specific address
func (r *ContactRepository) GetAddressByID(ctx context.Context, contactID, addressID uint) (*domain.ContactAddress, error) {
	var model contactAddressModel

	err := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", addressID, contactID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAddressNotFound
		}
		return nil, fmt.Errorf("failed to get contact address: %w", err)
	}

	return r.addressModelToDomain(&model), nil
}

// UpdateAddress updates a contact address
func (r *ContactRepository) UpdateAddress(ctx context.Context, address *domain.ContactAddress) error {
	model := r.addressDomainToModel(address)

	result := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", address.ID, address.ContactID).
		Updates(&model)

	if result.Error != nil {
		return fmt.Errorf("failed to update contact address: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

// DeleteAddress deletes a contact address
func (r *ContactRepository) DeleteAddress(ctx context.Context, contactID, addressID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", addressID, contactID).
		Delete(&contactAddressModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete contact address: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrAddressNotFound
	}

	return nil
}

// SetPrimaryAddress sets an address as primary
func (r *ContactRepository) SetPrimaryAddress(ctx context.Context, contactID, addressID uint) error {
	// First, get the address to know its type
	var address contactAddressModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", addressID, contactID).
		First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrAddressNotFound
		}
		return fmt.Errorf("failed to get address: %w", err)
	}

	// Unset other primary addresses of the same type
	if err := r.db.WithContext(ctx).
		Model(&contactAddressModel{}).
		Where("contact_id = ? AND type = ?", contactID, address.Type).
		Update("is_primary", false).Error; err != nil {
		return fmt.Errorf("failed to unset primary addresses: %w", err)
	}

	// Set this address as primary
	if err := r.db.WithContext(ctx).
		Model(&contactAddressModel{}).
		Where("id = ?", addressID).
		Update("is_primary", true).Error; err != nil {
		return fmt.Errorf("failed to set primary address: %w", err)
	}

	return nil
}

// CreatePhone creates a new contact phone
func (r *ContactRepository) CreatePhone(ctx context.Context, phone *domain.ContactPhone) error {
	// If this is set as primary, unset other primary phones of the same type
	if phone.IsPrimary {
		if err := r.db.WithContext(ctx).
			Model(&contactPhoneModel{}).
			Where("contact_id = ? AND type = ?", phone.ContactID, phone.Type).
			Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("failed to unset primary phones: %w", err)
		}
	}

	model := r.phoneDomainToModel(phone)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return fmt.Errorf("failed to create contact phone: %w", err)
	}

	phone.ID = model.ID
	phone.CreatedAt = model.CreatedAt.Time
	phone.UpdatedAt = model.UpdatedAt.Time

	return nil
}

// GetPhonesByContactID retrieves all phones for a contact
func (r *ContactRepository) GetPhonesByContactID(ctx context.Context, contactID uint) ([]*domain.ContactPhone, error) {
	var models []contactPhoneModel

	if err := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("is_primary DESC, created_at ASC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get contact phones: %w", err)
	}

	phones := make([]*domain.ContactPhone, len(models))
	for i, model := range models {
		phones[i] = r.phoneModelToDomain(&model)
	}

	return phones, nil
}

// GetPhoneByID retrieves a specific phone
func (r *ContactRepository) GetPhoneByID(ctx context.Context, contactID, phoneID uint) (*domain.ContactPhone, error) {
	var model contactPhoneModel

	err := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", phoneID, contactID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPhoneNotFound
		}
		return nil, fmt.Errorf("failed to get contact phone: %w", err)
	}

	return r.phoneModelToDomain(&model), nil
}

// UpdatePhone updates a contact phone
func (r *ContactRepository) UpdatePhone(ctx context.Context, phone *domain.ContactPhone) error {
	model := r.phoneDomainToModel(phone)

	result := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", phone.ID, phone.ContactID).
		Updates(&model)

	if result.Error != nil {
		return fmt.Errorf("failed to update contact phone: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPhoneNotFound
	}

	return nil
}

// DeletePhone deletes a contact phone
func (r *ContactRepository) DeletePhone(ctx context.Context, contactID, phoneID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", phoneID, contactID).
		Delete(&contactPhoneModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete contact phone: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrPhoneNotFound
	}

	return nil
}

// SetPrimaryPhone sets a phone as primary
func (r *ContactRepository) SetPrimaryPhone(ctx context.Context, contactID, phoneID uint) error {
	// First, get the phone to know its type
	var phone contactPhoneModel
	if err := r.db.WithContext(ctx).
		Where("id = ? AND contact_id = ?", phoneID, contactID).
		First(&phone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrPhoneNotFound
		}
		return fmt.Errorf("failed to get phone: %w", err)
	}

	// Unset other primary phones of the same type
	if err := r.db.WithContext(ctx).
		Model(&contactPhoneModel{}).
		Where("contact_id = ? AND type = ?", contactID, phone.Type).
		Update("is_primary", false).Error; err != nil {
		return fmt.Errorf("failed to unset primary phones: %w", err)
	}

	// Set this phone as primary
	if err := r.db.WithContext(ctx).
		Model(&contactPhoneModel{}).
		Where("id = ?", phoneID).
		Update("is_primary", true).Error; err != nil {
		return fmt.Errorf("failed to set primary phone: %w", err)
	}

	return nil
}

// BulkCreate creates multiple contacts
func (r *ContactRepository) BulkCreate(ctx context.Context, contacts []*domain.Contact) error {
	models := make([]contactModel, len(contacts))
	for i, contact := range contacts {
		models[i] = *r.domainToModel(contact)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(models, 100).Error; err != nil {
		return fmt.Errorf("failed to bulk create contacts: %w", err)
	}

	// Update domain objects with generated IDs
	for i, model := range models {
		contacts[i].ID = model.ID
		contacts[i].CreatedAt = model.CreatedAt.Time
		contacts[i].UpdatedAt = model.UpdatedAt.Time
	}

	return nil
}

// BulkUpdate updates multiple contacts
func (r *ContactRepository) BulkUpdate(ctx context.Context, contacts []*domain.Contact) error {
	for _, contact := range contacts {
		if err := r.Update(ctx, contact); err != nil {
			return err
		}
	}
	return nil
}

// BulkDelete deletes multiple contacts
func (r *ContactRepository) BulkDelete(ctx context.Context, organizationID uint, contactIDs []uint) error {
	result := r.db.WithContext(ctx).
		Where("organization_id = ? AND id IN ?", organizationID, contactIDs).
		Delete(&contactModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete contacts: %w", result.Error)
	}

	return nil
}

// GetContactStats retrieves contact statistics
func (r *ContactRepository) GetContactStats(ctx context.Context, organizationID uint) (*repository.ContactStats, error) {
	stats := &repository.ContactStats{}

	// Total contacts
	if err := r.db.WithContext(ctx).
		Model(&contactModel{}).
		Where("organization_id = ?", organizationID).
		Count(&stats.TotalContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total contacts: %w", err)
	}

	// Active contacts
	if err := r.db.WithContext(ctx).
		Model(&contactModel{}).
		Where("organization_id = ? AND is_active = ?", organizationID, true).
		Count(&stats.ActiveContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active contacts: %w", err)
	}

	// Inactive contacts
	stats.InactiveContacts = stats.TotalContacts - stats.ActiveContacts

	// Count by type
	var typeCounts []struct {
		Type  string
		Count int64
	}

	if err := r.db.WithContext(ctx).
		Model(&contactModel{}).
		Select("type, COUNT(*) as count").
		Where("organization_id = ? AND is_active = ?", organizationID, true).
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count contacts by type: %w", err)
	}

	for _, tc := range typeCounts {
		switch tc.Type {
		case "customer":
			stats.CustomerCount = tc.Count
		case "supplier":
			stats.SupplierCount = tc.Count
		case "lead":
			stats.LeadCount = tc.Count
		case "partner":
			stats.PartnerCount = tc.Count
		}
	}

	// Recent contacts (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := r.db.WithContext(ctx).
		Model(&contactModel{}).
		Where("organization_id = ? AND created_at >= ?", organizationID, thirtyDaysAgo).
		Count(&stats.RecentContacts).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent contacts: %w", err)
	}

	return stats, nil
}

// Helper methods for model conversion

func (r *ContactRepository) domainToModel(contact *domain.Contact) *contactModel {
	return &contactModel{
		ID:             contact.ID,
		OrganizationID: contact.OrganizationID,
		Type:           string(contact.Type),
		CompanyName:    contact.CompanyName,
		FirstName:      contact.FirstName,
		LastName:       contact.LastName,
		Email:          contact.Email,
		Phone:          contact.Phone,
		Mobile:         contact.Mobile,
		Website:        contact.Website,
		TaxNumber:      contact.TaxNumber,
		Notes:          contact.Notes,
		IsActive:       contact.IsActive,
		CreatedAt:      Timestamp{Time: contact.CreatedAt},
		UpdatedAt:      Timestamp{Time: contact.UpdatedAt},
	}
}

func (r *ContactRepository) modelToDomain(model *contactModel) *domain.Contact {
	return &domain.Contact{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		Type:           domain.ContactType(model.Type),
		CompanyName:    model.CompanyName,
		FirstName:      model.FirstName,
		LastName:       model.LastName,
		Email:          model.Email,
		Phone:          model.Phone,
		Mobile:         model.Mobile,
		Website:        model.Website,
		TaxNumber:      model.TaxNumber,
		Notes:          model.Notes,
		IsActive:       model.IsActive,
		CreatedAt:      model.CreatedAt.Time,
		UpdatedAt:      model.UpdatedAt.Time,
	}
}

func (r *ContactRepository) addressDomainToModel(address *domain.ContactAddress) *contactAddressModel {
	return &contactAddressModel{
		ID:           address.ID,
		ContactID:    address.ContactID,
		Type:         string(address.Type),
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		Country:      address.Country,
		PostalCode:   address.PostalCode,
		IsPrimary:    address.IsPrimary,
		CreatedAt:    Timestamp{Time: address.CreatedAt},
		UpdatedAt:    Timestamp{Time: address.UpdatedAt},
	}
}

func (r *ContactRepository) addressModelToDomain(model *contactAddressModel) *domain.ContactAddress {
	return &domain.ContactAddress{
		ID:           model.ID,
		ContactID:    model.ContactID,
		Type:         domain.AddressType(model.Type),
		AddressLine1: model.AddressLine1,
		AddressLine2: model.AddressLine2,
		City:         model.City,
		State:        model.State,
		Country:      model.Country,
		PostalCode:   model.PostalCode,
		IsPrimary:    model.IsPrimary,
		CreatedAt:    model.CreatedAt.Time,
		UpdatedAt:    model.UpdatedAt.Time,
	}
}

func (r *ContactRepository) phoneDomainToModel(phone *domain.ContactPhone) *contactPhoneModel {
	return &contactPhoneModel{
		ID:        phone.ID,
		ContactID: phone.ContactID,
		Type:      string(phone.Type),
		Number:    phone.Number,
		Extension: phone.Extension,
		IsPrimary: phone.IsPrimary,
		CreatedAt: Timestamp{Time: phone.CreatedAt},
		UpdatedAt: Timestamp{Time: phone.UpdatedAt},
	}
}

func (r *ContactRepository) phoneModelToDomain(model *contactPhoneModel) *domain.ContactPhone {
	return &domain.ContactPhone{
		ID:        model.ID,
		ContactID: model.ContactID,
		Type:      domain.PhoneType(model.Type),
		Number:    model.Number,
		Extension: model.Extension,
		IsPrimary: model.IsPrimary,
		CreatedAt: model.CreatedAt.Time,
		UpdatedAt: model.UpdatedAt.Time,
	}
}
