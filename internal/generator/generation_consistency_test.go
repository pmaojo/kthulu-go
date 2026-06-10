package generator

import (
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findFile returns the content of the generated file at path, failing the test
// if it does not exist.
func findFile(t *testing.T, structure *ProjectStructure, path string) string {
	t.Helper()
	for _, f := range structure.Files {
		if f.Path == path {
			return f.Content
		}
	}
	t.Fatalf("expected generated file %q, got files: %v", path, filePaths(structure))
	return ""
}

func filePaths(structure *ProjectStructure) []string {
	paths := make([]string, 0, len(structure.Files))
	for _, f := range structure.Files {
		paths = append(paths, f.Path)
	}
	return paths
}

// TestGenerateProject_PluralModuleNames guards against regressions where a
// plural feature name (e.g. "payments") produced mismatched entity names
// between the backend module (Payment) and the rest of the generated code.
func TestGenerateProject_PluralModuleNames(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "shop",
		ProjectModule: "github.com/example/shop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "user", "payments"},
		CustomValues:  map[string]string{"module_path": "github.com/example/shop"},
	})
	require.NoError(t, err)

	core := findFile(t, structure, "internal/modules/payments/core/payments.go")
	assert.Contains(t, core, "type Payment struct", "entity must be singularized")
	assert.Contains(t, core, "type PaymentService interface")
	assert.NotContains(t, core, ",string", "JSON tags must not force string-encoded numbers")

	providers := findFile(t, structure, "internal/core/providers.go")
	assert.Contains(t, providers, "&paymentsCore.Payment{}", "AutoMigrate must reference the singular entity")
	assert.NotContains(t, providers, "&paymentsCore.Payments{}")

	bootstrap := findFile(t, structure, "pkg/bootstrap/app.go")
	assert.Contains(t, bootstrap, "paymentsCore.PaymentService", "invoke params must use the singular service type")
	assert.Contains(t, bootstrap, "paymentsAPI.NewPaymentHandler", "route wiring must use the singular handler constructor")

	gthRoutes := findFile(t, structure, "internal/adapters/http/gth/routes.go")
	assert.Contains(t, gthRoutes, "paymentscore.PaymentService")
	assert.Contains(t, gthRoutes, "partials.PaymentTableRows")
	assert.Contains(t, gthRoutes, "pages.PaymentsPage")
	assert.NotContains(t, gthRoutes, "PaymentsService")

	partial := findFile(t, structure, "internal/views/partials/payments_table_rows.templ")
	assert.Contains(t, partial, "PaymentTableRows(items []core.Payment)")

	handler := findFile(t, structure, "internal/modules/payments/api/payments_handler.go")
	assert.Contains(t, handler, "decodePayment", "handler must use the tolerant JSON decoder")
	assert.Contains(t, handler, `sub.HandleFunc("", h.List)`, "collection routes must work without a trailing slash")
}

