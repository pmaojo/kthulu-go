// @kthulu:module:org
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Domain errors for organization module
var (
	ErrOrganizationNotFound      = errors.New("organization not found")
	ErrOrganizationAlreadyExists = errors.New("organization already exists")
	ErrInvalidOrganizationName   = errors.New("invalid organization name")
	ErrInvalidDomain             = errors.New("invalid domain")
	ErrUserNotInOrganization     = errors.New("user not in organization")
	ErrInvitationNotFound        = errors.New("invitation not found")
	ErrInvitationExpired         = errors.New("invitation expired")
	ErrInvitationAlreadyAccepted = errors.New("invitation already accepted")
	ErrInsufficientPermissions   = errors.New("insufficient permissions")
)

// OrganizationType represents the type of organization
type OrganizationType string

const (
	OrganizationTypeCompany   OrganizationType = "company"
	OrganizationTypeNonProfit OrganizationType = "nonprofit"
	OrganizationTypePersonal  OrganizationType = "personal"
	OrganizationTypeEducation OrganizationType = "education"
)

// Organization represents a tenant organization in the system
type Organization struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name" validate:"required,min=2,max=100"`
	Slug        string           `json:"slug" validate:"required,min=2,max=50,alphanum"`
	Description string           `json:"description,omitempty" validate:"max=500"`
	Type        OrganizationType `json:"type" validate:"required,oneof=company nonprofit personal education"`
	Domain      string           `json:"domain,omitempty" validate:"omitempty,fqdn"`
	LogoURL     string           `json:"logoUrl,omitempty" validate:"omitempty,url"`
	Website     string           `json:"website,omitempty" validate:"omitempty,url"`
	Phone       string           `json:"phone,omitempty" validate:"omitempty,e164"`
	Address     string           `json:"address,omitempty" validate:"max=200"`
	City        string           `json:"city,omitempty" validate:"max=100"`
	State       string           `json:"state,omitempty" validate:"max=100"`
	Country     string           `json:"country,omitempty" validate:"max=100"`
	PostalCode  string           `json:"postalCode,omitempty" validate:"max=20"`
	IsActive    bool             `json:"isActive"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`

	// Relationships
	Users []OrganizationUser `json:"users,omitempty"`
}

// OrganizationRole represents the role of a user within an organization
type OrganizationRole string

const (
	OrganizationRoleOwner  OrganizationRole = "owner"
	OrganizationRoleAdmin  OrganizationRole = "admin"
	OrganizationRoleMember OrganizationRole = "member"
	OrganizationRoleGuest  OrganizationRole = "guest"
)

// OrganizationUser represents the relationship between a user and an organization
type OrganizationUser struct {
	ID             uint             `json:"id"`
	OrganizationID uint             `json:"organizationId" validate:"required"`
	UserID         uint             `json:"userId" validate:"required"`
	Role           OrganizationRole `json:"role" validate:"required,oneof=owner admin member guest"`
	JoinedAt       time.Time        `json:"joinedAt"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`

	// Relationships
	Organization *Organization `json:"organization,omitempty"`
	User         *User         `json:"user,omitempty"`
}

// InvitationStatus represents the status of an invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// Invitation represents an invitation to join an organization
type Invitation struct {
	ID             uint             `json:"id"`
	OrganizationID uint             `json:"organizationId" validate:"required"`
	InviterID      uint             `json:"inviterId" validate:"required"`
	Email          string           `json:"email" validate:"required,email"`
	Role           OrganizationRole `json:"role" validate:"required,oneof=admin member guest"`
	Token          string           `json:"token" validate:"required"`
	Status         InvitationStatus `json:"status" validate:"required,oneof=pending accepted declined expired"`
	Message        string           `json:"message,omitempty" validate:"max=500"`
	ExpiresAt      time.Time        `json:"expiresAt"`
	AcceptedAt     *time.Time       `json:"acceptedAt,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`

	// Relationships
	Organization *Organization `json:"organization,omitempty"`
	Inviter      *User         `json:"inviter,omitempty"`
}

var orgValidator = validator.New()

// NewOrganization creates a new organization with validation
func NewOrganization(name, slug string, orgType OrganizationType, ownerID uint) (*Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(strings.ToLower(slug))

	if name == "" {
		return nil, ErrInvalidOrganizationName
	}

	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidOrganizationName
	}

	if slug == "" || len(slug) < 2 || len(slug) > 50 {
		return nil, errors.New("invalid organization slug")
	}

	// Validate organization type
	switch orgType {
	case OrganizationTypeCompany, OrganizationTypeNonProfit, OrganizationTypePersonal, OrganizationTypeEducation:
		// Valid types
	default:
		return nil, errors.New("invalid organization type")
	}

	now := time.Now()
	org := &Organization{
		Name:      name,
		Slug:      slug,
		Type:      orgType,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate the organization
	if err := orgValidator.Struct(org); err != nil {
		return nil, err
	}

	return org, nil
}

