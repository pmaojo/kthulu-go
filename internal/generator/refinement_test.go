package generator

import (
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGenerator() *TemplateGenerator {
	return NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))
}

// auth: jwt used to add golang-jwt to go.mod and a line to the README and
// nothing else: no login, no token issuance, no middleware, every CRUD route
// open. This pins that the auth module actually generates a working flow and
// that CRUD routes are mounted behind it.
func TestGenerateProject_JWTAuthIsReal(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "workshop",
		ProjectModule: "github.com/example/workshop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "none",
		Features:      []string{"auth", "customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	middleware := findFile(t, structure, "internal/infrastructure/middleware/auth.go")
	assert.Contains(t, middleware, "func AuthMiddleware(")
	assert.Contains(t, middleware, "JWT_SECRET", "must fail closed without a configured signing key")

	authCore := findFile(t, structure, "internal/modules/auth/core/auth.go")
	assert.Contains(t, authCore, "PasswordHash string")
	assert.Contains(t, authCore, `json:"-"`, "the password hash must never be serialised back to a client")

	authService := findFile(t, structure, "internal/modules/auth/core/auth_service.go")
	assert.Contains(t, authService, "bcrypt.GenerateFromPassword")
	assert.Contains(t, authService, "bcrypt.CompareHashAndPassword")
	assert.Contains(t, authService, "jwt.NewWithClaims")

	authHandler := findFile(t, structure, "internal/modules/auth/api/auth_handler.go")
	assert.Contains(t, authHandler, `"/register"`)
	assert.Contains(t, authHandler, `"/login"`)

	// A CRUD module's routes are mounted behind the middleware when the
	// project has authentication configured.
	customerHandler := findFile(t, structure, "internal/modules/customer/api/customer_handler.go")
	assert.Contains(t, customerHandler, "middleware.AuthMiddleware")

	// The auth module mounts its own routes and cannot require the token it
	// is the one issuing.
	assert.NotContains(t, authHandler, `sub.Use(middleware.AuthMiddleware)`)
}

// Fields are silently absent otherwise: an "auth: none" project must not get
// authentication scaffolding, and CRUD routes must not be gated behind a
// middleware nothing configured.
func TestGenerateProject_NoAuthConfiguredLeavesRoutesOpen(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "openshop",
		ProjectModule: "github.com/example/openshop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	customerHandler := findFile(t, structure, "internal/modules/customer/api/customer_handler.go")
	assert.NotContains(t, customerHandler, "middleware.AuthMiddleware")
}

// user and organization used to be named as having curated defaults and then
// explicitly skipped before any field was assigned, so both were generated
// with nothing but an id and timestamps, and the fieldless-module warning
// exempted them too, so nobody was told.
func TestGenerateProject_UserAndOrganizationGetRealFields(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "org-app",
		ProjectModule: "github.com/example/orgapp",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"user", "organization"},
	})
	require.NoError(t, err)

	user := findFile(t, structure, "internal/modules/user/core/user.go")
	assert.Contains(t, user, "Email string")
	assert.Contains(t, user, "PasswordHash string")
	assert.Contains(t, user, `json:"-"`, "the password hash must not be serialised")

	org := findFile(t, structure, "internal/modules/organization/core/organization.go")
	assert.Contains(t, org, "Slug string")
}

// A table named "Repair_orders" (the Go type, title-cased) is created quoted
// by GORM; any hand-written SQL spelling it repair_orders, unquoted, matches
// nothing on Postgres, which folds unquoted identifiers to lower case.
func TestGenerateProject_TableNamesAreSnakeCase(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "luthier",
		ProjectModule: "github.com/example/luthier",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"repair_order"},
		ModuleFields:  map[string][]string{"repair_order": {"description:string:required"}},
	})
	require.NoError(t, err)

	core := findFile(t, structure, "internal/modules/repair_order/core/repair_order.go")
	assert.Contains(t, core, `return "repair_orders"`)
	assert.NotContains(t, core, `return "Repair_orders"`,
		"the table name string must be snake_case; Go identifiers like ListRepair_orders still use PascalCase and are fine")

	found := false
	for _, f := range structure.Files {
		if strings.HasPrefix(f.Path, "migrations/") {
			found = true
			assert.Contains(t, f.Content, "CREATE TABLE IF NOT EXISTS repair_orders")
		}
	}
	assert.True(t, found, "expected a migration file for repair_order")
}