// TestGenerateProject_RelationsAndEnterprisePaths guards the belongs_to field
// DSL and the enterprise module layout (internal/adapters/http/modules).
func TestGenerateProject_RelationsAndEnterprisePaths(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "crm",
		ProjectModule: "github.com/example/crm",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Enterprise:    true,
		Features:      []string{"auth", "user", "contact", "invoice"},
		ModuleFields: map[string][]string{
			"contact": {"name:string", "email:string", "vip:bool"},
			"invoice": {"amount:float", "status:string", "issued_at:time", "customer:belongs_to:contact"},
		},
		CustomValues: map[string]string{"module_path": "github.com/example/crm"},
	})
	require.NoError(t, err)

	const modulesPath = "internal/adapters/http/modules"

	invoiceCore := findFile(t, structure, modulesPath+"/invoice/core/invoice.go")
	assert.Contains(t, invoiceCore, "CustomerID uint", "belongs_to must generate a foreign key")
	assert.Contains(t, invoiceCore, "*contactDomain.Contact", "belongs_to must generate a relation field")
	assert.Contains(t, invoiceCore, "github.com/example/crm/"+modulesPath+"/contact/core",
		"relation import must point at the related module core package")
	assert.NotContains(t, invoiceCore, "/contact/domain", "the related entity lives in core, not domain")
	assert.Contains(t, invoiceCore, `gorm:"column:amount"`)

	gthRoutes := findFile(t, structure, "internal/adapters/http/gth/routes.go")
	assert.Contains(t, gthRoutes, "github.com/example/crm/"+modulesPath+"/invoice/core",
		"GTH routes must respect the enterprise module layout")
	assert.NotContains(t, gthRoutes, "github.com/example/crm/internal/modules/")

	invoicePage := findFile(t, structure, "internal/views/pages/invoice_page.templ")
	assert.Contains(t, invoicePage, "github.com/example/crm/"+modulesPath+"/invoice/core")

	// Relation pointer fields must not be rendered in tables or forms.
	partial := findFile(t, structure, "internal/views/partials/invoice_table_rows.templ")
	assert.NotContains(t, partial, "item.Customer ", "relation fields must be skipped in table rows")
	assert.Contains(t, partial, `item.IssuedAt.Format`, "time fields must be formatted")

	form := findFile(t, structure, "internal/views/components/invoice_form.templ")
	assert.NotContains(t, strings.ReplaceAll(form, "item.CustomerID", ""), "item.Customer",
		"relation struct fields must be skipped in forms")

	handler := findFile(t, structure, modulesPath+"/invoice/api/invoice_handler.go")
	assert.Contains(t, handler, "coerceNumber", "numeric form values must be coerced")
	assert.Contains(t, handler, "coerceTime", "time form values must be coerced")
}

// TestGenerateProject_ValidationRules guards the validation rule DSL
// (name:type:rules) end to end: generated Validate() methods, service-level
// enforcement and the 422 handler mapping.
func TestGenerateProject_ValidationRules(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "shop",
		ProjectModule: "github.com/example/shop",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "user", "product"},
		ModuleFields: map[string][]string{
			"product": {
				"name:string:required,min=2",
				"price:int:required,min=1",
				"status:string:oneof=draft|active",
				"contact_email:string:email",
			},
		},
		CustomValues: map[string]string{"module_path": "github.com/example/shop"},
	})
	require.NoError(t, err)

	validation := findFile(t, structure, "internal/modules/product/core/product_validation.go")
	assert.Contains(t, validation, "type ValidationErrors map[string]string")
	assert.Contains(t, validation, "func (e *Product) Validate() error")
	assert.Contains(t, validation, `errs["name"] = "is required"`)
	assert.Contains(t, validation, "utf8.RuneCountInString(e.Name) < 2")
	assert.Contains(t, validation, "e.Price < 1")
	assert.Contains(t, validation, `must be one of: draft, active`)
	assert.Contains(t, validation, "emailPattern.MatchString(e.ContactEmail)")

	service := findFile(t, structure, "internal/modules/product/core/product_service.go")
	assert.Contains(t, service, "entity.Validate()", "services must enforce validation on create/update")

	handler := findFile(t, structure, "internal/modules/product/api/product_handler.go")
	assert.Contains(t, handler, "StatusUnprocessableEntity", "validation failures must map to 422")
	assert.Contains(t, handler, "core.ValidationErrors")

	// Modules without rules still get a compiling Validate() so the service
	// call never breaks.
	userValidation := findFile(t, structure, "internal/modules/user/core/user_validation.go")
	assert.Contains(t, userValidation, "func (e *User) Validate() error")
}

