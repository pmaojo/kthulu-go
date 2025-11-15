package db

import (
	"context"
	"testing"
	"time"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
	"github.com/pmaojo/kthulu-go/backend/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryFindPaginated_SortByEmail(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	emails := []string{"b@example.com", "a@example.com"}
	for _, e := range emails {
		emailVO, err := domain.NewEmail(e)
		require.NoError(t, err)
		user := &domain.User{
			Email:        emailVO,
			PasswordHash: "hash",
			RoleID:       1,
		}
		require.NoError(t, repo.Create(ctx, user))
	}

	params := repository.NewPaginationParams(1, 10, "email", "asc")
	result, err := repo.FindPaginated(ctx, params)
	require.NoError(t, err)
	require.Len(t, result.Data, 2)
	assert.Equal(t, "a@example.com", result.Data[0].Email.String())
	assert.Equal(t, "b@example.com", result.Data[1].Email.String())
}

func TestUserRepositoryFindPaginated_InvalidSortDefaults(t *testing.T) {
	testDB := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(t, testDB)

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	emailOld, err := domain.NewEmail("old@example.com")
	require.NoError(t, err)
	olderTime := time.Now().Add(-time.Hour)
	userOld := &domain.User{
		Email:        emailOld,
		PasswordHash: "hash",
		RoleID:       1,
		CreatedAt:    olderTime,
		UpdatedAt:    olderTime,
	}
	require.NoError(t, repo.Create(ctx, userOld))

	emailNew, err := domain.NewEmail("new@example.com")
	require.NoError(t, err)
	newerTime := time.Now()
	userNew := &domain.User{
		Email:        emailNew,
		PasswordHash: "hash",
		RoleID:       1,
		CreatedAt:    newerTime,
		UpdatedAt:    newerTime,
	}
	require.NoError(t, repo.Create(ctx, userNew))

	params := repository.NewPaginationParams(1, 10, "invalid_field", "asc")
	result, err := repo.FindPaginated(ctx, params)
	require.NoError(t, err)
	require.Len(t, result.Data, 2)
	assert.Equal(t, "new@example.com", result.Data[0].Email.String())
	assert.Equal(t, "old@example.com", result.Data[1].Email.String())
}
