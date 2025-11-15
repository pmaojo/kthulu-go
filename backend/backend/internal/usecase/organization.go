// @kthulu:module:org
package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// OrganizationUseCase orchestrates organization management workflows.
type OrganizationUseCase struct {
	organizations repository.OrganizationRepository
	orgUsers      repository.OrganizationUserRepository
	invitations   repository.InvitationRepository
	users         repository.UserRepository
	notifier      repository.NotificationProvider
	logger        core.Logger
}

// NewOrganizationUseCase builds an OrganizationUseCase instance.
func NewOrganizationUseCase(
	organizations repository.OrganizationRepository,
	orgUsers repository.OrganizationUserRepository,
	invitations repository.InvitationRepository,
	users repository.UserRepository,
	notifier repository.NotificationProvider,
	logger core.Logger,
) *OrganizationUseCase {
	return &OrganizationUseCase{
		organizations: organizations,
		orgUsers:      orgUsers,
		invitations:   invitations,
		users:         users,
		notifier:      notifier,
		logger:        logger,
	}
}

// CreateOrganizationRequest contains the data needed to create an organization
type CreateOrganizationRequest struct {
	Name        string                  `json:"name" validate:"required,min=2,max=100"`
	Slug        string                  `json:"slug" validate:"required,min=2,max=50,alphanum"`
	Description string                  `json:"description,omitempty" validate:"max=500"`
	Type        domain.OrganizationType `json:"type" validate:"required,oneof=company nonprofit personal education"`
	Domain      string                  `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Website     string                  `json:"website,omitempty" validate:"omitempty,url"`
	Phone       string                  `json:"phone,omitempty" validate:"omitempty,e164"`
	Address     string                  `json:"address,omitempty" validate:"max=200"`
	City        string                  `json:"city,omitempty" validate:"max=100"`
	State       string                  `json:"state,omitempty" validate:"max=100"`
	Country     string                  `json:"country,omitempty" validate:"max=100"`
	PostalCode  string                  `json:"postalCode,omitempty" validate:"max=20"`
}

// CreateOrganization creates a new organization with the user as owner
func (u *OrganizationUseCase) CreateOrganization(ctx context.Context, userID uint, req CreateOrganizationRequest) (*domain.Organization, error) {
	u.logger.Info("Create organization request", "userId", userID, "name", req.Name, "slug", req.Slug)

	// Check if slug is already taken
	exists, err := u.organizations.ExistsBySlug(ctx, req.Slug)
	if err != nil {
		u.logger.Error("Failed to check organization slug existence", "slug", req.Slug, "error", err)
		return nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		u.logger.Warn("Organization creation attempted with existing slug", "slug", req.Slug, "userId", userID)
		return nil, domain.ErrOrganizationAlreadyExists
	}

	// Check if domain is already taken (if provided)
	if req.Domain != "" {
		exists, err := u.organizations.ExistsByDomain(ctx, req.Domain)
		if err != nil {
			u.logger.Error("Failed to check organization domain existence", "domain", req.Domain, "error", err)
			return nil, fmt.Errorf("failed to check domain existence: %w", err)
		}
		if exists {
			u.logger.Warn("Organization creation attempted with existing domain", "domain", req.Domain, "userId", userID)
			return nil, errors.New("domain already taken")
		}
	}

	// Create organization
	org, err := domain.NewOrganization(req.Name, req.Slug, req.Type, userID)
	if err != nil {
		u.logger.Error("Failed to create organization domain object", "error", err)
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Set additional fields
	org.Description = req.Description
	if req.Domain != "" {
		if err := org.SetDomain(req.Domain); err != nil {
			u.logger.Error("Failed to set organization domain", "domain", req.Domain, "error", err)
			return nil, fmt.Errorf("failed to set domain: %w", err)
		}
	}
	if err := org.UpdateContactInfo(req.Website, req.Phone, req.Address, req.City, req.State, req.Country, req.PostalCode); err != nil {
		u.logger.Error("Failed to set organization contact info", "error", err)
		return nil, fmt.Errorf("failed to set contact info: %w", err)
	}

	// Save organization
	if err := u.organizations.Create(ctx, org); err != nil {
		u.logger.Error("Failed to persist organization", "error", err)
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add user as owner
	orgUser, err := domain.NewOrganizationUser(org.ID, userID, domain.OrganizationRoleOwner)
	if err != nil {
		u.logger.Error("Failed to create organization user relationship", "error", err)
		return nil, fmt.Errorf("failed to create organization user: %w", err)
	}

	if err := u.orgUsers.Create(ctx, orgUser); err != nil {
		u.logger.Error("Failed to persist organization user relationship", "error", err)
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	u.logger.Info("Organization created successfully", "organizationId", org.ID, "userId", userID)
	return org, nil
}

// GetOrganization retrieves an organization by ID
func (u *OrganizationUseCase) GetOrganization(ctx context.Context, userID, organizationID uint) (*domain.Organization, error) {
	u.logger.Info("Get organization request", "userId", userID, "organizationId", organizationID)

	// Check if user is in organization
	inOrg, err := u.orgUsers.IsUserInOrganization(ctx, organizationID, userID)
	if err != nil {
		u.logger.Error("Failed to check user organization membership", "userId", userID, "organizationId", organizationID, "error", err)
		return nil, fmt.Errorf("failed to check organization membership: %w", err)
	}
	if !inOrg {
		u.logger.Warn("User attempted to access organization they're not a member of", "userId", userID, "organizationId", organizationID)
		return nil, domain.ErrUserNotInOrganization
	}

	// Get organization
	org, err := u.organizations.FindByID(ctx, organizationID)
	if err != nil {
		u.logger.Error("Failed to find organization", "organizationId", organizationID, "error", err)
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	u.logger.Info("Organization retrieved successfully", "organizationId", organizationID, "userId", userID)
	return org, nil
}

// UpdateOrganizationRequest contains the data needed to update an organization
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Domain      *string `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Website     *string `json:"website,omitempty" validate:"omitempty,url"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,e164"`
	Address     *string `json:"address,omitempty" validate:"omitempty,max=200"`
	City        *string `json:"city,omitempty" validate:"omitempty,max=100"`
	State       *string `json:"state,omitempty" validate:"omitempty,max=100"`
	Country     *string `json:"country,omitempty" validate:"omitempty,max=100"`
	PostalCode  *string `json:"postalCode,omitempty" validate:"omitempty,max=20"`
}

