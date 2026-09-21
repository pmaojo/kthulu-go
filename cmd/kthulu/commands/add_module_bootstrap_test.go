package commands

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	genparser "github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/generator"
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestProject writes a minimal server project to dir, the same way
// `kthulu create` would, without touching the network (no go mod tidy, no go
// build): just the generator's own templates.
func generateTestProject(t *testing.T, dir, projectModule string, features []string) {
	t.Helper()
	gen := generator.NewTemplateGenerator(resolver.NewDependencyResolver(&genparser.ProjectAnalysis{
		Modules:      make(map[string]*genparser.Module),
		Dependencies: []genparser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&generator.GeneratorConfig{
		ProjectName:   filepath.Base(dir),
		ProjectModule: projectModule,
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "none",
		OutputPath:    dir,
		Features:      features,
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
		CustomValues:  map[string]string{"module_path": projectModule},
	})
	require.NoError(t, err)
	require.NoError(t, gen.WriteProject(structure))
}

// `kthulu add module` used to corrupt cmd/server/main.go: it patched an AST
// shape (a local `apiRouter` variable, fx.New(directArgs...)) that the
// generator stopped producing once routing moved into pkg/bootstrap/app.go
// behind fx.New(opts...). Every module added after project creation left the
// project unable to build. This exercises the real command end to end.
func TestRunAddModule_WiresNewModuleWithoutCorruptingMainGo(t *testing.T) {
	dir := t.TempDir()
	projectModule := "github.com/example/luthier"
	generateTestProject(t, dir, projectModule, []string{"auth", "customer"})

	originalWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(originalWd)

	err := runAddModule("wood_stock", []string{"quantity:int", "species:string"}, nil, "", false, true, "", true, false)
	require.NoError(t, err)

	// main.go no longer needs editing at all; it must still parse cleanly.
	mainPath := filepath.Join(dir, "cmd", "server", "main.go")
	mainSrc, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	_, err = parser.ParseFile(token.NewFileSet(), "main.go", mainSrc, 0)
	assert.NoError(t, err, "main.go must still be valid Go after add module")

	bootstrap, err := os.ReadFile(filepath.Join(dir, "pkg", "bootstrap", "app.go"))
	require.NoError(t, err)
	bootstrapSrc := string(bootstrap)
	assert.Contains(t, bootstrapSrc, "wood_stock.Providers()", "the new module must be provided")
	assert.Contains(t, bootstrapSrc, "wood_stockHandler.RegisterRoutes(apiRouter)", "the new module's routes must be registered")
	assert.Contains(t, bootstrapSrc, "customer.Providers()", "the module that was already there must survive")
	_, err = parser.ParseFile(token.NewFileSet(), "app.go", bootstrap, 0)
	assert.NoError(t, err, "pkg/bootstrap/app.go must be valid Go after add module")

	// The default route prefix used to gain an extra leading slash on top of
	// the one the handler template already adds, mounting the route at
	// "//wood_stock" where nothing could ever reach it.
	handler, err := os.ReadFile(filepath.Join(dir, "internal", "modules", "wood_stock", "api", "wood_stock_handler.go"))
	require.NoError(t, err)
	handlerSrc := string(handler)
	assert.Contains(t, handlerSrc, `PathPrefix("/wood_stock")`)
	assert.NotContains(t, handlerSrc, `PathPrefix("//wood_stock")`)

	// --protected was requested and the project has authentication configured.
	assert.Contains(t, handlerSrc, "middleware.AuthMiddleware")
}

// An explicit --prefix must not double up with the leading slash the
// handler template already adds.
func TestRunAddModule_ExplicitPrefixIsNotDoubled(t *testing.T) {
	dir := t.TempDir()
	generateTestProject(t, dir, "github.com/example/luthier", []string{"auth", "customer"})

	originalWd, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(originalWd)

	err := runAddModule("invoice_line", []string{"amount:float"}, nil, "", false, true, "/api/v2/invoice-lines", false, false)
	require.NoError(t, err)

	handler, err := os.ReadFile(filepath.Join(dir, "internal", "modules", "invoice_line", "api", "invoice_line_handler.go"))
	require.NoError(t, err)
	handlerSrc := string(handler)
	assert.Contains(t, handlerSrc, `PathPrefix("/api/v2/invoice-lines")`)
	assert.NotContains(t, handlerSrc, `PathPrefix("//api/v2/invoice-lines")`)
}
