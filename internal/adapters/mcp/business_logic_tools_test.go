package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const minimalServiceSrc = `package order

import (
	"context"
	"fmt"
)

// OrderService is the service for the order module.
type OrderService struct {
	repo OrderRepository
}

// FindAll returns all orders.
func (s *OrderService) FindAll(ctx context.Context) ([]Order, error) {
	return nil, fmt.Errorf("not implemented")
}
`

const minimalModelSrc = `package order

import "gorm.io/gorm"

// Order is the main domain entity.
type Order struct {
	gorm.Model
	TotalCents int ` + "`" + `gorm:"not null" json:"total_cents"` + "`" + `
}

// OrderRepository defines the data access interface.
type OrderRepository interface {
	FindAll() ([]Order, error)
}
`

func createServiceFixture(t *testing.T, workingDir, module, content string) string {
	t.Helper()
	dir := filepath.Join(workingDir, "internal", module)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "service.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func createModelFixture(t *testing.T, workingDir, module, content string) string {
	t.Helper()
	dir := filepath.Join(workingDir, "internal", module)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "model.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func callGenerateMethodHandler(t *testing.T, tool RegisteredTool, args GenerateMethodArgs) (*mcp_golang.ToolResponse, error) {
	t.Helper()
	handler := tool.Handler.(func(context.Context, GenerateMethodArgs) (*mcp_golang.ToolResponse, error))
	return handler(context.Background(), args)
}

func callAddHookHandler(t *testing.T, tool RegisteredTool, args AddHookArgs) (*mcp_golang.ToolResponse, error) {
	t.Helper()
	handler := tool.Handler.(func(context.Context, AddHookArgs) (*mcp_golang.ToolResponse, error))
	return handler(context.Background(), args)
}

// ---------------------------------------------------------------------------
// Unit helpers
// ---------------------------------------------------------------------------

func TestBuildMethodStub(t *testing.T) {
	code := buildMethodStub("OrderService", "CalculateTax", "Calculate 21% VAT on total_cents")
	if !strings.Contains(code, "func (o *OrderService) CalculateTax") {
		t.Errorf("stub missing method signature, got:\n%s", code)
	}
	if !strings.Contains(code, "TODO: implement CalculateTax") {
		t.Errorf("stub missing TODO comment, got:\n%s", code)
	}
	if !strings.Contains(code, "Calculate 21% VAT on total_cents") {
		t.Errorf("stub missing description, got:\n%s", code)
	}
	if !strings.Contains(code, "not implemented") {
		t.Errorf("stub missing 'not implemented', got:\n%s", code)
	}
}

func TestParseServiceStructName(t *testing.T) {
	dir := t.TempDir()
	path := createServiceFixture(t, dir, "order", minimalServiceSrc)
	name, err := parseServiceStructName(path)
	if err != nil {
		t.Fatal(err)
	}
	if name != "OrderService" {
		t.Errorf("expected OrderService, got %q", name)
	}
}

func TestMethodExists(t *testing.T) {
	dir := t.TempDir()
	path := createServiceFixture(t, dir, "order", minimalServiceSrc)
	if !methodExists(path, "FindAll") {
		t.Error("expected FindAll to exist")
	}
	if methodExists(path, "CalculateTax") {
		t.Error("expected CalculateTax to not exist")
	}
}

func TestStripMarkdownFences(t *testing.T) {
	input := "```go\nfunc Foo() {}\n```"
	got := stripMarkdownFences(input)
	if strings.Contains(got, "```") {
		t.Errorf("markdown fences not stripped: %q", got)
	}
	if !strings.Contains(got, "func Foo()") {
		t.Errorf("code missing after strip: %q", got)
	}
}