// UpdateOrganization updates an organization
func (u *OrganizationUseCase) UpdateOrganization(ctx context.Context, userID, organizationID uint, req UpdateOrganizationRequest) (*domain.Organization, error) {
	u.logger.Info("Update organization request", "userId", userID, "organizationId", organizationID)

	// Check if user can manage organization
	canManage, err := u.canManageOrganization(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		u.logger.Warn("User attempted to update organization without permissions", "userId", userID, "organizationId", organizationID)
		return nil, domain.ErrInsufficientPermissions
	}

	// Get organization
	org, err := u.organizations.FindByID(ctx, organizationID)
	if err != nil {
		u.logger.Error("Failed to find organization for update", "organizationId", organizationID, "error", err)
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// Track if any changes were made
	hasChanges := false

	// Update basic info if provided
	if req.Name != nil || req.Description != nil {
		name := org.Name
		description := org.Description

		if req.Name != nil {
			name = *req.Name
		}
		if req.Description != nil {
			description = *req.Description
		}

		if err := org.UpdateBasicInfo(name, description); err != nil {
			u.logger.Error("Failed to update organization basic info", "error", err)
			return nil, fmt.Errorf("failed to update basic info: %w", err)
		}
		hasChanges = true
	}

	// Update domain if provided
	if req.Domain != nil {
		// Check if domain is already taken by another organization
		if *req.Domain != "" && *req.Domain != org.Domain {
			exists, err := u.organizations.ExistsByDomain(ctx, *req.Domain)
			if err != nil {
				u.logger.Error("Failed to check domain existence", "domain", *req.Domain, "error", err)
				return nil, fmt.Errorf("failed to check domain existence: %w", err)
			}
			if exists {
				u.logger.Warn("Organization update attempted with existing domain", "domain", *req.Domain, "organizationId", organizationID)
				return nil, errors.New("domain already taken")
			}
		}

		if err := org.SetDomain(*req.Domain); err != nil {
			u.logger.Error("Failed to set organization domain", "domain", *req.Domain, "error", err)
			return nil, fmt.Errorf("failed to set domain: %w", err)
		}
		hasChanges = true
	}

	// Update contact info if any contact fields are provided
	if req.Website != nil || req.Phone != nil || req.Address != nil || req.City != nil || req.State != nil || req.Country != nil || req.PostalCode != nil {
		website := org.Website
		phone := org.Phone
		address := org.Address
		city := org.City
		state := org.State
		country := org.Country
		postalCode := org.PostalCode

		if req.Website != nil {
			website = *req.Website
		}
		if req.Phone != nil {
			phone = *req.Phone
		}
		if req.Address != nil {
			address = *req.Address
		}
		if req.City != nil {
			city = *req.City
		}
		if req.State != nil {
			state = *req.State
		}
		if req.Country != nil {
			country = *req.Country
		}
		if req.PostalCode != nil {
			postalCode = *req.PostalCode
		}

		if err := org.UpdateContactInfo(website, phone, address, city, state, country, postalCode); err != nil {
			u.logger.Error("Failed to update organization contact info", "error", err)
			return nil, fmt.Errorf("failed to update contact info: %w", err)
		}
		hasChanges = true
	}

	// Save changes if any were made
	if hasChanges {
		if err := u.organizations.Update(ctx, org); err != nil {
			u.logger.Error("Failed to persist organization changes", "organizationId", organizationID, "error", err)
			return nil, fmt.Errorf("failed to update organization: %w", err)
		}
	}

	u.logger.Info("Organization updated successfully", "organizationId", organizationID, "userId", userID)
	return org, nil
}

// ListUserOrganizations lists organizations for a user
func (u *OrganizationUseCase) ListUserOrganizations(ctx context.Context, userID uint) ([]*domain.Organization, error) {
	u.logger.Info("List user organizations request", "userId", userID)

	// Get user's organization relationships
	orgUsers, err := u.orgUsers.FindByUser(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to find user organizations", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user organizations: %w", err)
	}

	// Get organizations
	organizations := make([]*domain.Organization, 0, len(orgUsers))
	for _, orgUser := range orgUsers {
		org, err := u.organizations.FindByID(ctx, orgUser.OrganizationID)
		if err != nil {
			u.logger.Error("Failed to find organization", "organizationId", orgUser.OrganizationID, "error", err)
			continue // Skip this organization but continue with others
		}
		organizations = append(organizations, org)
	}

	u.logger.Info("User organizations retrieved", "userId", userID, "count", len(organizations))
	return organizations, nil
}

