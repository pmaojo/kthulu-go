// @kthulu:module:contacts
package usecase

import (
	"context"
	"fmt"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.uber.org/zap"
)

// ContactUseCase handles business logic for contact management
type ContactUseCase struct {
	contactRepo repository.ContactRepository
	logger      *zap.Logger
}

// NewContactUseCase creates a new contact use case
func NewContactUseCase(contactRepo repository.ContactRepository, logger *zap.Logger) *ContactUseCase {
	return &ContactUseCase{
		contactRepo: contactRepo,
		logger:      logger,
	}
}

// CreateContact creates a new contact
func (uc *ContactUseCase) CreateContact(ctx context.Context, organizationID uint, req CreateContactRequest) (*domain.Contact, error) {
	uc.logger.Info("Creating new contact",
		zap.Uint("organization_id", organizationID),
		zap.String("type", string(req.Type)),
		zap.String("company_name", req.CompanyName),
		zap.String("email", req.Email),
	)

	// Check if contact with email already exists
	if req.Email != "" {
		existing, err := uc.contactRepo.GetByEmail(ctx, organizationID, req.Email)
		if err == nil && existing != nil {
			return nil, domain.ErrContactAlreadyExists
		}
	}

	// Create new contact
	contact, err := domain.NewContact(
		organizationID,
		req.Type,
		req.CompanyName,
		req.FirstName,
		req.LastName,
		req.Email,
	)
	if err != nil {
		uc.logger.Error("Failed to create contact domain object", zap.Error(err))
		return nil, err
	}

	// Update additional fields
	if err := contact.UpdateBasicInfo(
		req.CompanyName,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Mobile,
		req.Website,
		req.TaxNumber,
		req.Notes,
	); err != nil {
		uc.logger.Error("Failed to update contact basic info", zap.Error(err))
		return nil, err
	}

	// Save to repository
	if err := uc.contactRepo.Create(ctx, contact); err != nil {
		uc.logger.Error("Failed to save contact to repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	uc.logger.Info("Contact created successfully",
		zap.Uint("contact_id", contact.ID),
		zap.String("display_name", contact.GetDisplayName()),
	)

	return contact, nil
}

// GetContact retrieves a contact by ID
func (uc *ContactUseCase) GetContact(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error) {
	contact, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		uc.logger.Error("Failed to get contact",
			zap.Uint("organization_id", organizationID),
			zap.Uint("contact_id", contactID),
			zap.Error(err),
		)
		return nil, err
	}

	// Load related data
	if err := uc.loadContactRelations(ctx, contact); err != nil {
		uc.logger.Warn("Failed to load contact relations", zap.Error(err))
		// Don't fail the request, just log the warning
	}

	return contact, nil
}

