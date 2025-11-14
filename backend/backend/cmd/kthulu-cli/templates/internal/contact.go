// @kthulu:module:contacts
// Este módulo representa la funcionalidad de gestión de contacts.
// Se construye mediante Fx y se registra dinámicamente en el router central.
// Estructura:
// - Repositorio de contacts (infrastructure/db)
// - Caso de uso (usecase)
// - Handler HTTP (adapters/http)
// - Registro de rutas dinámico (via RouteRegistry)
package modules

import (
	"go.uber.org/fx"

	adapterhttp "backend/internal/adapters/http"
	"backend/internal/usecase"
)

// ContactModule provides contact functionality
var ContactModule = fx.Options(
	// Use cases
	fx.Provide(
		usecase.NewContactUseCase,
	),

	// HTTP handlers
	fx.Provide(
		adapterhttp.NewContactHandler,
	),

	// Register routes
	fx.Invoke(func(handler *adapterhttp.ContactHandler, registry *RouteRegistry) {
		registry.Register(handler)
	}),
)