// A belongs_to foreign key used to reference the related module's Go type
// name (e.g. "Customers") rather than the table it is actually created as
// ("customers"), so the FOREIGN KEY clause pointed at a table that did not
// exist under Postgres's folding rules.
func TestGenerateProject_ForeignKeyReferencesSnakeCaseTable(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "luthier",
		ProjectModule: "github.com/example/luthier",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"customer", "instrument"},
		ModuleFields: map[string][]string{
			"customer":   {"full_name:string:required"},
			"instrument": {"kind:string:required", "owner:belongs_to:customer"},
		},
	})
	require.NoError(t, err)

	found := false
	for _, f := range structure.Files {
		if strings.HasPrefix(f.Path, "migrations/") && strings.Contains(f.Content, "FOREIGN KEY") {
			found = true
			assert.Contains(t, f.Content, "REFERENCES customers(id)")
		}
	}
	assert.True(t, found, "expected a migration with a belongs_to foreign key")
}

// A caller-supplied field named "password" or "secret" must never round-trip
// through the JSON API, curated defaults or not.
func TestGenerateProject_SecretFieldsAreNeverSerialised(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "vault-app",
		ProjectModule: "github.com/example/vaultapp",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"credential"},
		ModuleFields:  map[string][]string{"credential": {"label:string", "secret:string", "api_key:string"}},
	})
	require.NoError(t, err)

	core := findFile(t, structure, "internal/modules/credential/core/credential.go")
	assert.Contains(t, core, `Secret string `+"`"+`json:"-"`)
	assert.Contains(t, core, `ApiKey string `+"`"+`json:"-"`)
	assert.Contains(t, core, `Label string `+"`"+`json:"label"`)
}

// declaring frontend: "none" must actually skip the GTH scaffold. It used to
// be overridden back to the template's default whenever the value was
// literally "none", so the only way to say "no frontend" was ignored.
func TestGenerateProject_ExplicitNoFrontendIsHonoured(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "headless",
		ProjectModule: "github.com/example/headless",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	for _, f := range structure.Files {
		assert.NotContains(t, f.Path, "internal/views/", "frontend: none must not generate GTH views")
	}
}