// UpdateContact updates an existing contact
func (uc *ContactUseCase) UpdateContact(ctx context.Context, organizationID, contactID uint, req UpdateContactRequest) (*domain.Contact, error) {
	uc.logger.Info("Updating contact",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
	)

	// Get existing contact
	contact, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if new email already exists
	if req.Email != "" && req.Email != contact.Email {
		existing, err := uc.contactRepo.GetByEmail(ctx, organizationID, req.Email)
		if err == nil && existing != nil && existing.ID != contactID {
			return nil, domain.ErrContactAlreadyExists
		}
	}

	// Update contact information
	if err := contact.UpdateBasicInfo(
		req.CompanyName,
		req.FirstName,
		req.LastName,
		req.Email,
		req.Phone,
		req.Mobile,
		req.Website,
		req.TaxNumber,
		req.Notes,
	); err != nil {
		uc.logger.Error("Failed to update contact basic info", zap.Error(err))
		return nil, err
	}

	// Save changes
	if err := uc.contactRepo.Update(ctx, contact); err != nil {
		uc.logger.Error("Failed to update contact in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	uc.logger.Info("Contact updated successfully", zap.Uint("contact_id", contactID))
	return contact, nil
}

// DeleteContact deletes a contact
func (uc *ContactUseCase) DeleteContact(ctx context.Context, organizationID, contactID uint) error {
	uc.logger.Info("Deleting contact",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
	)

	if err := uc.contactRepo.Delete(ctx, organizationID, contactID); err != nil {
		uc.logger.Error("Failed to delete contact", zap.Error(err))
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	uc.logger.Info("Contact deleted successfully", zap.Uint("contact_id", contactID))
	return nil
}

// ListContacts retrieves a list of contacts with filtering and pagination
func (uc *ContactUseCase) ListContacts(ctx context.Context, organizationID uint, filters repository.ContactFilters) (*ContactListResponse, error) {
	// Validate and set defaults for filters
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	contacts, total, err := uc.contactRepo.List(ctx, organizationID, filters)
	if err != nil {
		uc.logger.Error("Failed to list contacts", zap.Error(err))
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	return &ContactListResponse{
		Contacts:   contacts,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: (total + int64(filters.PageSize) - 1) / int64(filters.PageSize),
	}, nil
}

// SetContactActive sets the active status of a contact
func (uc *ContactUseCase) SetContactActive(ctx context.Context, organizationID, contactID uint, active bool) error {
	uc.logger.Info("Setting contact active status",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Bool("active", active),
	)

	contact, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		return err
	}

	contact.SetActive(active)

	if err := uc.contactRepo.Update(ctx, contact); err != nil {
		uc.logger.Error("Failed to update contact active status", zap.Error(err))
		return fmt.Errorf("failed to update contact status: %w", err)
	}

	uc.logger.Info("Contact active status updated successfully", zap.Uint("contact_id", contactID))
	return nil
}

// ConvertLeadToCustomer converts a lead to a customer
func (uc *ContactUseCase) ConvertLeadToCustomer(ctx context.Context, organizationID, contactID uint) (*domain.Contact, error) {
	uc.logger.Info("Converting lead to customer",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
	)

	contact, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		return nil, err
	}

	if err := contact.ConvertToCustomer(); err != nil {
		return nil, err
	}

	if err := uc.contactRepo.Update(ctx, contact); err != nil {
		uc.logger.Error("Failed to convert lead to customer", zap.Error(err))
		return nil, fmt.Errorf("failed to convert lead: %w", err)
	}

	uc.logger.Info("Lead converted to customer successfully", zap.Uint("contact_id", contactID))
	return contact, nil
}

// GetContactStats retrieves contact statistics for an organization
func (uc *ContactUseCase) GetContactStats(ctx context.Context, organizationID uint) (*repository.ContactStats, error) {
	stats, err := uc.contactRepo.GetContactStats(ctx, organizationID)
	if err != nil {
		uc.logger.Error("Failed to get contact stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get contact stats: %w", err)
	}

	return stats, nil
}

// AddContactAddress adds an address to a contact
func (uc *ContactUseCase) AddContactAddress(ctx context.Context, organizationID, contactID uint, req CreateAddressRequest) (*domain.ContactAddress, error) {
	uc.logger.Info("Adding address to contact",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.String("type", string(req.Type)),
	)

	// Verify contact exists and belongs to organization
	_, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		return nil, err
	}

	// Create new address
	address, err := domain.NewContactAddress(
		contactID,
		req.Type,
		req.AddressLine1,
		req.AddressLine2,
		req.City,
		req.State,
		req.Country,
		req.PostalCode,
		req.IsPrimary,
	)
	if err != nil {
		return nil, err
	}

	// If this is set as primary, unset other primary addresses
	if req.IsPrimary {
		// This would be handled by the repository implementation
	}

	if err := uc.contactRepo.CreateAddress(ctx, address); err != nil {
		uc.logger.Error("Failed to create contact address", zap.Error(err))
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	uc.logger.Info("Contact address created successfully", zap.Uint("address_id", address.ID))
	return address, nil
}

// AddContactPhone adds a phone number to a contact
func (uc *ContactUseCase) AddContactPhone(ctx context.Context, organizationID, contactID uint, req CreatePhoneRequest) (*domain.ContactPhone, error) {
	uc.logger.Info("Adding phone to contact",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.String("type", string(req.Type)),
	)

	// Verify contact exists and belongs to organization
	_, err := uc.contactRepo.GetByID(ctx, organizationID, contactID)
	if err != nil {
		return nil, err
	}

	// Create new phone
	phone, err := domain.NewContactPhone(
		contactID,
		req.Type,
		req.Number,
		req.Extension,
		req.IsPrimary,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.contactRepo.CreatePhone(ctx, phone); err != nil {
		uc.logger.Error("Failed to create contact phone", zap.Error(err))
		return nil, fmt.Errorf("failed to create phone: %w", err)
	}

	uc.logger.Info("Contact phone created successfully", zap.Uint("phone_id", phone.ID))
	return phone, nil
}

// UpdateContactAddress updates an existing contact address
func (uc *ContactUseCase) UpdateContactAddress(ctx context.Context, organizationID, contactID, addressID uint, req UpdateAddressRequest) (*domain.ContactAddress, error) {
	uc.logger.Info("Updating contact address",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("address_id", addressID),
	)

	// Verify contact exists
	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return nil, err
	}

	// Get existing address to retain timestamps
	existing, err := uc.contactRepo.GetAddressByID(ctx, contactID, addressID)
	if err != nil {
		return nil, err
	}

	address, err := domain.NewContactAddress(
		contactID,
		req.Type,
		req.AddressLine1,
		req.AddressLine2,
		req.City,
		req.State,
		req.Country,
		req.PostalCode,
		req.IsPrimary,
	)
	if err != nil {
		return nil, err
	}
	address.ID = addressID
	address.CreatedAt = existing.CreatedAt

	if err := uc.contactRepo.UpdateAddress(ctx, address); err != nil {
		uc.logger.Error("Failed to update contact address", zap.Error(err))
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	if req.IsPrimary {
		if err := uc.contactRepo.SetPrimaryAddress(ctx, contactID, addressID); err != nil {
			uc.logger.Error("Failed to set primary address", zap.Error(err))
			return nil, fmt.Errorf("failed to set primary address: %w", err)
		}
		address.IsPrimary = true
	}

	uc.logger.Info("Contact address updated successfully", zap.Uint("address_id", addressID))
	return address, nil
}

// DeleteContactAddress deletes a contact address
func (uc *ContactUseCase) DeleteContactAddress(ctx context.Context, organizationID, contactID, addressID uint) error {
	uc.logger.Info("Deleting contact address",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("address_id", addressID),
	)

	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return err
	}

	if err := uc.contactRepo.DeleteAddress(ctx, contactID, addressID); err != nil {
		uc.logger.Error("Failed to delete contact address", zap.Error(err))
		return fmt.Errorf("failed to delete address: %w", err)
	}

	return nil
}

// SetPrimaryAddress sets an address as primary
func (uc *ContactUseCase) SetPrimaryAddress(ctx context.Context, organizationID, contactID, addressID uint) error {
	uc.logger.Info("Setting primary contact address",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("address_id", addressID),
	)

	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return err
	}

	if err := uc.contactRepo.SetPrimaryAddress(ctx, contactID, addressID); err != nil {
		uc.logger.Error("Failed to set primary address", zap.Error(err))
		return fmt.Errorf("failed to set primary address: %w", err)
	}

	return nil
}

// UpdateContactPhone updates an existing contact phone
func (uc *ContactUseCase) UpdateContactPhone(ctx context.Context, organizationID, contactID, phoneID uint, req UpdatePhoneRequest) (*domain.ContactPhone, error) {
	uc.logger.Info("Updating contact phone",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("phone_id", phoneID),
	)

	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return nil, err
	}

	existing, err := uc.contactRepo.GetPhoneByID(ctx, contactID, phoneID)
	if err != nil {
		return nil, err
	}

	phone, err := domain.NewContactPhone(
		contactID,
		req.Type,
		req.Number,
		req.Extension,
		req.IsPrimary,
	)
	if err != nil {
		return nil, err
	}
	phone.ID = phoneID
	phone.CreatedAt = existing.CreatedAt

	if err := uc.contactRepo.UpdatePhone(ctx, phone); err != nil {
		uc.logger.Error("Failed to update contact phone", zap.Error(err))
		return nil, fmt.Errorf("failed to update phone: %w", err)
	}

	if req.IsPrimary {
		if err := uc.contactRepo.SetPrimaryPhone(ctx, contactID, phoneID); err != nil {
			uc.logger.Error("Failed to set primary phone", zap.Error(err))
			return nil, fmt.Errorf("failed to set primary phone: %w", err)
		}
		phone.IsPrimary = true
	}

	uc.logger.Info("Contact phone updated successfully", zap.Uint("phone_id", phoneID))
	return phone, nil
}

