package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"

	// Database drivers: pure-Go SQLite and PostgreSQL via pgx.
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// DatabaseService exposes schema introspection and ad-hoc querying for the
// project's SQLite or PostgreSQL databases.
type DatabaseService struct{}

// NewDatabaseService creates a new DatabaseService.
func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// GetTools returns all database tools bound to the working directory.
func (s *DatabaseService) GetTools(workingDir string) []RegisteredTool {
	return []RegisteredTool{
		s.schemaTool(workingDir),
		s.queryTool(workingDir),
	}
}

// DbSchemaArgs defines arguments for schema introspection.
type DbSchemaArgs struct {
	Driver string `json:"driver" jsonschema:"description=Database driver: 'sqlite' or 'postgres',required"`
	DSN    string `json:"dsn" jsonschema:"description=Connection string. For sqlite: path to the database file (relative to project root). For postgres: a standard connection URL or DSN.,required"`
	Table  string `json:"table,omitempty" jsonschema:"description=Limit output to a single table"`
}

// DbQueryArgs defines arguments for running an SQL query.
type DbQueryArgs struct {
	Driver     string `json:"driver" jsonschema:"description=Database driver: 'sqlite' or 'postgres',required"`
	DSN        string `json:"dsn" jsonschema:"description=Connection string. For sqlite: path to the database file (relative to project root). For postgres: a standard connection URL or DSN.,required"`
	SQL        string `json:"sql" jsonschema:"description=SQL statement to execute,required"`
	MaxRows    int    `json:"max_rows,omitempty" jsonschema:"description=Maximum rows to return (default 100)"`
	AllowWrite bool   `json:"allow_write,omitempty" jsonschema:"description=Allow statements that modify data (INSERT/UPDATE/DELETE/DDL). Read-only by default."`
}

func (s *DatabaseService) schemaTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "db_schema",
		Description: "Introspect a SQLite or PostgreSQL database schema: tables, columns with types/nullability, primary keys, and indexes.",
		Handler: func(ctx context.Context, args DbSchemaArgs) (*mcp_golang.ToolResponse, error) {
			schema, err := s.Schema(ctx, workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(schema)), nil
		},
	}
}

func (s *DatabaseService) queryTool(workingDir string) RegisteredTool {
	return RegisteredTool{
		Name:        "db_query",
		Description: "Run an SQL query against a SQLite or PostgreSQL database and return the rows. Read-only unless allow_write=true.",
		Handler: func(ctx context.Context, args DbQueryArgs) (*mcp_golang.ToolResponse, error) {
			result, err := s.Query(ctx, workingDir, args)
			if err != nil {
				return nil, err
			}
			return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
		},
	}
}

// Schema introspects the database schema.
func (s *DatabaseService) Schema(ctx context.Context, workingDir string, args DbSchemaArgs) (string, error) {
	db, driver, err := s.open(resolveWorkdir(workingDir), args.Driver, args.DSN)
	if err != nil {
		return "", err
	}
	defer db.Close()

	switch driver {
	case "sqlite":
		return s.sqliteSchema(ctx, db, args.Table)
	default:
		return s.postgresSchema(ctx, db, args.Table)
	}
}

// Query runs an SQL statement and renders the result set.
func (s *DatabaseService) Query(ctx context.Context, workingDir string, args DbQueryArgs) (string, error) {
	query := strings.TrimSpace(args.SQL)
	if query == "" {
		return "", fmt.Errorf("argument 'sql' is required")
	}

	if !args.AllowWrite && !isReadOnlyStatement(query) {
		return "", fmt.Errorf("statement appears to modify data; set allow_write=true to run it")
	}

	db, _, err := s.open(resolveWorkdir(workingDir), args.Driver, args.DSN)
	if err != nil {
		return "", err
	}
	defer db.Close()

	maxRows := args.MaxRows
	if maxRows <= 0 {
		maxRows = 100
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	return renderRows(rows, maxRows)
}

// open validates the driver, resolves SQLite paths, and opens a connection.
func (s *DatabaseService) open(workingDir, driver, dsn string) (*sql.DB, string, error) {
	driver = strings.ToLower(strings.TrimSpace(driver))
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, "", fmt.Errorf("argument 'dsn' is required")
	}

	switch driver {
	case "sqlite", "sqlite3":
		path := dsn
		if !strings.HasPrefix(path, "file:") && !strings.HasPrefix(path, ":memory:") {
			resolved, err := resolveWorkspacePath(workingDir, path)
			if err != nil {
				return nil, "", err
			}
			path = resolved
		}
		db, err := sql.Open("sqlite", path)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open sqlite database: %w", err)
		}
		return db, "sqlite", nil
	case "postgres", "postgresql", "pgx":
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return nil, "", fmt.Errorf("failed to open postgres database: %w", err)
		}
		return db, "postgres", nil
	default:
		return nil, "", fmt.Errorf("unsupported driver %q (use 'sqlite' or 'postgres')", driver)
	}
}

