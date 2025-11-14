// @kthulu:core
package testutils

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"backend/core"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *gorm.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = gormDB.Exec(`
                CREATE TABLE IF NOT EXISTS roles (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        name TEXT UNIQUE NOT NULL,
                        description TEXT,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS permissions (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        name TEXT UNIQUE NOT NULL,
                        description TEXT,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS role_permissions (
                        role_model_id INTEGER NOT NULL,
                        permission_model_id INTEGER NOT NULL,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        PRIMARY KEY (role_model_id, permission_model_id)
                );

                CREATE TABLE IF NOT EXISTS users (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        email TEXT UNIQUE NOT NULL,
                        password_hash TEXT NOT NULL,
                        role_id INTEGER DEFAULT 1,
                        confirmed_at DATETIME,
                        confirmation_code TEXT,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS organizations (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        name TEXT NOT NULL,
                        slug TEXT UNIQUE NOT NULL,
                        description TEXT,
                        type TEXT,
                        domain TEXT,
                        logo_url TEXT,
                        website TEXT,
                        phone TEXT,
                        address TEXT,
                        city TEXT,
                        state TEXT,
                        country TEXT,
                        postal_code TEXT,
                        is_active INTEGER DEFAULT 1,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS contacts (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        organization_id INTEGER NOT NULL,
                        type TEXT,
                        company_name TEXT,
                        first_name TEXT,
                        last_name TEXT,
                        email TEXT,
                        phone TEXT,
                        mobile TEXT,
                        website TEXT,
                        tax_number TEXT,
                        notes TEXT,
                        is_active INTEGER DEFAULT 1,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS refresh_tokens (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        user_id INTEGER NOT NULL,
                        token TEXT UNIQUE NOT NULL,
                        expires_at DATETIME NOT NULL,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                );

                INSERT OR IGNORE INTO roles (id, name, description) VALUES (1, 'user', 'Default user role');
        `).Error
	require.NoError(t, err)

	return gormDB
}

// CleanupTestDB closes the test database connection
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Close()
		require.NoError(t, err)
	}
}

// SetupTestDBWithFile creates a test database with a temporary file
func SetupTestDBWithFile(t *testing.T) (*sql.DB, string) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	dbPath := tmpFile.Name()

	// Create test configuration
	cfg := &core.Config{
		Database: core.DatabaseConfig{
			URL:    dbPath,
			Driver: "sqlite",
		},
	}

	// Create logger for tests
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create database connection
	db, err := core.NewDB(cfg, logger)
	require.NoError(t, err)

	// Run migrations
	err = core.Migrate(db, logger)
	require.NoError(t, err)

	return db, dbPath
}

// CleanupTestDBWithFile closes the database and removes the file
func CleanupTestDBWithFile(t *testing.T, db *sql.DB, dbPath string) {
	if db != nil {
		err := db.Close()
		require.NoError(t, err)
	}

	if dbPath != "" && dbPath != ":memory:" {
		err := os.Remove(dbPath)
		require.NoError(t, err)
	}
}

// CreateTestUser creates a test user for use in tests
func CreateTestUser(t *testing.T, db *sql.DB, email string) uint {
	query := `INSERT INTO users (email, password_hash, created_at, updated_at) 
			  VALUES (?, ?, datetime('now'), datetime('now'))`

	result, err := db.Exec(query, email, "hashedpassword")
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return uint(id)
}

// CreateTestOrganization creates a test organization for use in tests
func CreateTestOrganization(t *testing.T, db *sql.DB, name, slug string) uint {
	query := `INSERT INTO organizations (name, slug, created_at, updated_at) 
			  VALUES (?, ?, datetime('now'), datetime('now'))`

	result, err := db.Exec(query, name, slug)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return uint(id)
}

// CreateTestProduct creates a test product for use in tests
func CreateTestProduct(t *testing.T, db *sql.DB, orgID uint, name, sku string, price float64) uint {
	query := `INSERT INTO products (organization_id, name, sku, price, is_active, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))`

	result, err := db.Exec(query, orgID, name, sku, price)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return uint(id)
}

// CreateTestContact creates a test contact for use in tests
func CreateTestContact(t *testing.T, db *sql.DB, orgID uint, name, email string) uint {
	query := `INSERT INTO contacts (organization_id, name, email, type, created_at, updated_at) 
			  VALUES (?, ?, ?, 'customer', datetime('now'), datetime('now'))`

	result, err := db.Exec(query, orgID, name, email)
	require.NoError(t, err)

	id, err := result.LastInsertId()
	require.NoError(t, err)

	return uint(id)
}

// AssertDatabaseState provides utilities for asserting database state in tests
type AssertDatabaseState struct {
	db *sql.DB
	t  *testing.T
}

// NewAssertDatabaseState creates a new database state asserter
func NewAssertDatabaseState(t *testing.T, db *sql.DB) *AssertDatabaseState {
	return &AssertDatabaseState{db: db, t: t}
}

// UserExists checks if a user exists with the given email
func (a *AssertDatabaseState) UserExists(email string) bool {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	require.NoError(a.t, err)
	return count > 0
}

// OrganizationExists checks if an organization exists with the given slug
func (a *AssertDatabaseState) OrganizationExists(slug string) bool {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM organizations WHERE slug = ?", slug).Scan(&count)
	require.NoError(a.t, err)
	return count > 0
}

// ProductExists checks if a product exists with the given SKU in an organization
func (a *AssertDatabaseState) ProductExists(orgID uint, sku string) bool {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM products WHERE organization_id = ? AND sku = ?", orgID, sku).Scan(&count)
	require.NoError(a.t, err)
	return count > 0
}

// ContactExists checks if a contact exists with the given email in an organization
func (a *AssertDatabaseState) ContactExists(orgID uint, email string) bool {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM contacts WHERE organization_id = ? AND email = ?", orgID, email).Scan(&count)
	require.NoError(a.t, err)
	return count > 0
}

// CountRecords returns the number of records in a table
func (a *AssertDatabaseState) CountRecords(tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := a.db.QueryRow(query).Scan(&count)
	require.NoError(a.t, err)
	return count
}

// TableExists checks if a table exists in the database
func (a *AssertDatabaseState) TableExists(tableName string) bool {
	var count int
	err := a.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	require.NoError(a.t, err)
	return count > 0
}
