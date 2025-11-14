// @kthulu:core
package contracts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/domain"
	db "backend/internal/infrastructure/db"
	"backend/internal/repository"
	"backend/internal/testutils"
)

// TestRepositoryContracts verifies that all repository implementations satisfy their interfaces
func TestRepositoryContracts(t *testing.T) {
	// Test interface compliance at compile time
	t.Run("InterfaceCompliance", func(t *testing.T) {
		// User repository
		var _ repository.UserRepository = (*db.UserRepository)(nil)

		// Auth repositories
		var _ repository.RefreshTokenRepository = (*db.RefreshTokenRepository)(nil)
		var _ repository.RoleRepository = (*db.RoleRepository)(nil)

		// Organization repository
		var _ repository.OrganizationRepository = (*db.OrganizationRepository)(nil)

		// Contact repository
		var _ repository.ContactRepository = (*db.ContactRepository)(nil)

		// Product repository
		var _ repository.ProductRepository = (*db.ProductRepository)(nil)

		// Invoice repository
		var _ repository.InvoiceRepository = (*db.InvoiceRepository)(nil)

		// Note: Inventory and Calendar repositories are not yet fully implemented
		// var _ repository.InventoryRepository = (*db.InventoryRepository)(nil)
		// var _ repository.CalendarRepository = (*db.CalendarRepository)(nil)
	})
}

// TestUserRepositoryContract tests the user repository contract
func TestUserRepositoryContract(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := db.NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		email, err := domain.NewEmail("test@example.com")
		require.NoError(t, err)

		user := &domain.User{
			Email:        email,
			PasswordHash: "hashedpassword",
			RoleID:       1, // Default role ID
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)

		retrieved, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email.String(), retrieved.Email.String())
		assert.Equal(t, user.PasswordHash, retrieved.PasswordHash)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		email, err := domain.NewEmail("email@example.com")
		require.NoError(t, err)

		user := &domain.User{
			Email:        email,
			PasswordHash: "hashedpassword",
			RoleID:       1,
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.FindByEmail(ctx, user.Email.String())
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email.String(), retrieved.Email.String())
	})

	t.Run("Update", func(t *testing.T) {
		email, err := domain.NewEmail("update@example.com")
		require.NoError(t, err)

		user := &domain.User{
			Email:        email,
			PasswordHash: "hashedpassword",
			RoleID:       1,
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)

		err = user.UpdateEmail("updated@example.com")
		require.NoError(t, err)

		err = repo.Update(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", retrieved.Email.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, 99999)
		assert.Error(t, err)

		_, err = repo.FindByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})
}

// TestOrganizationRepositoryContract tests the organization repository contract
func TestOrganizationRepositoryContract(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := db.NewOrganizationRepository(testDB)
	ctx := context.Background()

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		org := &domain.Organization{
			Name:        "Test Organization",
			Slug:        "test-org",
			Description: "Test description",
			Type:        domain.OrganizationTypeCompany,
		}

		err := repo.Create(ctx, org)
		require.NoError(t, err)
		assert.NotZero(t, org.ID)

		retrieved, err := repo.FindByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Equal(t, org.Name, retrieved.Name)
		assert.Equal(t, org.Slug, retrieved.Slug)
	})

	t.Run("FindBySlug", func(t *testing.T) {
		org := &domain.Organization{
			Name: "Slug Test Org",
			Slug: "slug-test-org",
			Type: domain.OrganizationTypeCompany,
		}

		err := repo.Create(ctx, org)
		require.NoError(t, err)

		retrieved, err := repo.FindBySlug(ctx, org.Slug)
		require.NoError(t, err)
		assert.Equal(t, org.ID, retrieved.ID)
		assert.Equal(t, org.Name, retrieved.Name)
	})

	t.Run("List", func(t *testing.T) {
		// Create test organizations
		for i := 0; i < 3; i++ {
			org := &domain.Organization{
				Name: fmt.Sprintf("List Test Org %d", i),
				Slug: fmt.Sprintf("list-test-org-%d", i),
				Type: domain.OrganizationTypeCompany,
			}
			err := repo.Create(ctx, org)
			require.NoError(t, err)
		}

		orgs, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orgs), 3)
	})
}

