package mcp_test

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

const astFixture = `package sample

// Greeter says hello.
type Greeter struct {
	Name string
}

// Greet returns a greeting.
func (g *Greeter) Greet(prefix string) (string, error) {
	return prefix + g.Name, nil
}

// MaxRetries bounds retry loops.
const MaxRetries = 3

var defaultGreeter = Greeter{Name: "kthulu"}

// Run is the entry point.
func Run() error {
	return nil
}
`

func TestGoASTServiceOutline(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "sample.go", astFixture)

	service := mcp.NewGoASTService()
	outline, err := service.Outline(dir, mcp.GoOutlineArgs{Path: "sample.go"})
	require.NoError(t, err)

	require.Contains(t, outline, "package sample")
	require.Contains(t, outline, "[struct] type Greeter {1 fields}")
	require.Contains(t, outline, "[method] func (*Greeter) Greet(prefix string) (string, error)")
	require.Contains(t, outline, "[const] const MaxRetries")
	require.Contains(t, outline, "[var] var defaultGreeter")
	require.Contains(t, outline, "[func] func Run() error")
}

func TestGoASTServiceOutlineDirectorySkipsTests(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "sample.go", astFixture)
	writeTestFile(t, dir, "sample_test.go", "package sample\n\nfunc TestRun() {}\n")

	service := mcp.NewGoASTService()

	outline, err := service.Outline(dir, mcp.GoOutlineArgs{Path: "."})
	require.NoError(t, err)
	require.Contains(t, outline, "Run")
	require.NotContains(t, outline, "TestRun")

	outline, err = service.Outline(dir, mcp.GoOutlineArgs{Path: ".", IncludeTests: true})
	require.NoError(t, err)
	require.Contains(t, outline, "TestRun")
}

func TestGoASTServiceFindSymbol(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "sample.go", astFixture)

	service := mcp.NewGoASTService()

	result, err := service.FindSymbol(dir, mcp.GoFindSymbolArgs{Name: "greet"})
	require.NoError(t, err)
	require.Contains(t, result, "Greeter")
	require.Contains(t, result, "Greet")

	result, err = service.FindSymbol(dir, mcp.GoFindSymbolArgs{Name: "Greet", Exact: true})
	require.NoError(t, err)
	require.Contains(t, result, "1 symbol(s)")
	require.Contains(t, result, "sample.go:9")
}

func TestGoASTServiceSymbolSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "sample.go", astFixture)

	service := mcp.NewGoASTService()

	// Whole-project lookup.
	source, err := service.SymbolSource(dir, mcp.GoSymbolSourceArgs{Name: "Greet"})
	require.NoError(t, err)
	require.Contains(t, source, "// Greet returns a greeting.")
	require.Contains(t, source, "return prefix + g.Name, nil")

	// Missing symbols are reported.
	_, err = service.SymbolSource(dir, mcp.GoSymbolSourceArgs{Name: "DoesNotExist"})
	require.ErrorContains(t, err, "not found")
}