// InviteUserRequest contains the data needed to invite a user to an organization
type InviteUserRequest struct {
	Email   string                  `json:"email" validate:"required,email"`
	Role    domain.OrganizationRole `json:"role" validate:"required,oneof=admin member guest"`
	Message string                  `json:"message,omitempty" validate:"max=500"`
}

// InviteUser invites a user to join an organization
func (u *OrganizationUseCase) InviteUser(ctx context.Context, inviterID, organizationID uint, req InviteUserRequest) (*domain.Invitation, error) {
	u.logger.Info("Invite user request", "inviterId", inviterID, "organizationId", organizationID, "email", req.Email)

	// Check if inviter can invite users
	canInvite, err := u.canInviteUsers(ctx, inviterID, organizationID)
	if err != nil {
		return nil, err
	}
	if !canInvite {
		u.logger.Warn("User attempted to invite without permissions", "inviterId", inviterID, "organizationId", organizationID)
		return nil, domain.ErrInsufficientPermissions
	}

	// Check if user is already in organization
	existingUser, err := u.users.FindByEmail(ctx, req.Email)
	if err == nil {
		// User exists, check if already in organization
		inOrg, err := u.orgUsers.IsUserInOrganization(ctx, organizationID, existingUser.ID)
		if err != nil {
			u.logger.Error("Failed to check if user is in organization", "email", req.Email, "organizationId", organizationID, "error", err)
			return nil, fmt.Errorf("failed to check organization membership: %w", err)
		}
		if inOrg {
			u.logger.Warn("Invitation attempted for user already in organization", "email", req.Email, "organizationId", organizationID)
			return nil, errors.New("user is already a member of this organization")
		}
	}

	// Check if there's already a pending invitation
	exists, err := u.invitations.ExistsPendingByEmail(ctx, organizationID, req.Email)
	if err != nil {
		u.logger.Error("Failed to check pending invitation existence", "email", req.Email, "organizationId", organizationID, "error", err)
		return nil, fmt.Errorf("failed to check pending invitations: %w", err)
	}
	if exists {
		u.logger.Warn("Invitation attempted for email with pending invitation", "email", req.Email, "organizationId", organizationID)
		return nil, errors.New("invitation already pending for this email")
	}

	// Generate invitation token
	token, err := u.generateInvitationToken()
	if err != nil {
		u.logger.Error("Failed to generate invitation token", "error", err)
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Create invitation (expires in 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invitation, err := domain.NewInvitation(organizationID, inviterID, req.Email, req.Role, token, expiresAt)
	if err != nil {
		u.logger.Error("Failed to create invitation domain object", "error", err)
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	invitation.Message = req.Message

	// Save invitation
	if err := u.invitations.Create(ctx, invitation); err != nil {
		u.logger.Error("Failed to persist invitation", "error", err)
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Send invitation email via notification service
	link := fmt.Sprintf("https://example.com/invitations/accept?token=%s", invitation.Token)
	notifReq := repository.NotificationRequest{
		To:      req.Email,
		Subject: "You're invited to join an organization",
		Body:    fmt.Sprintf("You've been invited to join an organization. Click the link to accept: %s", link),
		Type:    repository.NotificationTypeInvitation,
		Data: map[string]interface{}{
			"token":          invitation.Token,
			"organizationId": organizationID,
		},
	}
	if err := u.notifier.SendNotification(ctx, notifReq); err != nil {
		u.logger.Error("Failed to send invitation email", "email", req.Email, "error", err)
	}

	u.logger.Info("User invitation created successfully", "invitationId", invitation.ID, "email", req.Email, "organizationId", organizationID)
	return invitation, nil
}

// AcceptInvitation accepts an invitation to join an organization
func (u *OrganizationUseCase) AcceptInvitation(ctx context.Context, userID uint, token string) (*domain.Organization, error) {
	u.logger.Info("Accept invitation request", "userId", userID, "token", token)

	// Find invitation by token
	invitation, err := u.invitations.FindByToken(ctx, token)
	if err != nil {
		u.logger.Error("Failed to find invitation by token", "token", token, "error", err)
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}

	// Get user to verify email matches
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		u.logger.Error("Failed to find user for invitation acceptance", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify email matches invitation
	if user.Email.String() != invitation.Email {
		u.logger.Warn("User attempted to accept invitation for different email", "userId", userID, "userEmail", user.Email.String(), "invitationEmail", invitation.Email)
		return nil, errors.New("invitation email does not match user email")
	}

	// Check if user is already in organization
	inOrg, err := u.orgUsers.IsUserInOrganization(ctx, invitation.OrganizationID, userID)
	if err != nil {
		u.logger.Error("Failed to check organization membership", "userId", userID, "organizationId", invitation.OrganizationID, "error", err)
		return nil, fmt.Errorf("failed to check organization membership: %w", err)
	}
	if inOrg {
		u.logger.Warn("User attempted to accept invitation for organization they're already in", "userId", userID, "organizationId", invitation.OrganizationID)
		return nil, errors.New("user is already a member of this organization")
	}

	// Accept invitation
	if err := invitation.Accept(); err != nil {
		u.logger.Error("Failed to accept invitation", "invitationId", invitation.ID, "error", err)
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	// Add user to organization
	orgUser, err := domain.NewOrganizationUser(invitation.OrganizationID, userID, invitation.Role)
	if err != nil {
		u.logger.Error("Failed to create organization user relationship", "error", err)
		return nil, fmt.Errorf("failed to create organization user: %w", err)
	}

	if err := u.orgUsers.Create(ctx, orgUser); err != nil {
		u.logger.Error("Failed to persist organization user relationship", "error", err)
		return nil, fmt.Errorf("failed to add user to organization: %w", err)
	}

	// Update invitation
	if err := u.invitations.Update(ctx, invitation); err != nil {
		u.logger.Error("Failed to update invitation", "invitationId", invitation.ID, "error", err)
		return nil, fmt.Errorf("failed to update invitation: %w", err)
	}

	// Get organization
	org, err := u.organizations.FindByID(ctx, invitation.OrganizationID)
	if err != nil {
		u.logger.Error("Failed to find organization after invitation acceptance", "organizationId", invitation.OrganizationID, "error", err)
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	u.logger.Info("Invitation accepted successfully", "userId", userID, "organizationId", invitation.OrganizationID)
	return org, nil
}

// canManageOrganization checks if a user can manage an organization
func (u *OrganizationUseCase) canManageOrganization(ctx context.Context, userID, organizationID uint) (bool, error) {
	role, err := u.orgUsers.GetUserRole(ctx, organizationID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotInOrganization) {
			return false, nil
		}
		return false, err
	}

	return role == domain.OrganizationRoleOwner || role == domain.OrganizationRoleAdmin, nil
}

// canInviteUsers checks if a user can invite users to an organization
func (u *OrganizationUseCase) canInviteUsers(ctx context.Context, userID, organizationID uint) (bool, error) {
	role, err := u.orgUsers.GetUserRole(ctx, organizationID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotInOrganization) {
			return false, nil
		}
		return false, err
	}

	return role == domain.OrganizationRoleOwner || role == domain.OrganizationRoleAdmin, nil
}

// generateInvitationToken generates a secure random token for invitations
func (u *OrganizationUseCase) generateInvitationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
