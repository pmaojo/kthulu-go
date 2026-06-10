package commands

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/jinzhu/inflection"
	"github.com/spf13/cobra"
)

var (
	diffDryRun bool
	diffName   string
)

var migrateDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Generate a migration from the gap between your entities and the database",
	Long: `Compare the entity structs in your modules against the live database
schema and generate a goose SQL migration for the difference.

Only additive changes are generated (CREATE TABLE, ADD COLUMN). Columns that
exist in the database but not in any entity are reported as comments — they
are never dropped automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entities, err := collectEntityTables(".")
		if err != nil {
			return err
		}
		if len(entities) == 0 {
			return fmt.Errorf("no entity structs found under internal/modules or internal/adapters/http/modules")
		}

		driver, dsn, err := resolveConsoleDB()
		if err != nil {
			return err
		}
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer db.Close()
		if err := db.Ping(); err != nil {
			return fmt.Errorf("connect to database (%s): %w", driver, err)
		}

		schema, err := introspectSchema(db, driver)
		if err != nil {
			return fmt.Errorf("introspect schema: %w", err)
		}

		up, down, notes := diffSchema(entities, schema, driver)
		if len(up) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "✅ Database schema is up to date with your entities.")
			for _, n := range notes {
				fmt.Fprintln(cmd.OutOrStdout(), n)
			}
			return nil
		}

		migration := renderGooseMigration(up, down, notes)
		if diffDryRun {
			fmt.Fprintln(cmd.OutOrStdout(), migration)
			return nil
		}

		name := diffName
		if name == "" {
			name = "schema_diff"
		}
		if err := os.MkdirAll("migrations", 0o755); err != nil {
			return err
		}
		file := filepath.Join("migrations", fmt.Sprintf("%s_%s.sql", time.Now().UTC().Format("20060102150405"), name))
		if err := os.WriteFile(file, []byte(migration), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✅ Wrote %s (%d statement(s))\n", file, len(up))
		fmt.Fprintln(cmd.OutOrStdout(), "   Review it, then apply with: kthulu migrate up")
		for _, n := range notes {
			fmt.Fprintln(cmd.OutOrStdout(), n)
		}
		return nil
	},
}

func init() {
	migrateDiffCmd.Flags().BoolVar(&diffDryRun, "dry-run", false, "Print the migration instead of writing a file")
	migrateDiffCmd.Flags().StringVar(&diffName, "name", "", "Migration name suffix (default schema_diff)")
	migrateCmd.AddCommand(migrateDiffCmd)
}

// entityColumn is one column derived from an entity struct field.
type entityColumn struct {
	Name    string
	SQLType string
	IsPK    bool
}

// entityTable is the SQL shape of one entity struct.
type entityTable struct {
	Table   string
	Columns []entityColumn
}

// collectEntityTables parses entity structs from the project's module core
// packages and derives their SQL table shapes.
func collectEntityTables(root string) ([]entityTable, error) {
	var files []string
	for _, base := range []string{
		filepath.Join(root, "internal", "modules"),
		filepath.Join(root, "internal", "adapters", "http", "modules"),
	} {
		matches, _ := filepath.Glob(filepath.Join(base, "*", "core", "*.go"))
		files = append(files, matches...)
	}

	var tables []entityTable
	seen := map[string]bool{}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		fileTables, err := parseEntityFile(file)
		if err != nil {
			fmt.Printf("⚠️  Skipping %s: %v\n", file, err)
			continue
		}
		for _, t := range fileTables {
			if !seen[strings.ToLower(t.Table)] {
				seen[strings.ToLower(t.Table)] = true
				tables = append(tables, t)
			}
		}
	}
	sort.Slice(tables, func(i, j int) bool { return tables[i].Table < tables[j].Table })
	return tables, nil
}

func parseEntityFile(path string) ([]entityTable, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	// TableName() overrides: map struct name -> table name.
	tableNames := map[string]string{}
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "TableName" || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := exprTypeName(fn.Recv.List[0].Type)
		if recv == "" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
				if lit, ok := ret.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					tableNames[recv] = strings.Trim(lit.Value, `"`)
				}
			}
			return true
		})
	}

	var tables []entityTable
	ast.Inspect(node, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := spec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		cols := structColumns(structType)
		if cols == nil {
			return false
		}
		table := tableNames[spec.Name.Name]
		if table == "" {
			table = inflection.Plural(toSnake(spec.Name.Name))
		}
		tables = append(tables, entityTable{Table: table, Columns: cols})
		return false
	})
	return tables, nil
}

// structColumns derives SQL columns from a struct, or nil if the struct does
// not look like a persisted entity (no ID field).
func structColumns(structType *ast.StructType) []entityColumn {
	var cols []entityColumn
	hasID := false
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue // embedded
		}
		name := field.Names[0].Name
		if !ast.IsExported(name) {
			continue
		}

		gormTag := ""
		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			gormTag = reflect.StructTag(tag).Get("gorm")
		}
		if gormTag == "-" {
			continue
		}

		goType := exprTypeName(field.Type)
		sqlType, ok := goTypeToSQL(goType)
		if !ok {
			continue // relations, slices, nested structs
		}

		column := gormColumnName(gormTag)
		if column == "" {
			column = toSnake(name)
		}

		isPK := name == "ID" || strings.Contains(gormTag, "primaryKey")
		if isPK {
			hasID = true
		}
		cols = append(cols, entityColumn{Name: column, SQLType: sqlType, IsPK: isPK})
	}
	if !hasID {
		return nil
	}
	return cols
}

