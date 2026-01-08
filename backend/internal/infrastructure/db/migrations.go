package db

import "gorm.io/gorm"

// AutoMigrateModels runs GORM AutoMigrate for all database models.
func AutoMigrateModels(db *gorm.DB) error {
	models := []any{
		&UserModel{},
		&RoleModel{},
		&PermissionModel{},
		&RefreshTokenModel{},
		&organizationModel{},
		&organizationUserModel{},
		&invitationModel{},
		&contactModel{},
		&contactAddressModel{},
		&contactPhoneModel{},
	}
	return db.AutoMigrate(models...)
}