// TestGenerateProject_QueueRuntime guards the background job runtime: the
// queues feature must generate infrastructure (not a CRUD module) and wire it
// into the fx bootstrap.
func TestGenerateProject_QueueRuntime(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "worker",
		ProjectModule: "github.com/example/worker",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "user", "queues"},
		CustomValues:  map[string]string{"module_path": "github.com/example/worker"},
	})
	require.NoError(t, err)

	queueFile := findFile(t, structure, "internal/infrastructure/queue/queue.go")
	assert.Contains(t, queueFile, "func (q *Queue) Enqueue(")
	assert.Contains(t, queueFile, "func (q *Queue) Every(")
	assert.Contains(t, queueFile, "MaxAttempts")

	findFile(t, structure, "internal/infrastructure/queue/queue_test.go")

	jobsFile := findFile(t, structure, "internal/jobs/jobs.go")
	assert.Contains(t, jobsFile, "func Register(q *queue.Queue) error")
	assert.Contains(t, jobsFile, "github.com/example/worker/internal/infrastructure/queue")

	bootstrap := findFile(t, structure, "pkg/bootstrap/app.go")
	assert.Contains(t, bootstrap, "fx.Provide(newQueue)")
	assert.Contains(t, bootstrap, "fx.Invoke(startQueue)")
	assert.Contains(t, bootstrap, "jobs.Register(q)")

	// "queues" is infrastructure, not a CRUD module.
	for _, f := range structure.Files {
		assert.NotContains(t, f.Path, "modules/queues/", "queues must not be generated as a CRUD module")
	}
}

// TestGenerateProject_MailAndStorage guards the mail and storage drivers:
// generated as infrastructure with tests and wired into the fx bootstrap.
func TestGenerateProject_MailAndStorage(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "infra",
		ProjectModule: "github.com/example/infra",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "user", "mail", "storage"},
		CustomValues:  map[string]string{"module_path": "github.com/example/infra"},
	})
	require.NoError(t, err)

	mailFile := findFile(t, structure, "internal/infrastructure/mail/mailer.go")
	assert.Contains(t, mailFile, "func NewMailer() Mailer")
	findFile(t, structure, "internal/infrastructure/mail/mailer_test.go")

	storageFile := findFile(t, structure, "internal/infrastructure/storage/storage.go")
	assert.Contains(t, storageFile, "func NewStorage() Storage")
	findFile(t, structure, "internal/infrastructure/storage/storage_test.go")

	bootstrap := findFile(t, structure, "pkg/bootstrap/app.go")
	assert.Contains(t, bootstrap, "fx.Provide(mail.NewMailer)")
	assert.Contains(t, bootstrap, "fx.Provide(storage.NewStorage)")

	for _, f := range structure.Files {
		assert.NotContains(t, f.Path, "modules/mail/", "mail must not be generated as a CRUD module")
		assert.NotContains(t, f.Path, "modules/storage/", "storage must not be generated as a CRUD module")
	}
}

// TestGenerateProject_InfraCatalog guards the catalog-driven infrastructure
// runtimes (cache, events, i18n, policy, rate, seeder, session, validate).
func TestGenerateProject_InfraCatalog(t *testing.T) {
	gen := NewTemplateGenerator(resolver.NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	}))

	structure, err := gen.GenerateProject(&GeneratorConfig{
		ProjectName:   "kitchen",
		ProjectModule: "github.com/example/kitchen",
		TemplateType:  "server",
		Database:      "sqlite",
		Auth:          "jwt",
		Frontend:      "templ",
		Features:      []string{"auth", "user", "cache", "events", "i18n", "policy", "rate", "seeder", "session", "validate"},
		CustomValues:  map[string]string{"module_path": "github.com/example/kitchen"},
	})
	require.NoError(t, err)

	for _, name := range []string{"cache", "events", "i18n", "policy", "rate", "seeder", "session", "validate"} {
		findFile(t, structure, "internal/infrastructure/"+name+"/"+name+".go")
		for _, f := range structure.Files {
			assert.NotContains(t, f.Path, "modules/"+name+"/", name+" must not be generated as a CRUD module")
		}
	}

	bootstrap := findFile(t, structure, "pkg/bootstrap/app.go")
	for _, provider := range []string{
		"fx.Provide(cache.NewCache)",
		"fx.Provide(events.NewEventBus)",
		"fx.Provide(i18n.NewTranslator)",
		"fx.Provide(policy.NewGate)",
		"fx.Provide(rate.NewRateLimiter)",
		"fx.Provide(seeder.NewSeeder)",
		"fx.Provide(session.NewSessionStore)",
		"fx.Provide(session.NewManager)",
		"fx.Provide(validate.NewValidator)",
	} {
		assert.Contains(t, bootstrap, provider)
	}
}
