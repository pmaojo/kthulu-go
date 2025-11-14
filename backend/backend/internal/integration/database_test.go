// @kthulu:core
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/testutils"
)

// TestDatabaseIntegration tests database operations across modules
func TestDatabaseIntegration(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	ctx := context.Background()

	t.Run("Basic Database Operations", func(t *testing.T) {
		// Test basic database connectivity
		var count int64
		err := testDB.WithContext(ctx).Raw("SELECT COUNT(*) FROM roles").Scan(&count).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1)) // Should have default role

		// Test inserting a user
		result := testDB.WithContext(ctx).Exec(
			"INSERT INTO users (email, password_hash, role_id) VALUES (?, ?, ?)",
			"integration@test.com", "hashedpassword", 1,
		)
		require.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		// Test retrieving the user
		var user struct {
			ID           uint   `json:"id"`
			Email        string `json:"email"`
			PasswordHash string `json:"password_hash"`
			RoleID       uint   `json:"role_id"`
		}

		err = testDB.WithContext(ctx).Raw(
			"SELECT id, email, password_hash, role_id FROM users WHERE email = ?",
			"integration@test.com",
		).Scan(&user).Error
		require.NoError(t, err)

		assert.Equal(t, "integration@test.com", user.Email)
		assert.Equal(t, "hashedpassword", user.PasswordHash)
		assert.Equal(t, uint(1), user.RoleID)
	})

	t.Run("Organization Operations", func(t *testing.T) {
		// Test inserting an organization
		result := testDB.WithContext(ctx).Exec(
			"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
			"Integration Test Organization", "integration-test-org", "Test organization", "company",
		)
		require.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		// Test retrieving the organization
		var org struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
		}

		err := testDB.WithContext(ctx).Raw(
			"SELECT id, name, slug, description FROM organizations WHERE slug = ?",
			"integration-test-org",
		).Scan(&org).Error
		require.NoError(t, err)

		assert.Equal(t, "Integration Test Organization", org.Name)
		assert.Equal(t, "integration-test-org", org.Slug)
		assert.Equal(t, "Test organization", org.Description)
	})

	t.Run("Refresh Token Operations", func(t *testing.T) {
		// Test inserting a refresh token
		result := testDB.WithContext(ctx).Exec(
			"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)",
			1, "test-refresh-token", time.Now().Add(24*time.Hour),
		)
		require.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		// Test retrieving the refresh token
		var token struct {
			ID        uint      `json:"id"`
			UserID    uint      `json:"user_id"`
			Token     string    `json:"token"`
			ExpiresAt time.Time `json:"expires_at"`
		}

		err := testDB.WithContext(ctx).Raw(
			"SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = ?",
			"test-refresh-token",
		).Scan(&token).Error
		require.NoError(t, err)

		assert.Equal(t, uint(1), token.UserID)
		assert.Equal(t, "test-refresh-token", token.Token)
		assert.True(t, token.ExpiresAt.After(time.Now()))
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		// Start transaction
		tx := testDB.Begin()
		require.NoError(t, tx.Error)

		// Insert organization in transaction
		result := tx.Exec(
			"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
			"Transaction Test Org", "transaction-test-org", "Test org for transaction", "company",
		)
		require.NoError(t, result.Error)

		// Verify organization exists in transaction
		var count int64
		err := tx.Raw("SELECT COUNT(*) FROM organizations WHERE slug = ?", "transaction-test-org").Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Rollback transaction
		tx.Rollback()

		// Verify organization doesn't exist after rollback
		err = testDB.Raw("SELECT COUNT(*) FROM organizations WHERE slug = ?", "transaction-test-org").Scan(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		// Insert initial organization
		result := testDB.WithContext(ctx).Exec(
			"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
			"Concurrent Test Org", "concurrent-test-org", "Initial description", "company",
		)
		require.NoError(t, result.Error)

		// Simulate concurrent updates
		done := make(chan bool, 2)

		// Goroutine 1: Update name
		go func() {
			defer func() { done <- true }()

			result := testDB.Exec(
				"UPDATE organizations SET name = ? WHERE slug = ?",
				"Updated by Goroutine 1", "concurrent-test-org",
			)
			assert.NoError(t, result.Error)
		}()

		// Goroutine 2: Update description
		go func() {
			defer func() { done <- true }()

			// Small delay to ensure some concurrency
			time.Sleep(10 * time.Millisecond)

			result := testDB.Exec(
				"UPDATE organizations SET description = ? WHERE slug = ?",
				"Updated by Goroutine 2", "concurrent-test-org",
			)
			assert.NoError(t, result.Error)
		}()

		// Wait for both goroutines to complete
		<-done
		<-done

		// Verify final state
		var org struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		err := testDB.Raw(
			"SELECT name, description FROM organizations WHERE slug = ?",
			"concurrent-test-org",
		).Scan(&org).Error
		require.NoError(t, err)

		// At least one update should have succeeded
		assert.True(t,
			org.Name == "Updated by Goroutine 1" || org.Description == "Updated by Goroutine 2",
			"At least one concurrent update should have succeeded")
	})

	t.Run("Database Constraints", func(t *testing.T) {
		// Insert first organization
		result := testDB.WithContext(ctx).Exec(
			"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
			"Constraint Test Org 1", "constraint-test-org", "First org", "company",
		)
		require.NoError(t, result.Error)

		// Try to insert another organization with same slug (should fail due to UNIQUE constraint)
		result = testDB.WithContext(ctx).Exec(
			"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
			"Constraint Test Org 2", "constraint-test-org", "Second org", "company",
		)
		assert.Error(t, result.Error, "Should not allow duplicate slugs")
	})

	t.Run("Pagination Integration", func(t *testing.T) {
		// Create multiple organizations for pagination testing
		for i := 0; i < 25; i++ {
			result := testDB.WithContext(ctx).Exec(
				"INSERT INTO organizations (name, slug, description, type) VALUES (?, ?, ?, ?)",
				fmt.Sprintf("Pagination Test Org %d", i),
				fmt.Sprintf("pagination-test-org-%d", i),
				fmt.Sprintf("Description %d", i),
				"company",
			)
			require.NoError(t, result.Error)
		}

		// Test pagination - get first 10
		var page1Orgs []struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		}

		err := testDB.Raw("SELECT id, name, slug FROM organizations WHERE slug LIKE 'pagination-test-org-%' ORDER BY id LIMIT 10 OFFSET 0").Scan(&page1Orgs).Error
		require.NoError(t, err)
		assert.Equal(t, 10, len(page1Orgs))

		// Test pagination - get next 10
		var page2Orgs []struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		}

		err = testDB.Raw("SELECT id, name, slug FROM organizations WHERE slug LIKE 'pagination-test-org-%' ORDER BY id LIMIT 10 OFFSET 10").Scan(&page2Orgs).Error
		require.NoError(t, err)
		assert.Equal(t, 10, len(page2Orgs))

		// Verify no overlap between pages
		page1IDs := make(map[uint]bool)
		for _, org := range page1Orgs {
			page1IDs[org.ID] = true
		}

		for _, org := range page2Orgs {
			assert.False(t, page1IDs[org.ID], "Organization should not appear in both pages")
		}
	})
}
