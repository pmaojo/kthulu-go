package mcp_test

import (
	"context"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func TestDatabaseServiceSQLiteSchemaAndQuery(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	service := mcp.NewDatabaseService()

	setup := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT NOT NULL, active INTEGER DEFAULT 1)",
		"CREATE UNIQUE INDEX idx_users_email ON users(email)",
		"INSERT INTO users (email) VALUES ('first@example.com'), ('second@example.com')",
	}
	for _, statement := range setup {
		_, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: statement, AllowWrite: true})
		require.NoError(t, err)
	}

	schema, err := service.Schema(ctx, dir, mcp.DbSchemaArgs{Driver: "sqlite", DSN: "app.db"})
	require.NoError(t, err)
	require.Contains(t, schema, "TABLE users")
	require.Contains(t, schema, "email TEXT NOT NULL")
	require.Contains(t, schema, "id INTEGER PRIMARY KEY")
	require.Contains(t, schema, "UNIQUE INDEX idx_users_email")

	result, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "SELECT email FROM users ORDER BY id"})
	require.NoError(t, err)
	require.Contains(t, result, "first@example.com")
	require.Contains(t, result, "second@example.com")
	require.Contains(t, result, "2 row(s)")
}

func TestDatabaseServiceQueryIsReadOnlyByDefault(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	service := mcp.NewDatabaseService()

	_, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "CREATE TABLE t (id INTEGER)"})
	require.ErrorContains(t, err, "allow_write")

	_, err = service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "DELETE FROM t"})
	require.ErrorContains(t, err, "allow_write")
}

func TestDatabaseServiceRejectsUnknownDriver(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	service := mcp.NewDatabaseService()

	_, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "oracle", DSN: "whatever", SQL: "SELECT 1"})
	require.ErrorContains(t, err, "unsupported driver")
}

func TestDatabaseServiceQueryMaxRows(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	service := mcp.NewDatabaseService()

	_, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "CREATE TABLE n (v INTEGER)", AllowWrite: true})
	require.NoError(t, err)
	_, err = service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "INSERT INTO n VALUES (1), (2), (3)", AllowWrite: true})
	require.NoError(t, err)

	result, err := service.Query(ctx, dir, mcp.DbQueryArgs{Driver: "sqlite", DSN: "app.db", SQL: "SELECT v FROM n", MaxRows: 2})
	require.NoError(t, err)
	require.Contains(t, result, "2 row(s)")
	require.Contains(t, result, "truncated")
}