// UpdateBasicInfo updates the basic information of the organization
func (o *Organization) UpdateBasicInfo(name, description string) error {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" || len(name) < 2 || len(name) > 100 {
		return ErrInvalidOrganizationName
	}

	if len(description) > 500 {
		return errors.New("description too long")
	}

	o.Name = name
	o.Description = description
	o.UpdatedAt = time.Now()

	return orgValidator.Struct(o)
}

// UpdateContactInfo updates the contact information of the organization
func (o *Organization) UpdateContactInfo(website, phone, address, city, state, country, postalCode string) error {
	o.Website = strings.TrimSpace(website)
	o.Phone = strings.TrimSpace(phone)
	o.Address = strings.TrimSpace(address)
	o.City = strings.TrimSpace(city)
	o.State = strings.TrimSpace(state)
	o.Country = strings.TrimSpace(country)
	o.PostalCode = strings.TrimSpace(postalCode)
	o.UpdatedAt = time.Now()

	return orgValidator.Struct(o)
}

// SetDomain sets the organization's domain
func (o *Organization) SetDomain(domain string) error {
	domain = strings.TrimSpace(strings.ToLower(domain))
	o.Domain = domain
	o.UpdatedAt = time.Now()

	return orgValidator.Struct(o)
}

// Deactivate deactivates the organization
func (o *Organization) Deactivate() {
	o.IsActive = false
	o.UpdatedAt = time.Now()
}

// Activate activates the organization
func (o *Organization) Activate() {
	o.IsActive = true
	o.UpdatedAt = time.Now()
}

// NewOrganizationUser creates a new organization user relationship
func NewOrganizationUser(organizationID, userID uint, role OrganizationRole) (*OrganizationUser, error) {
	if organizationID == 0 {
		return nil, errors.New("organization ID is required")
	}

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}

	// Validate role
	switch role {
	case OrganizationRoleOwner, OrganizationRoleAdmin, OrganizationRoleMember, OrganizationRoleGuest:
		// Valid roles
	default:
		return nil, errors.New("invalid organization role")
	}

	now := time.Now()
	orgUser := &OrganizationUser{
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           role,
		JoinedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return orgUser, orgValidator.Struct(orgUser)
}

// UpdateRole updates the user's role in the organization
func (ou *OrganizationUser) UpdateRole(newRole OrganizationRole) error {
	// Validate role
	switch newRole {
	case OrganizationRoleOwner, OrganizationRoleAdmin, OrganizationRoleMember, OrganizationRoleGuest:
		// Valid roles
	default:
		return errors.New("invalid organization role")
	}

	ou.Role = newRole
	ou.UpdatedAt = time.Now()

	return nil
}

// CanManageUsers returns true if the user can manage other users in the organization
func (ou *OrganizationUser) CanManageUsers() bool {
	return ou.Role == OrganizationRoleOwner || ou.Role == OrganizationRoleAdmin
}

// CanInviteUsers returns true if the user can invite other users to the organization
func (ou *OrganizationUser) CanInviteUsers() bool {
	return ou.Role == OrganizationRoleOwner || ou.Role == OrganizationRoleAdmin
}

// CanManageOrganization returns true if the user can manage organization settings
func (ou *OrganizationUser) CanManageOrganization() bool {
	return ou.Role == OrganizationRoleOwner || ou.Role == OrganizationRoleAdmin
}

// NewInvitation creates a new invitation
func NewInvitation(organizationID, inviterID uint, email string, role OrganizationRole, token string, expiresAt time.Time) (*Invitation, error) {
	if organizationID == 0 {
		return nil, errors.New("organization ID is required")
	}

	if inviterID == 0 {
		return nil, errors.New("inviter ID is required")
	}

	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Validate role (owners cannot be invited, they must be promoted)
	switch role {
	case OrganizationRoleAdmin, OrganizationRoleMember, OrganizationRoleGuest:
		// Valid roles for invitation
	default:
		return nil, errors.New("invalid role for invitation")
	}

	if token == "" {
		return nil, errors.New("invitation token is required")
	}

	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expiration time must be in the future")
	}

	now := time.Now()
	invitation := &Invitation{
		OrganizationID: organizationID,
		InviterID:      inviterID,
		Email:          emailVO.String(),
		Role:           role,
		Token:          token,
		Status:         InvitationStatusPending,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return invitation, orgValidator.Struct(invitation)
}

// Accept marks the invitation as accepted
func (i *Invitation) Accept() error {
	if i.Status != InvitationStatusPending {
		return ErrInvitationAlreadyAccepted
	}

	if time.Now().After(i.ExpiresAt) {
		return ErrInvitationExpired
	}

	now := time.Now()
	i.Status = InvitationStatusAccepted
	i.AcceptedAt = &now
	i.UpdatedAt = now

	return nil
}

// Decline marks the invitation as declined
func (i *Invitation) Decline() error {
	if i.Status != InvitationStatusPending {
		return errors.New("invitation cannot be declined")
	}

	i.Status = InvitationStatusDeclined
	i.UpdatedAt = time.Now()

	return nil
}

// IsExpired returns true if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt) || i.Status == InvitationStatusExpired
}

// MarkExpired marks the invitation as expired
func (i *Invitation) MarkExpired() {
	i.Status = InvitationStatusExpired
	i.UpdatedAt = time.Now()
}