// TestContactRepositoryContract tests the contact repository contract
func TestContactRepositoryContract(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := db.NewContactRepository(testDB)
	ctx := context.Background()

	// Create test organization first
	orgRepo := db.NewOrganizationRepository(testDB)
	org := &domain.Organization{
		Name: "Test Org for Contacts",
		Slug: "test-org-contacts",
		Type: domain.OrganizationTypeCompany,
	}
	err := orgRepo.Create(ctx, org)
	require.NoError(t, err)

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		contact := &domain.Contact{
			OrganizationID: org.ID,
			CompanyName:    "Test Contact Company",
			Email:          "contact@example.com",
			Type:           domain.ContactTypeCustomer,
			IsActive:       true,
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)
		assert.NotZero(t, contact.ID)

		retrieved, err := repo.GetByID(ctx, org.ID, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.CompanyName, retrieved.CompanyName)
		assert.Equal(t, contact.Email, retrieved.Email)
		assert.Equal(t, contact.Type, retrieved.Type)
	})

	t.Run("List", func(t *testing.T) {
		// Create test contacts
		for i := 0; i < 3; i++ {
			contact := &domain.Contact{
				OrganizationID: org.ID,
				CompanyName:    fmt.Sprintf("List Contact %d", i),
				Email:          fmt.Sprintf("list%d@example.com", i),
				Type:           domain.ContactTypeCustomer,
				IsActive:       true,
			}
			err := repo.Create(ctx, contact)
			require.NoError(t, err)
		}

		filters := repository.DefaultContactFilters()
		contacts, total, err := repo.List(ctx, org.ID, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contacts), 3)
		assert.GreaterOrEqual(t, total, int64(3))
	})
}

// TestProductRepositoryContract tests the product repository contract
func TestProductRepositoryContract(t *testing.T) {
	t.Skip("Product repository contract tests require complex setup - skipping for now")
}

// TestInventoryRepositoryContract tests the inventory repository contract
func TestInventoryRepositoryContract(t *testing.T) {
	t.Skip("Inventory repository not yet fully implemented")
}

// TestCalendarRepositoryContract tests the calendar repository contract
func TestCalendarRepositoryContract(t *testing.T) {
	t.Skip("Calendar repository not yet fully implemented")
}

// TestRoleRepositoryContract tests the role repository contract
func TestRoleRepositoryContract(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := db.NewRoleRepository(testDB)
	ctx := context.Background()

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		role := &domain.Role{
			Name:        "Test Role",
			Description: "Test role description",
		}

		err := repo.Create(ctx, role)
		require.NoError(t, err)
		assert.NotZero(t, role.ID)

		retrieved, err := repo.FindByID(ctx, role.ID)
		require.NoError(t, err)
		assert.Equal(t, role.Name, retrieved.Name)
		assert.Equal(t, role.Description, retrieved.Description)
	})

	t.Run("FindByName", func(t *testing.T) {
		role := &domain.Role{
			Name:        "Unique Role Name",
			Description: "Test role",
		}

		err := repo.Create(ctx, role)
		require.NoError(t, err)

		retrieved, err := repo.FindByName(ctx, role.Name)
		require.NoError(t, err)
		assert.Equal(t, role.ID, retrieved.ID)
		assert.Equal(t, role.Name, retrieved.Name)
	})

	t.Run("List", func(t *testing.T) {
		// Create test roles
		for i := 0; i < 3; i++ {
			role := &domain.Role{
				Name:        fmt.Sprintf("List Role %d", i),
				Description: fmt.Sprintf("Test role %d", i),
			}
			err := repo.Create(ctx, role)
			require.NoError(t, err)
		}

		roles, err := repo.List(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(roles), 3)
	})
}

// TestRefreshTokenRepositoryContract tests the refresh token repository contract
func TestRefreshTokenRepositoryContract(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := db.NewRefreshTokenRepository(testDB)
	ctx := context.Background()

	// Create test user first
	userRepo := db.NewUserRepository(testDB)
	email, err := domain.NewEmail("token@example.com")
	require.NoError(t, err)

	user := &domain.User{
		Email:        email,
		PasswordHash: "hashedpassword",
		RoleID:       1,
	}
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("CreateAndRetrieve", func(t *testing.T) {
		token := &domain.RefreshToken{
			UserID:    user.ID,
			Token:     "test-refresh-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := repo.Create(ctx, token)
		require.NoError(t, err)
		assert.NotZero(t, token.ID)

		retrieved, err := repo.FindByToken(ctx, token.Token)
		require.NoError(t, err)
		assert.Equal(t, token.UserID, retrieved.UserID)
		assert.Equal(t, token.Token, retrieved.Token)
	})

	t.Run("DeleteByUserID", func(t *testing.T) {
		token := &domain.RefreshToken{
			UserID:    user.ID,
			Token:     "delete-test-token",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		err := repo.Create(ctx, token)
		require.NoError(t, err)

		err = repo.DeleteByUserID(ctx, user.ID)
		require.NoError(t, err)

		_, err = repo.FindByToken(ctx, token.Token)
		assert.Error(t, err)
	})

	t.Run("DeleteExpired", func(t *testing.T) {
		// Create expired token
		expiredToken := &domain.RefreshToken{
			UserID:    user.ID,
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-time.Hour),
		}

		err := repo.Create(ctx, expiredToken)
		require.NoError(t, err)

		_, err = repo.DeleteExpired(ctx)
		require.NoError(t, err)

		_, err = repo.FindByToken(ctx, expiredToken.Token)
		assert.Error(t, err)
	})
}

// TestInvoiceRepositoryContract tests the invoice repository contract
func TestInvoiceRepositoryContract(t *testing.T) {
	t.Skip("Invoice repository contract tests require complex setup - skipping for now")
}