func TestTitleCase(t *testing.T) {
	cases := map[string]string{
		"order":   "Order",
		"product": "Product",
		"":        "",
		"Order":   "Order",
	}
	for input, want := range cases {
		got := titleCase(input)
		if got != want {
			t.Errorf("titleCase(%q) = %q, want %q", input, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// generate_method tool tests
// ---------------------------------------------------------------------------

func TestGenerateMethodTool_Registered(t *testing.T) {
	var found bool
	for _, builder := range pluginBuilders {
		for _, tool := range builder(nil, "") {
			if tool.Name == "generate_method" {
				found = true
			}
		}
	}
	if !found {
		t.Error("generate_method not registered in plugin list")
	}
}

func TestGenerateMethodTool_ValidationErrors(t *testing.T) {
	tool := generateMethodTool(nil, "")
	cases := []struct {
		name    string
		args    GenerateMethodArgs
		wantErr string
	}{
		{"no module", GenerateMethodArgs{MethodName: "Foo", Description: "bar"}, "module is required"},
		{"no method_name", GenerateMethodArgs{Module: "order", Description: "bar"}, "method_name is required"},
		{"no description", GenerateMethodArgs{Module: "order", MethodName: "Foo"}, "description is required"},
	}
	for _, tc := range cases {
		_, err := callGenerateMethodHandler(t, tool, tc.args)
		if err == nil {
			t.Fatalf("%s: expected error, got nil", tc.name)
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("%s: expected error containing %q, got %q", tc.name, tc.wantErr, err.Error())
		}
	}
}

func TestGenerateMethodTool_MissingServiceFile(t *testing.T) {
	dir := t.TempDir()
	tool := generateMethodTool(nil, dir)
	_, err := callGenerateMethodHandler(t, tool, GenerateMethodArgs{
		Module:      "nonexistent",
		MethodName:  "DoSomething",
		Description: "does something",
	})
	if err == nil {
		t.Fatal("expected error for missing service.go")
	}
}

func TestGenerateMethodTool_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	createServiceFixture(t, dir, "order", minimalServiceSrc)
	_ = os.Unsetenv("OPENAI_API_KEY")

	tool := generateMethodTool(nil, dir)
	_, err := callGenerateMethodHandler(t, tool, GenerateMethodArgs{
		Module:      "order",
		MethodName:  "FindAll",
		Description: "return all orders",
		DryRun:      true,
	})
	if err == nil {
		t.Fatal("expected error for duplicate method")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %s", err.Error())
	}
}

func TestGenerateMethodTool_DryRunNoAPIKey(t *testing.T) {
	dir := t.TempDir()
	createServiceFixture(t, dir, "order", minimalServiceSrc)
	_ = os.Unsetenv("OPENAI_API_KEY")

	tool := generateMethodTool(nil, dir)
	resp, err := callGenerateMethodHandler(t, tool, GenerateMethodArgs{
		Module:      "order",
		MethodName:  "CalculateTax",
		Description: "Calculate 21% VAT on total_cents",
		DryRun:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil || len(resp.Content) == 0 {
		t.Fatal("expected a non-empty response")
	}
	text := resp.Content[0].TextContent.Text
	if !strings.Contains(text, "DRY RUN") {
		t.Errorf("expected DRY RUN in response, got: %s", text)
	}
	if !strings.Contains(text, "CalculateTax") {
		t.Errorf("expected CalculateTax in response, got: %s", text)
	}
}

// ---------------------------------------------------------------------------
// add_hook tool tests
// ---------------------------------------------------------------------------

func TestAddHookTool_Registered(t *testing.T) {
	var found bool
	for _, builder := range pluginBuilders {
		for _, tool := range builder(nil, "") {
			if tool.Name == "add_hook" {
				found = true
			}
		}
	}
	if !found {
		t.Error("add_hook not registered in plugin list")
	}
}

func TestAddHookTool_ValidationErrors(t *testing.T) {
	tool := addHookTool(nil, "")
	cases := []struct {
		name    string
		args    AddHookArgs
		wantErr string
	}{
		{"no module", AddHookArgs{Lifecycle: "before_create", Description: "x"}, "module is required"},
		{"no lifecycle", AddHookArgs{Module: "order", Description: "x"}, "lifecycle is required"},
		{"no description", AddHookArgs{Module: "order", Lifecycle: "before_create"}, "description is required"},
	}
	for _, tc := range cases {
		_, err := callAddHookHandler(t, tool, tc.args)
		if err == nil {
			t.Fatalf("%s: expected error, got nil", tc.name)
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("%s: expected %q, got %q", tc.name, tc.wantErr, err.Error())
		}
	}
}

func TestAddHookTool_InvalidLifecycle(t *testing.T) {
	dir := t.TempDir()
	createModelFixture(t, dir, "order", minimalModelSrc)
	tool := addHookTool(nil, dir)
	_, err := callAddHookHandler(t, tool, AddHookArgs{
		Module:      "order",
		Lifecycle:   "invalid_event",
		Description: "do something",
	})
	if err == nil {
		t.Fatal("expected error for invalid lifecycle")
	}
	if !strings.Contains(err.Error(), "unknown lifecycle") {
		t.Errorf("expected 'unknown lifecycle', got: %s", err.Error())
	}
}

func TestAddHookTool_MissingModelFile(t *testing.T) {
	dir := t.TempDir()
	tool := addHookTool(nil, dir)
	_, err := callAddHookHandler(t, tool, AddHookArgs{
		Module:      "nonexistent",
		Lifecycle:   "before_create",
		Description: "send welcome email",
	})
	if err == nil {
		t.Fatal("expected error for missing model.go")
	}
}

func TestBuildHookCode(t *testing.T) {
	code := buildHookCode("Order", "BeforeCreate", "send welcome email via mail service")
	if !strings.Contains(code, "func (o *Order) BeforeCreate(tx *gorm.DB) error") {
		t.Errorf("hook missing signature, got:\n%s", code)
	}
	if !strings.Contains(code, "send welcome email via mail service") {
		t.Errorf("hook missing description, got:\n%s", code)
	}
	if !strings.Contains(code, "TODO: send welcome email via mail service") {
		t.Errorf("hook missing TODO, got:\n%s", code)
	}
	if !strings.Contains(code, "return nil") {
		t.Errorf("hook missing return nil, got:\n%s", code)
	}
}

func TestLifecycleToMethod(t *testing.T) {
	expected := map[string]string{
		"before_create": "BeforeCreate",
		"after_create":  "AfterCreate",
		"before_update": "BeforeUpdate",
		"after_update":  "AfterUpdate",
		"before_delete": "BeforeDelete",
		"after_delete":  "AfterDelete",
	}
	for lifecycle, method := range expected {
		got, ok := lifecycleToMethod[lifecycle]
		if !ok {
			t.Errorf("lifecycle %q not in map", lifecycle)
			continue
		}
		if got != method {
			t.Errorf("lifecycle %q: expected %q got %q", lifecycle, method, got)
		}
	}
	if len(lifecycleToMethod) != len(expected) {
		t.Errorf("expected %d lifecycle entries, got %d", len(expected), len(lifecycleToMethod))
	}
}

func TestParseModelStructName(t *testing.T) {
	dir := t.TempDir()
	path := createModelFixture(t, dir, "order", minimalModelSrc)
	name, err := parseModelStructName(path)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Order" {
		t.Errorf("expected Order, got %q", name)
	}
}

func TestHookEventDescription(t *testing.T) {
	cases := map[string]string{
		"BeforeCreate": "before inserting",
		"AfterCreate":  "after inserting",
		"BeforeUpdate": "before updating",
		"AfterUpdate":  "after updating",
		"BeforeDelete": "before deleting",
		"AfterDelete":  "after deleting",
	}
	for method, want := range cases {
		got := hookEventDescription(method)
		if got != want {
			t.Errorf("hookEventDescription(%q): expected %q, got %q", method, want, got)
		}
	}
}

func TestPluginRegistration_BothTools(t *testing.T) {
	var foundGenerateMethod, foundAddHook bool
	for _, builder := range pluginBuilders {
		tools := builder(nil, "")
		for _, tool := range tools {
			switch tool.Name {
			case "generate_method":
				foundGenerateMethod = true
			case "add_hook":
				foundAddHook = true
			}
		}
	}
	if !foundGenerateMethod {
		t.Error("generate_method not registered in plugin list")
	}
	if !foundAddHook {
		t.Error("add_hook not registered in plugin list")
	}
}
