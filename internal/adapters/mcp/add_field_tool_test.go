package mcp_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// orderModelSrc is a minimal model.go that the add_field tool can patch.
const orderModelSrc = `package order

import (
	"gorm.io/gorm"
)

// Order is the main domain entity.
type Order struct {
	gorm.Model
	TotalCents int    ` + "`" + `gorm:"not null" json:"total_cents" validate:"min=0"` + "`" + `
	Status     string ` + "`" + `gorm:"not null" json:"status" validate:"required,oneof=pending|paid"` + "`" + `
	UserID     uint   ` + "`" + `gorm:"index" json:"user_id"` + "`" + `
}
`

// setupProject creates a minimal kthulu-like project layout under a temp dir.
func setupProject(t *testing.T, module, modelContent string) string {
	t.Helper()
	dir := t.TempDir()
	modelDir := filepath.Join(dir, "internal", module)
	require.NoError(t, os.MkdirAll(modelDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(modelDir, "model.go"), []byte(modelContent), 0o644))
	return dir
}

// addFieldOKExecutor returns success for every command with a fixed stdout.
type addFieldOKExecutor struct {
	migrateStdout string
}

func (e *addFieldOKExecutor) Run(_ context.Context, _ string, args []string) (mcp.CommandResult, error) {
	if len(args) > 0 && args[0] == "migrate" {
		out := e.migrateStdout
		if out == "" {
			out = "ALTER TABLE orders ADD COLUMN test_col TEXT;"
		}
		return mcp.CommandResult{Stdout: out}, nil
	}
	return mcp.CommandResult{}, nil
}

// failingBuildExecutor makes "build" fail so we can test rollback.
type failingBuildExecutor struct{}

func (e *failingBuildExecutor) Run(_ context.Context, _ string, args []string) (mcp.CommandResult, error) {
	if len(args) > 0 && args[0] == "build" {
		return mcp.CommandResult{Stderr: "syntax error"}, &buildError{}
	}
	return mcp.CommandResult{}, nil
}

type buildError struct{}

func (e *buildError) Error() string { return "exit status 1" }

// findAddFieldTool returns the add_field RegisteredTool from the factory.
func findAddFieldTool(t *testing.T, executor mcp.CommandExecutor, workingDir string) mcp.RegisteredTool {
	t.Helper()
	root := &cobra.Command{Use: "kthulu"}
	tagParser := parser.NewTagParser(nil)
	factory := mcp.NewToolFactory(root, executor, tagParser)
	tools := factory.BuildTools(workingDir, nil)
	for _, tool := range tools {
		if tool.Name == "add_field" {
			return tool
		}
	}
	t.Fatal("add_field tool not found in factory")
	return mcp.RegisteredTool{}
}

// callAddField invokes the add_field handler through its typed function.
func callAddField(t *testing.T, tool mcp.RegisteredTool, module, field string) (*mcp_golang.ToolResponse, error) {
	t.Helper()
	handler := tool.Handler.(func(context.Context, mcp.AddFieldArgs) (*mcp_golang.ToolResponse, error))
	return handler(context.Background(), mcp.AddFieldArgs{Module: module, Field: field})
}

// TestAddFieldToolRegistered verifies the tool is registered via init().
func TestAddFieldToolRegistered(t *testing.T) {
	dir := t.TempDir()
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)
	require.Equal(t, "add_field", tool.Name)
	require.NotEmpty(t, tool.Description)
}

// TestAddFieldTimeField verifies a time field is added correctly.
func TestAddFieldTimeField(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	resp, err := callAddField(t, tool, "order", "shipped_at:time")
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Check the model file was patched.
	modelPath := filepath.Join(dir, "internal", "order", "model.go")
	content, err := os.ReadFile(modelPath)
	require.NoError(t, err)
	src := string(content)

	require.Contains(t, src, "ShippedAt", "CamelCase field name must appear")
	require.Contains(t, src, "*time.Time", "time fields must be pointer")
	require.Contains(t, src, `"shipped_at"`, "JSON tag must be present")

	// Response summary must mention the field.
	text := resp.Content[0].TextContent.Text
	require.Contains(t, text, "shipped_at")
	require.Contains(t, text, "Order")
}

