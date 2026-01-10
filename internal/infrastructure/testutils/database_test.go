package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer CleanupTestDB(t, db)

	_, err = sqlDB.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", "user@example.com", "hash")
	require.NoError(t, err)

	var count int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