// DeleteContactPhone deletes a contact phone
func (uc *ContactUseCase) DeleteContactPhone(ctx context.Context, organizationID, contactID, phoneID uint) error {
	uc.logger.Info("Deleting contact phone",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("phone_id", phoneID),
	)

	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return err
	}

	if err := uc.contactRepo.DeletePhone(ctx, contactID, phoneID); err != nil {
		uc.logger.Error("Failed to delete contact phone", zap.Error(err))
		return fmt.Errorf("failed to delete phone: %w", err)
	}

	return nil
}

// SetPrimaryPhone sets a phone as primary
func (uc *ContactUseCase) SetPrimaryPhone(ctx context.Context, organizationID, contactID, phoneID uint) error {
	uc.logger.Info("Setting primary contact phone",
		zap.Uint("organization_id", organizationID),
		zap.Uint("contact_id", contactID),
		zap.Uint("phone_id", phoneID),
	)

	if _, err := uc.contactRepo.GetByID(ctx, organizationID, contactID); err != nil {
		return err
	}

	if err := uc.contactRepo.SetPrimaryPhone(ctx, contactID, phoneID); err != nil {
		uc.logger.Error("Failed to set primary phone", zap.Error(err))
		return fmt.Errorf("failed to set primary phone: %w", err)
	}

	return nil
}