// TestAddFieldIntWithRules verifies an int field with validation rules.
func TestAddFieldIntWithRules(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	resp, err := callAddField(t, tool, "order", "discount_cents:int:min=0")
	require.NoError(t, err)

	modelPath := filepath.Join(dir, "internal", "order", "model.go")
	content, _ := os.ReadFile(modelPath)
	src := string(content)

	require.Contains(t, src, "DiscountCents")
	require.Contains(t, src, `"discount_cents"`)
	require.Contains(t, src, "min=0")

	text := resp.Content[0].TextContent.Text
	require.Contains(t, text, "discount_cents")
}

// TestAddFieldBelongsTo verifies that belongs_to produces two fields.
func TestAddFieldBelongsTo(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	resp, err := callAddField(t, tool, "order", "category:belongs_to:category")
	require.NoError(t, err)

	modelPath := filepath.Join(dir, "internal", "order", "model.go")
	content, _ := os.ReadFile(modelPath)
	src := string(content)

	// FK field
	require.Contains(t, src, "CategoryID")
	require.Contains(t, src, `"category_id"`)
	// Relation field
	require.Contains(t, src, "*Category")

	text := resp.Content[0].TextContent.Text
	require.Contains(t, text, "category")
	_ = resp
}

// TestAddFieldBoolField verifies a bool field gets default:false.
func TestAddFieldBoolField(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	_, err := callAddField(t, tool, "order", "is_express:bool")
	require.NoError(t, err)

	content, _ := os.ReadFile(filepath.Join(dir, "internal", "order", "model.go"))
	src := string(content)
	require.Contains(t, src, "IsExpress")
	require.Contains(t, src, "default:false")
}

// TestAddFieldStringRequired verifies string+required adds not null to gorm tag.
func TestAddFieldStringRequired(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	_, err := callAddField(t, tool, "order", "note:string:required")
	require.NoError(t, err)

	content, _ := os.ReadFile(filepath.Join(dir, "internal", "order", "model.go"))
	src := string(content)
	require.Contains(t, src, "Note")
	require.Contains(t, src, "not null")
}

// TestAddFieldInvalidType rejects unknown field types.
func TestAddFieldInvalidType(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	_, err := callAddField(t, tool, "order", "foo:uuid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown field type")
}

// TestAddFieldMissingModule returns an error when the module does not exist.
func TestAddFieldMissingModule(t *testing.T) {
	dir := t.TempDir()
	// No internal/ directory, so findModelFile must fail.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal"), 0o755))
	tool := findAddFieldTool(t, &addFieldOKExecutor{}, dir)

	_, err := callAddField(t, tool, "nonexistent", "foo:string")
	require.Error(t, err)
	require.Contains(t, strings.ToLower(err.Error()), "model.go")
}

// TestAddFieldRestoredOnBuildFailure verifies model.go is rolled back when
// the compiled binary cannot be built after the insertion.
func TestAddFieldRestoredOnBuildFailure(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	original, _ := os.ReadFile(filepath.Join(dir, "internal", "order", "model.go"))

	tool := findAddFieldTool(t, &failingBuildExecutor{}, dir)

	_, err := callAddField(t, tool, "order", "note:string")
	require.Error(t, err)
	require.Contains(t, err.Error(), "compilation failed")

	// The file must be restored.
	restored, _ := os.ReadFile(filepath.Join(dir, "internal", "order", "model.go"))
	require.Equal(t, string(original), string(restored), "model.go must be rolled back on build failure")
}

// TestAddFieldSummaryContent checks the response text includes the migration preview.
func TestAddFieldSummaryContent(t *testing.T) {
	dir := setupProject(t, "order", orderModelSrc)
	migrateSQL := "ALTER TABLE orders ADD COLUMN shipped_at TIMESTAMP;"
	tool := findAddFieldTool(t, &addFieldOKExecutor{migrateStdout: migrateSQL}, dir)

	resp, err := callAddField(t, tool, "order", "shipped_at:time")
	require.NoError(t, err)

	text := resp.Content[0].TextContent.Text
	require.Contains(t, text, "shipped_at")
	require.Contains(t, text, "Migration preview")
	require.Contains(t, text, "migrate_preview")
}
