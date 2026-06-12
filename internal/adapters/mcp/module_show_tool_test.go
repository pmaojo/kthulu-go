package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildFakeProject creates a minimal fake kthulu project structure under dir
// for use in module_show tests.
func buildFakeProject(t *testing.T, dir string) {
	t.Helper()

	mustMkdir := func(p string) {
		t.Helper()
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
	}
	mustWrite := func(p, content string) {
		t.Helper()
		mustMkdir(filepath.Dir(p))
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	// model.go
	mustWrite(filepath.Join(dir, "internal", "order", "model.go"), `package order

import "time"

type Order struct {
	ID         uint      `+"`"+`gorm:"primaryKey" json:"id"`+"`"+`
	TotalCents int       `+"`"+`gorm:"column:total_cents" json:"total_cents" validate:"min=0"`+"`"+`
	Status     string    `+"`"+`json:"status" validate:"oneof=pending|paid|shipped"`+"`"+`
	UserID     uint      `+"`"+`gorm:"index" json:"user_id"`+"`"+`
	CreatedAt  time.Time `+"`"+`json:"created_at"`+"`"+`
	UpdatedAt  time.Time `+"`"+`json:"updated_at"`+"`"+`
}
`)

	// service.go
	mustWrite(filepath.Join(dir, "internal", "order", "service.go"), `package order

import "context"

type CreateOrderInput struct{}

func Create(ctx context.Context, input *CreateOrderInput) (*Order, error) {
	return nil, nil
}

func FindByID(ctx context.Context, id uint) (*Order, error) {
	return nil, nil
}

func UpdateStatus(ctx context.Context, id uint, status string) error {
	return nil
}

func unexported(ctx context.Context) {}
`)

	// handler.go
	mustWrite(filepath.Join(dir, "internal", "order", "handler.go"), `package order

func Create() {}
func FindAll() {}
func FindByID() {}
func Update() {}
func Delete() {}
func helper() {}
`)

	// migrations
	mustWrite(filepath.Join(dir, "migrations", "001_create_orders.sql"), "CREATE TABLE orders ();")
	mustWrite(filepath.Join(dir, "migrations", "002_alter_users.sql"), "ALTER TABLE users ADD COLUMN foo TEXT;")
}

func TestFindModuleDir_ExactMatch(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	modDir, name, err := findModuleDir(dir, "order")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "order" {
		t.Fatalf("expected name 'order', got %q", name)
	}
	want := filepath.Join(dir, "internal", "order")
	if modDir != want {
		t.Fatalf("expected dir %s, got %s", want, modDir)
	}
}

func TestFindModuleDir_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	_, name, err := findModuleDir(dir, "ORDER")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "order" {
		t.Fatalf("expected canonical name 'order', got %q", name)
	}
}

func TestFindModuleDir_NotFound(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	_, _, err := findModuleDir(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing module")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "order") {
		t.Fatalf("expected available modules listed in error, got: %v", err)
	}
}

func TestParseModelFields(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	fields, err := parseModelFields(filepath.Join(dir, "internal", "order", "model.go"))
	if err != nil {
		t.Fatalf("parseModelFields: %v", err)
	}
	if len(fields) == 0 {
		t.Fatal("expected fields, got none")
	}

	// Check TotalCents field.
	var totalCents *moduleField
	for i := range fields {
		if fields[i].goName == "TotalCents" {
			totalCents = &fields[i]
		}
	}
	if totalCents == nil {
		t.Fatal("TotalCents field not found")
	}
	if totalCents.goType != "int" {
		t.Fatalf("expected type 'int', got %q", totalCents.goType)
	}
	if totalCents.gormTag != "column:total_cents" {
		t.Fatalf("unexpected gorm tag: %q", totalCents.gormTag)
	}
	if totalCents.valTag != "min=0" {
		t.Fatalf("unexpected validate tag: %q", totalCents.valTag)
	}
	if totalCents.jsonTag != "total_cents" {
		t.Fatalf("unexpected json tag: %q", totalCents.jsonTag)
	}
}

func TestParseExportedFunctions(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	sigs, err := parseExportedFunctions(filepath.Join(dir, "internal", "order", "service.go"))
	if err != nil {
		t.Fatalf("parseExportedFunctions: %v", err)
	}

	sigStr := strings.Join(sigs, "\n")
	if !strings.Contains(sigStr, "Create") {
		t.Error("expected Create in service sigs")
	}
	if !strings.Contains(sigStr, "FindByID") {
		t.Error("expected FindByID in service sigs")
	}
	if !strings.Contains(sigStr, "UpdateStatus") {
		t.Error("expected UpdateStatus in service sigs")
	}
	// unexported must not appear.
	if strings.Contains(sigStr, "unexported") {
		t.Error("unexported function should not appear")
	}
}

func TestParseExportedFunctionNames(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	names, err := parseExportedFunctionNames(filepath.Join(dir, "internal", "order", "handler.go"))
	if err != nil {
		t.Fatalf("parseExportedFunctionNames: %v", err)
	}

	nameStr := strings.Join(names, ", ")
	for _, want := range []string{"Create", "FindAll", "FindByID", "Update", "Delete"} {
		if !strings.Contains(nameStr, want) {
			t.Errorf("expected handler %q, got: %s", want, nameStr)
		}
	}
	if strings.Contains(nameStr, "helper") {
		t.Error("unexported 'helper' should not appear")
	}
}

func TestCountMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	count := countMigrationFiles(dir, "order")
	if count != 1 {
		t.Fatalf("expected 1 migration file referencing 'order', got %d", count)
	}
}

func TestBuildModuleReport(t *testing.T) {
	dir := t.TempDir()
	buildFakeProject(t, dir)

	modDir := filepath.Join(dir, "internal", "order")
	report, err := buildModuleReport(dir, modDir, "internal/order", "order")
	if err != nil {
		t.Fatalf("buildModuleReport: %v", err)
	}

	checks := []string{
		"Module: Order",
		"FIELDS (from model.go)",
		"total_cents",
		"min=0",
		"SERVICE METHODS (service.go)",
		"Create",
		"FindByID",
		"UpdateStatus",
		"HANDLERS (handler.go)",
		"FindAll",
		"MIGRATIONS",
	}
	for _, want := range checks {
		if !strings.Contains(report, want) {
			t.Errorf("report missing %q\n---report---\n%s", want, report)
		}
	}
	// unexported should be absent from service section.
	if strings.Contains(report, "unexported") {
		t.Error("report should not include unexported function")
	}
}

func TestExtractTagValue(t *testing.T) {
	raw := `gorm:"column:total_cents;not null" json:"total_cents,omitempty" validate:"min=0"`
	if v := extractTagValue(raw, "gorm"); v != "column:total_cents;not null" {
		t.Fatalf("gorm tag: %q", v)
	}
	if v := extractTagValue(raw, "json"); v != "total_cents,omitempty" {
		t.Fatalf("json tag: %q", v)
	}
	if v := extractTagValue(raw, "validate"); v != "min=0" {
		t.Fatalf("validate tag: %q", v)
	}
	if v := extractTagValue(raw, "missing"); v != "" {
		t.Fatalf("missing key should return empty, got %q", v)
	}
}

func TestToSnakeCase(t *testing.T) {
	cases := map[string]string{
		"TotalCents": "total_cents",
		"UserID":     "user_i_d",
		"ID":         "i_d",
		"status":     "status",
	}
	for in, want := range cases {
		if got := toSnakeCase(in); got != want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}
