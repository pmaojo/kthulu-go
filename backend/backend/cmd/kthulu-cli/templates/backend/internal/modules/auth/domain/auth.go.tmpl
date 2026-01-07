package domain

import (
	"time"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex"`
	Password  string    `json:"-"` // Hash
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthRepository interface {
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
}

type AuthService interface {
	Register(email, password string) (*User, error)
	Login(email, password string) (string, error)
}
