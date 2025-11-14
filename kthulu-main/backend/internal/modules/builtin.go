package modules

import (
	"os"

	"go.uber.org/fx"
)

// BuiltinModules lists the modules bundled with the service.
var BuiltinModules = map[string]fx.Option{
	"health":       HealthModule,
	"user":         UserModule,
	"access":       AccessModule,
	"notifier":     NotifierModule,
	"organization": OrganizationModule,
	"contact":      ContactModule,
	"product":      ProductModule,
	"invoice":      InvoiceModule,
	"inventory":    InventoryModule,
	"calendar":     CalendarModule,
	"realtime":     RealtimeModule,
	"static":       StaticModule,
	"verifactu":    VerifactuModule,
	"oauth-sso":    OAuthSSOModule,
	"secure":       SecureModule,
	"flags":        FlagsModule,
	"projects":     ProjectsModule,
	"modules":      ModulesModule,
	"templates":    TemplatesModule,
}

func init() {
	if os.Getenv("LEGACY_AUTH") == "true" {
		BuiltinModules["auth"] = AuthModule
	}
}

// RegisterBuiltinModules adds all builtin modules to the given registry.
func RegisterBuiltinModules(r *Registry) {
	for name, opt := range BuiltinModules {
		r.Register(name, opt)
	}
}