// RegenerateBootstrap is what `kthulu add module` now uses to wire a new
// module into pkg/bootstrap/app.go. It must produce a file that imports,
// provides and routes every module passed in, not just the ones from the
// generator's original config.
func TestRegenerateBootstrap_IncludesEveryModulePassed(t *testing.T) {
	gen := newTestGenerator()

	_, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "shop",
		ProjectModule: "github.com/example/shop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "none",
		Frontend:      "none",
		Features:      []string{"customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	content := gen.RegenerateBootstrap([]string{"customer", "wood_stock"})
	assert.Contains(t, content, "wood_stock.Providers()")
	assert.Contains(t, content, "wood_stockHandler.RegisterRoutes(apiRouter)")
	assert.Contains(t, content, "customer.Providers()", "must still carry the module that was already there")
}

// "user" is generally pulled in only as auth's dependency, not written
// literally in a plan's features list. The per-module GTH view step used to
// walk config.Features raw, so a module present only through dependency
// resolution — like user through auth — was wired into RegisterRoutes and
// pkg/bootstrap/app.go's call to it, with no views ever generated for it:
// the generated project failed to compile from the moment it was created,
// whenever frontend was "templ" and any feature list included auth.
func TestGenerateProject_TransitiveModuleGetsGTHViewsToo(t *testing.T) {
	gen := newTestGenerator()

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "gth-shop",
		ProjectModule: "github.com/example/gthshop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	// "user" was pulled in as auth's dependency and must have gotten the
	// same views as any other module.
	page := findFile(t, structure, "internal/views/pages/user_page.templ")
	assert.Contains(t, page, "UsersPage")

	routes := findFile(t, structure, "internal/adapters/http/gth/routes.go")
	assert.Contains(t, routes, "handleUserPage")
}

// RegenerateGTHRoutes is what `kthulu add module` now uses to rebuild
// internal/adapters/http/gth/routes.go, the GTH counterpart to
// RegenerateBootstrap. It must wire every module passed in.
func TestRegenerateGTHRoutes_IncludesEveryModulePassed(t *testing.T) {
	gen := newTestGenerator()

	_, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "gth-shop",
		ProjectModule: "github.com/example/gthshop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	content, err := gen.RegenerateGTHRoutes([]string{"auth", "customer", "user", "wood_stock"})
	require.NoError(t, err)
	assert.Contains(t, content, "handleWood_stockPage", "the new module must get a page handler")
	assert.Contains(t, content, "handleCustomerPage", "the module that was already there must survive")
	assert.NotContains(t, content, "handleAuthPage", "auth mounts its own routes and gets no admin CRUD page")
}

// pkg/bootstrap/app.go's call to gth.RegisterRoutes and that function's own
// signature are positional arguments on either side of one call: even when
// both files agree on which modules exist, they must also agree on what
// order the parameters come in. RegenerateBootstrap and RegenerateGTHRoutes
// are called separately by `add module`, from a module list built by
// walking a map (so its order is not guaranteed), which is exactly the
// scenario that produced a swapped customerService/userService pair that
// type-checked as "does not implement" rather than "not enough arguments".
func TestRegenerateBootstrapAndGTHRoutes_AgreeOnParameterOrderRegardlessOfInputOrder(t *testing.T) {
	gen := newTestGenerator()

	_, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "gth-shop",
		ProjectModule: "github.com/example/gthshop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "customer"},
		ModuleFields:  map[string][]string{"customer": {"full_name:string:required"}},
	})
	require.NoError(t, err)

	// Deliberately different orderings of the same module set, as two
	// independent map iterations might produce.
	bootstrap := gen.RegenerateBootstrap([]string{"wood_stock", "user", "customer", "auth"})
	routes, err := gen.RegenerateGTHRoutes([]string{"customer", "auth", "wood_stock", "user"})
	require.NoError(t, err)

	callLine := extractCall(bootstrap, "gth.RegisterRoutes(router,")
	sigLine := extractCall(routes, "func RegisterRoutes(router *mux.Router,")
	require.NotEmpty(t, callLine, "expected a gth.RegisterRoutes call in bootstrap")
	require.NotEmpty(t, sigLine, "expected a RegisterRoutes signature in routes.go")

	callOrder := serviceOrder(callLine)
	sigOrder := serviceOrder(sigLine)
	assert.Equal(t, sigOrder, callOrder,
		"the call's argument order must match the signature's parameter order:\ncall: %s\nsig:  %s", callLine, sigLine)
}

// extractCall returns the parenthesised argument list that starts at marker
// (marker must include the opening paren), tracking paren depth so it stops
// at the call or signature's own closing paren — whether that's on the same
// line (a call statement) or several lines down (a multi-line signature).
func extractCall(content, marker string) string {
	idx := strings.Index(content, marker)
	if idx == -1 {
		return ""
	}
	depth := 0
	for i := idx; i < len(content); i++ {
		switch content[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return content[idx : i+1]
			}
		}
	}
	return ""
}

// serviceOrder returns the module name prefixes of every "<name>Service"
// argument or parameter, in the order they appear. It looks only at each
// comma-separated entry's first word (the argument name, or the parameter
// name in a signature) so a typed signature entry like
// "customerService customercore.CustomerService" counts once, not twice.
func serviceOrder(call string) []string {
	var order []string
	for _, part := range strings.Split(call, ",") {
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		if idx := strings.Index(name, "Service"); idx > 0 {
			order = append(order, name[:idx])
		}
	}
	return order
}