func gormColumnName(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func exprTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			return pkg.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + exprTypeName(t.X)
	}
	return ""
}

func goTypeToSQL(goType string) (string, bool) {
	switch goType {
	case "string":
		return "TEXT", true
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "INTEGER", true
	case "float32", "float64":
		return "REAL", true
	case "bool":
		return "BOOLEAN", true
	case "time.Time", "*time.Time":
		return "TIMESTAMP", true
	default:
		return "", false
	}
}

func toSnake(s string) string {
	var out []rune
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			// Avoid underscores inside acronyms (ID -> id, UserID -> user_id).
			if i > 0 && (s[i-1] < 'A' || s[i-1] > 'Z') {
				out = append(out, '_')
			}
			r += 32
		}
		out = append(out, r)
	}
	return string(out)
}

// introspectSchema returns table -> set of column names (lowercased).
func introspectSchema(db *sql.DB, driver string) (map[string]map[string]bool, error) {
	schema := map[string]map[string]bool{}

	var query string
	switch driver {
	case "pgx":
		query = "SELECT table_name, column_name FROM information_schema.columns WHERE table_schema = 'public'"
	case "mysql":
		query = "SELECT table_name, column_name FROM information_schema.columns WHERE table_schema = DATABASE()"
	default:
		// SQLite: list tables, then pragma each.
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
		if err != nil {
			return nil, err
		}
		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				rows.Close()
				return nil, err
			}
			names = append(names, name)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
		for _, name := range names {
			cols := map[string]bool{}
			crows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%q)", name))
			if err != nil {
				return nil, err
			}
			for crows.Next() {
				var cid int
				var cname, ctype string
				var notnull, pk int
				var dflt interface{}
				if err := crows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
					crows.Close()
					return nil, err
				}
				cols[strings.ToLower(cname)] = true
			}
			crows.Close()
			if err := crows.Err(); err != nil {
				return nil, err
			}
			schema[strings.ToLower(name)] = cols
		}
		return schema, nil
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, err
		}
		key := strings.ToLower(table)
		if schema[key] == nil {
			schema[key] = map[string]bool{}
		}
		schema[key][strings.ToLower(column)] = true
	}
	return schema, rows.Err()
}

// diffSchema computes additive Up/Down statements plus advisory notes.
func diffSchema(entities []entityTable, schema map[string]map[string]bool, driver string) (up, down, notes []string) {
	for _, entity := range entities {
		existing, tableExists := schema[strings.ToLower(entity.Table)]
		if !tableExists {
			up = append(up, createTableSQL(entity, driver))
			down = append(down, fmt.Sprintf("DROP TABLE IF EXISTS %q;", entity.Table))
			continue
		}
		entityCols := map[string]bool{}
		for _, col := range entity.Columns {
			entityCols[strings.ToLower(col.Name)] = true
			if !existing[strings.ToLower(col.Name)] {
				up = append(up, fmt.Sprintf("ALTER TABLE %q ADD COLUMN %q %s;", entity.Table, col.Name, col.SQLType))
				down = append(down, fmt.Sprintf("ALTER TABLE %q DROP COLUMN %q;", entity.Table, col.Name))
			}
		}
		var extras []string
		for col := range existing {
			if !entityCols[col] {
				extras = append(extras, col)
			}
		}
		sort.Strings(extras)
		for _, col := range extras {
			notes = append(notes, fmt.Sprintf("-- note: column %s.%s exists in the database but not in any entity (not dropped)", entity.Table, col))
		}
	}
	return up, down, notes
}

func createTableSQL(entity entityTable, driver string) string {
	var defs []string
	for _, col := range entity.Columns {
		if col.IsPK {
			switch driver {
			case "pgx":
				defs = append(defs, fmt.Sprintf("%q SERIAL PRIMARY KEY", col.Name))
			case "mysql":
				defs = append(defs, fmt.Sprintf("%q INTEGER PRIMARY KEY AUTO_INCREMENT", col.Name))
			default:
				defs = append(defs, fmt.Sprintf("%q INTEGER PRIMARY KEY AUTOINCREMENT", col.Name))
			}
			continue
		}
		defs = append(defs, fmt.Sprintf("%q %s", col.Name, col.SQLType))
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %q (\n    %s\n);", entity.Table, strings.Join(defs, ",\n    "))
}

func renderGooseMigration(up, down, notes []string) string {
	var b strings.Builder
	b.WriteString("-- Generated by 'kthulu migrate diff'. Review before applying.\n")
	for _, n := range notes {
		b.WriteString(n + "\n")
	}
	b.WriteString("-- +goose Up\n")
	for _, stmt := range up {
		b.WriteString(stmt + "\n")
	}
	b.WriteString("\n-- +goose Down\n")
	// Reverse order so dependent drops happen sanely.
	for i := len(down) - 1; i >= 0; i-- {
		b.WriteString(down[i] + "\n")
	}
	return b.String()
}
