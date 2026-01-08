// @kthulu:core
package modules

import (
	"reflect"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// RouteRegistry manages route registration from multiple modules
type RouteRegistry struct {
	registrars []RouteRegistrar
	logger     *zap.Logger
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry(logger *zap.Logger) *RouteRegistry {
	return &RouteRegistry{
		registrars: make([]RouteRegistrar, 0),
		logger:     logger,
	}
}

// Register adds a route registrar to the registry
func (rr *RouteRegistry) Register(registrar RouteRegistrar) {
	if registrar == nil {
		rr.logger.Warn("nil route registrar provided")
		return
	}
	rr.registrars = append(rr.registrars, registrar)
	rr.logger.Info("Route registrar added", zap.String("type", getTypeName(registrar)))
}

// RegisterAllRoutes registers all routes from all registered handlers
func (rr *RouteRegistry) RegisterAllRoutes(r chi.Router) {
	rr.logger.Info("Registering routes from all modules", zap.Int("count", len(rr.registrars)))

	for _, registrar := range rr.registrars {
		if registrar == nil {
			rr.logger.Warn("nil route registrar skipped")
			continue
		}
		registrar.RegisterRoutes(r)
		rr.logger.Debug("Routes registered", zap.String("handler", getTypeName(registrar)))
	}
}

// getTypeName returns the type name of an interface using reflection
func getTypeName(i interface{}) string {
	if i == nil {
		return ""
	}
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
