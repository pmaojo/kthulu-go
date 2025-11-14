// @kthulu:module:user
package usecase

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"backend/core"
	"backend/internal/domain"
	"backend/internal/repository"
)

// UserUseCase orchestrates user profile management workflows.
type UserUseCase struct {
	users  repository.UserRepository
	roles  repository.RoleRepository
	logger core.Logger
}

// UserUseCaseParams defines dependencies for UserUseCase using named dependencies.
// NewUserUseCase builds a UserUseCase instance.
func NewUserUseCase(
	users repository.UserRepository,
	roles repository.RoleRepository,
	logger core.Logger,
) *UserUseCase {
	return &UserUseCase{
		users:  users,
		roles:  roles,
		logger: logger,
	}
}

// GetProfile retrieves user profile information by user ID.
func (u *UserUseCase) GetProfile(ctx context.Context, userID uint) (*domain.User, error) {
	ctx, span := startUseCaseSpan(ctx, "UserUseCase.GetProfile")
	defer span.End()

	u.logger.Info("Get user profile request", "userId", userID)

	// Find user by ID
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			u.logger.Warn("Profile request for non-existent user", "userId", userID)
			return nil, domain.ErrUserNotFound
		}
		u.logger.Error("Failed to find user for profile", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Load user role
	role, err := u.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		u.logger.Error("Failed to load user role for profile", "userId", userID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	u.logger.Info("User profile retrieved", "userId", userID)
	return user, nil
}

// UpdateProfileRequest contains the data needed to update a user profile
type UpdateProfileRequest struct {
	Email           *string `json:"email,omitempty" validate:"omitempty,email"`
	Password        *string `json:"password,omitempty" validate:"omitempty,min=8"`
	CurrentPassword *string `json:"currentPassword,omitempty"`
}

// UpdateProfile updates user profile information.
func (u *UserUseCase) UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*domain.User, error) {
	ctx, span := startUseCaseSpan(ctx, "UserUseCase.UpdateProfile")
	defer span.End()

	u.logger.Info("Update user profile request", "userId", userID)

	// Find user by ID
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			u.logger.Warn("Profile update for non-existent user", "userId", userID)
			return nil, domain.ErrUserNotFound
		}
		u.logger.Error("Failed to find user for profile update", "userId", userID, "error", err)
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Track if any changes were made
	hasChanges := false

	// Update email if provided
	if req.Email != nil && *req.Email != user.Email.String() {
		// Check if new email is already taken
		exists, err := u.users.ExistsByEmail(ctx, *req.Email)
		if err != nil {
			u.logger.Error("Failed to check email existence", "email", *req.Email, "error", err)
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			u.logger.Warn("Profile update attempted with existing email", "email", *req.Email, "userId", userID)
			return nil, domain.ErrUserAlreadyExists
		}

		// Update email using domain method (this will reset confirmation)
		if err := user.UpdateEmail(*req.Email); err != nil {
			u.logger.Error("Failed to update user email", "userId", userID, "email", *req.Email, "error", err)
			return nil, fmt.Errorf("failed to update email: %w", err)
		}

		hasChanges = true
		u.logger.Info("User email updated", "userId", userID, "newEmail", *req.Email)
	}

	// Update password if provided
	if req.Password != nil {
		// Validate current password if provided
		if req.CurrentPassword != nil {
			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(*req.CurrentPassword)); err != nil {
				u.logger.Warn("Invalid current password during profile update", "userId", userID)
				return nil, errors.New("invalid current password")
			}
		}

		// Validate new password strength
		if err := u.validatePassword(*req.Password); err != nil {
			u.logger.Warn("Password validation failed during profile update", "userId", userID, "error", err)
			return nil, err
		}

		// Hash new password
		hashed, err := u.hashPassword(*req.Password)
		if err != nil {
			u.logger.Error("Failed to hash new password", "userId", userID, "error", err)
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		// Update password using domain method
		if err := user.UpdatePassword(hashed); err != nil {
			u.logger.Error("Failed to update user password", "userId", userID, "error", err)
			return nil, fmt.Errorf("failed to update password: %w", err)
		}

		hasChanges = true
		u.logger.Info("User password updated", "userId", userID)
	}

	// Save changes if any were made
	if hasChanges {
		if err := u.users.Update(ctx, user); err != nil {
			u.logger.Error("Failed to persist user profile changes", "userId", userID, "error", err)
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Load user role for response
	role, err := u.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		u.logger.Error("Failed to load user role after profile update", "userId", userID, "roleId", user.RoleID, "error", err)
		return nil, fmt.Errorf("failed to load user role: %w", err)
	}
	user.Role = role

	u.logger.Info("User profile updated successfully", "userId", userID)
	return user, nil
}

// hashPassword hashes a password using bcrypt
func (u *UserUseCase) hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

// validatePassword validates password strength
func (u *UserUseCase) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Add more password validation rules as needed
	// For example: require uppercase, lowercase, numbers, special characters

	return nil
}
