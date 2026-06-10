package commands

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEntityFixture(t *testing.T, root string) {
	t.Helper()
	dir := filepath.Join(root, "internal", "modules", "product", "core")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package core

import "time"

type Product struct {
	ID           uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt    time.Time ` + "`json:\"created_at\"`" + `
	Name         string    ` + "`json:\"name\" gorm:\"column:name\"`" + `
	Price        int       ` + "`json:\"price\" gorm:\"column:price\"`" + `
	WeightKg     float64   ` + "`json:\"weight_kg\" gorm:\"column:weight_kg\"`" + `
	Hidden       string    ` + "`json:\"-\" gorm:\"-\"`" + `
	Relation     *Other    ` + "`json:\"relation,omitempty\" gorm:\"foreignKey:RelationID\"`" + `
}

func (Product) TableName() string { return "Products" }

type Other struct {
	ID   uint   ` + "`gorm:\"primaryKey\"`" + `
	Name string
}

type notAnEntity struct {
	Foo string
}
`
	if err := os.WriteFile(filepath.Join(dir, "product.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectEntityTables(t *testing.T) {
	root := t.TempDir()
	writeEntityFixture(t, root)

	tables, err := collectEntityTables(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 entities, got %d: %+v", len(tables), tables)
	}

	var product entityTable
	for _, tab := range tables {
		if tab.Table == "Products" {
			product = tab
		}
	}
	if product.Table == "" {
		t.Fatalf("Products table not found in %+v", tables)
	}

	cols := map[string]string{}
	for _, c := range product.Columns {
		cols[c.Name] = c.SQLType
	}
	if cols["name"] != "TEXT" || cols["price"] != "INTEGER" || cols["weight_kg"] != "REAL" || cols["created_at"] != "TIMESTAMP" {
		t.Fatalf("unexpected columns: %v", cols)
	}
	if _, ok := cols["hidden"]; ok {
		t.Fatal("gorm:\"-\" fields must be skipped")
	}
	if _, ok := cols["relation"]; ok {
		t.Fatal("relation fields must be skipped")
	}

	// Default table naming for entities without TableName(): snake plural.
	var other entityTable
	for _, tab := range tables {
		if tab.Table == "others" {
			other = tab
		}
	}
	if other.Table == "" {
		t.Fatalf("expected default table name 'others', got %+v", tables)
	}
}

func TestDiffSchemaAdditive(t *testing.T) {
	root := t.TempDir()
	writeEntityFixture(t, root)
	entities, err := collectEntityTables(root)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", filepath.Join(root, "diff.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Existing Products table missing weight_kg, with one legacy column.
	if _, err := db.Exec(`CREATE TABLE Products (id INTEGER PRIMARY KEY, created_at TIMESTAMP, name TEXT, price INTEGER, legacy TEXT)`); err != nil {
		t.Fatal(err)
	}

	schema, err := introspectSchema(db, "sqlite")
	if err != nil {
		t.Fatal(err)
	}

	up, down, notes := diffSchema(entities, schema, "sqlite")
	joined := strings.Join(up, "\n")

	if !strings.Contains(joined, `ALTER TABLE "Products" ADD COLUMN "weight_kg" REAL;`) {
		t.Fatalf("missing add column, got:\n%s", joined)
	}
	if !strings.Contains(joined, `CREATE TABLE IF NOT EXISTS "others"`) {
		t.Fatalf("missing create table, got:\n%s", joined)
	}
	if strings.Contains(joined, "DROP") {
		t.Fatalf("up statements must be additive, got:\n%s", joined)
	}
	if len(down) != 2 {
		t.Fatalf("expected 2 down statements, got %v", down)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "legacy") {
		t.Fatalf("expected legacy column note, got %v", notes)
	}

	// Applying the up statements must converge to an empty diff.
	for _, stmt := range up {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("apply %q: %v", stmt, err)
		}
	}
	schema, err = introspectSchema(db, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	up, _, _ = diffSchema(entities, schema, "sqlite")
	if len(up) != 0 {
		t.Fatalf("expected empty diff after applying, got %v", up)
	}
}
