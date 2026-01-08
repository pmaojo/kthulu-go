// @kthulu:module:auth
package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Domain errors
var (
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotConfirmed  = errors.New("user email not confirmed")
	ErrInvalidRole       = errors.New("invalid role")
)

// User represents a system user with rich domain behavior.
type User struct {
	ID               uint       `json:"id"`
	Email            Email      `json:"email"`
	PasswordHash     string     `json:"-"`
	ConfirmedAt      *time.Time `json:"confirmedAt,omitempty"`
	ConfirmationCode string     `json:"-"`
	RoleID           uint       `json:"roleId"`
	Role             *Role      `json:"role,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// Email is a value object for email addresses
type Email struct {
	value string
}

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return Email{}, ErrInvalidEmail
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: email}, nil
}

// String returns the email as a string
func (e Email) String() string {
	return e.value
}

// Value returns the email value for database storage (implements driver.Valuer)
func (e Email) Value() (driver.Value, error) {
	return e.value, nil
}

// Scan implements sql.Scanner interface for database reading
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = Email{}
		return nil
	}

	switch v := value.(type) {
	case string:
		email, err := NewEmail(v)
		if err != nil {
			return err
		}
		*e = email
		return nil
	case []byte:
		email, err := NewEmail(string(v))
		if err != nil {
			return err
		}
		*e = email
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}
}

// MarshalJSON implements json.Marshaler
func (e Email) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, e.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (e *Email) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	emailStr := strings.Trim(string(data), `"`)
	email, err := NewEmail(emailStr)
	if err != nil {
		return err
	}
	*e = email
	return nil
}

var userValidator = validator.New()

// NewUser constructs a User with domain validations and business rules.
func NewUser(email, passwordHash string, roleID uint) (*User, error) {
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}

	if roleID == 0 {
		return nil, ErrInvalidRole
	}

	now := time.Now()
	user := &User{
		Email:        emailVO,
		PasswordHash: passwordHash,
		RoleID:       roleID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return user, nil
}

// IsConfirmed returns true if the user's email is confirmed
func (u *User) IsConfirmed() bool {
	return u.ConfirmedAt != nil
}

// Confirm marks the user's email as confirmed
func (u *User) Confirm() {
	now := time.Now()
	u.ConfirmedAt = &now
	u.ConfirmationCode = ""
	u.UpdatedAt = now
}

// UpdateEmail updates the user's email and marks as unconfirmed
func (u *User) UpdateEmail(newEmail string) error {
	emailVO, err := NewEmail(newEmail)
	if err != nil {
		return err
	}

	u.Email = emailVO
	u.ConfirmedAt = nil // Reset confirmation when email changes
	u.UpdatedAt = time.Now()

	return nil
}

// UpdatePassword updates the user's password hash
func (u *User) UpdatePassword(newPasswordHash string) error {
	if newPasswordHash == "" {
		return ErrInvalidPassword
	}

	u.PasswordHash = newPasswordHash
	u.UpdatedAt = time.Now()

	return nil
}

// UpdateRole updates the user's role
func (u *User) UpdateRole(roleID uint) error {
	if roleID == 0 {
		return ErrInvalidRole
	}

	u.RoleID = roleID
	u.UpdatedAt = time.Now()

	return nil
}

// CanLogin returns true if the user can log in (must be confirmed)
func (u *User) CanLogin() bool {
	return u.IsConfirmed()
}

// GetDisplayName returns a display name for the user (email for now)
func (u *User) GetDisplayName() string {
	return u.Email.String()
}