// loadContactRelations loads addresses and phones for a contact
func (uc *ContactUseCase) loadContactRelations(ctx context.Context, contact *domain.Contact) error {
	// Load addresses
	addresses, err := uc.contactRepo.GetAddressesByContactID(ctx, contact.ID)
	if err != nil {
		return err
	}
	// Convert []*domain.ContactAddress to []domain.ContactAddress
	contact.Addresses = make([]domain.ContactAddress, len(addresses))
	for i, addr := range addresses {
		contact.Addresses[i] = *addr
	}

	// Load phones
	phones, err := uc.contactRepo.GetPhonesByContactID(ctx, contact.ID)
	if err != nil {
		return err
	}
	// Convert []*domain.ContactPhone to []domain.ContactPhone
	contact.Phones = make([]domain.ContactPhone, len(phones))
	for i, phone := range phones {
		contact.Phones[i] = *phone
	}

	return nil
}

// Request/Response DTOs

// CreateContactRequest represents a request to create a new contact
type CreateContactRequest struct {
	Type        domain.ContactType `json:"type" validate:"required,oneof=customer supplier lead partner"`
	CompanyName string             `json:"companyName,omitempty" validate:"max=200"`
	FirstName   string             `json:"firstName,omitempty" validate:"max=100"`
	LastName    string             `json:"lastName,omitempty" validate:"max=100"`
	Email       string             `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone       string             `json:"phone,omitempty" validate:"max=20"`
	Mobile      string             `json:"mobile,omitempty" validate:"max=20"`
	Website     string             `json:"website,omitempty" validate:"omitempty,url,max=500"`
	TaxNumber   string             `json:"taxNumber,omitempty" validate:"max=50"`
	Notes       string             `json:"notes,omitempty"`
}

// UpdateContactRequest represents a request to update a contact
type UpdateContactRequest struct {
	CompanyName string `json:"companyName,omitempty" validate:"max=200"`
	FirstName   string `json:"firstName,omitempty" validate:"max=100"`
	LastName    string `json:"lastName,omitempty" validate:"max=100"`
	Email       string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone       string `json:"phone,omitempty" validate:"max=20"`
	Mobile      string `json:"mobile,omitempty" validate:"max=20"`
	Website     string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	TaxNumber   string `json:"taxNumber,omitempty" validate:"max=50"`
	Notes       string `json:"notes,omitempty"`
}

// CreateAddressRequest represents a request to create a contact address
type CreateAddressRequest struct {
	Type         domain.AddressType `json:"type" validate:"required,oneof=billing shipping office home other"`
	AddressLine1 string             `json:"addressLine1" validate:"required,max=200"`
	AddressLine2 string             `json:"addressLine2,omitempty" validate:"max=200"`
	City         string             `json:"city" validate:"required,max=100"`
	State        string             `json:"state,omitempty" validate:"max=100"`
	Country      string             `json:"country" validate:"required,max=100"`
	PostalCode   string             `json:"postalCode,omitempty" validate:"max=20"`
	IsPrimary    bool               `json:"isPrimary"`
}

// CreatePhoneRequest represents a request to create a contact phone
type CreatePhoneRequest struct {
	Type      domain.PhoneType `json:"type" validate:"required,oneof=work mobile home fax other"`
	Number    string           `json:"number" validate:"required,max=20"`
	Extension string           `json:"extension,omitempty" validate:"max=10"`
	IsPrimary bool             `json:"isPrimary"`
}

// UpdateAddressRequest represents a request to update a contact address
type UpdateAddressRequest = CreateAddressRequest

// UpdatePhoneRequest represents a request to update a contact phone
type UpdatePhoneRequest = CreatePhoneRequest

// ContactListResponse represents a paginated list of contacts
type ContactListResponse struct {
	Contacts   []*domain.Contact `json:"contacts"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int64             `json:"totalPages"`
}
