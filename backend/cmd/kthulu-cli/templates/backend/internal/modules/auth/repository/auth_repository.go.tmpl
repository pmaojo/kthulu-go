package repository

import (
	"gorm.io/gorm"
	"{{.ModuleName}}/internal/modules/auth/domain"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domain.AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}