func (s *DatabaseService) sqliteSchema(ctx context.Context, db *sql.DB, table string) (string, error) {
	query := "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", err
		}
		if table == "" || name == table {
			tables = append(tables, name)
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(tables) == 0 {
		return "No matching tables found", nil
	}

	var builder strings.Builder
	for _, name := range tables {
		builder.WriteString(fmt.Sprintf("TABLE %s\n", name))

		colRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%q)", name))
		if err != nil {
			return "", fmt.Errorf("failed to describe %s: %w", name, err)
		}
		for colRows.Next() {
			var cid int
			var colName, colType string
			var notNull, pk int
			var dflt sql.NullString
			if err := colRows.Scan(&cid, &colName, &colType, &notNull, &dflt, &pk); err != nil {
				colRows.Close()
				return "", err
			}
			attrs := []string{}
			if pk > 0 {
				attrs = append(attrs, "PRIMARY KEY")
			}
			if notNull == 1 {
				attrs = append(attrs, "NOT NULL")
			}
			if dflt.Valid {
				attrs = append(attrs, "DEFAULT "+dflt.String)
			}
			builder.WriteString(fmt.Sprintf("  %s %s %s\n", colName, colType, strings.Join(attrs, " ")))
		}
		colRows.Close()

		idxRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA index_list(%q)", name))
		if err == nil {
			for idxRows.Next() {
				var seq int
				var idxName, origin string
				var unique, partial int
				if err := idxRows.Scan(&seq, &idxName, &unique, &origin, &partial); err != nil {
					break
				}
				uniqueLabel := ""
				if unique == 1 {
					uniqueLabel = "UNIQUE "
				}
				builder.WriteString(fmt.Sprintf("  %sINDEX %s\n", uniqueLabel, idxName))
			}
			idxRows.Close()
		}
		builder.WriteString("\n")
	}

	return truncateOutput(builder.String()), nil
}

func (s *DatabaseService) postgresSchema(ctx context.Context, db *sql.DB, table string) (string, error) {
	query := `
		SELECT table_name, column_name, data_type, is_nullable, COALESCE(column_default, '')
		FROM information_schema.columns
		WHERE table_schema = 'public' AND ($1 = '' OR table_name = $1)
		ORDER BY table_name, ordinal_position`
	rows, err := db.QueryContext(ctx, query, table)
	if err != nil {
		return "", fmt.Errorf("failed to introspect schema: %w", err)
	}
	defer rows.Close()

	var builder strings.Builder
	currentTable := ""
	count := 0
	for rows.Next() {
		var tableName, columnName, dataType, isNullable, columnDefault string
		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnDefault); err != nil {
			return "", err
		}
		count++
		if tableName != currentTable {
			if currentTable != "" {
				builder.WriteString("\n")
			}
			builder.WriteString(fmt.Sprintf("TABLE %s\n", tableName))
			currentTable = tableName
		}
		attrs := []string{}
		if isNullable == "NO" {
			attrs = append(attrs, "NOT NULL")
		}
		if columnDefault != "" {
			attrs = append(attrs, "DEFAULT "+columnDefault)
		}
		builder.WriteString(fmt.Sprintf("  %s %s %s\n", columnName, dataType, strings.Join(attrs, " ")))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if count == 0 {
		return "No matching tables found in schema 'public'", nil
	}

	idxRows, err := db.QueryContext(ctx,
		`SELECT tablename, indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' AND ($1 = '' OR tablename = $1) ORDER BY tablename, indexname`, table)
	if err == nil {
		defer idxRows.Close()
		first := true
		for idxRows.Next() {
			var tableName, indexName, indexDef string
			if err := idxRows.Scan(&tableName, &indexName, &indexDef); err != nil {
				break
			}
			if first {
				builder.WriteString("\nINDEXES\n")
				first = false
			}
			builder.WriteString(fmt.Sprintf("  %s ON %s: %s\n", indexName, tableName, indexDef))
		}
	}

	return truncateOutput(builder.String()), nil
}

// renderRows formats a result set as aligned text rows.
func renderRows(rows *sql.Rows, maxRows int) (string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString(strings.Join(columns, " | "))
	builder.WriteString("\n")
	builder.WriteString(strings.Repeat("-", len(strings.Join(columns, " | "))))
	builder.WriteString("\n")

	values := make([]any, len(columns))
	pointers := make([]any, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	count := 0
	truncated := false
	for rows.Next() {
		if count >= maxRows {
			truncated = true
			break
		}
		if err := rows.Scan(pointers...); err != nil {
			return "", err
		}
		cells := make([]string, len(columns))
		for i, value := range values {
			cells[i] = renderCell(value)
		}
		builder.WriteString(strings.Join(cells, " | "))
		builder.WriteString("\n")
		count++
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	builder.WriteString(fmt.Sprintf("\n%d row(s)", count))
	if truncated {
		builder.WriteString(" (truncated; increase max_rows for more)")
	}
	return truncateOutput(builder.String()), nil
}

func renderCell(value any) string {
	switch v := value.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// isReadOnlyStatement reports whether an SQL statement is safe to run without
// allow_write. It is a guardrail against accidents, not a security boundary.
func isReadOnlyStatement(query string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(query))
	for _, prefix := range []string{"select", "with", "show", "explain", "pragma", "describe", "values", "table"} {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}
