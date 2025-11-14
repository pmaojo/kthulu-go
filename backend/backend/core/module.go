// @kthulu:core
package core

import "go.uber.org/fx"

// Module wires core services for Fx dependency injection.
// It provides all essential infrastructure components.
var Module = fx.Options(
	fx.Provide(
		NewDB,            // Provides *sql.DB
		NewGormDB,        // Provides *gorm.DB (wraps sql.DB)
		NewZapLogger,     // Provides *zap.Logger - this is what most handlers need
		NewLogger,        // Provides core.Logger interface (wraps zap)
		NewSugaredLogger, // Provides *zap.SugaredLogger for convenience
		NewJWT,
		NewFeatureFlagClient, // Provides feature flag client
	),
	// Note: Migrations should be run separately via cmd/migrate/main.go
)
