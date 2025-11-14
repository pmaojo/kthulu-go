//go:build wireinject

package main

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/wire"
	"go.uber.org/fx"

	"backend/core"
	"backend/core/metrics"
	"backend/internal/modules"
	flagcfg "backend/internal/modules/flags"
	"backend/internal/observability"
)

type routerProviderType = func(p struct {
	fx.In
	RouteRegistry *modules.RouteRegistry
	DB            *sql.DB
	Logger        observability.Logger
	Config        *core.Config
	TokenManager  core.TokenManager
	Flags         flagcfg.HeaderConfig
	Metrics       *metrics.PrometheusMetrics
}) chi.Router

type httpServerProviderType = func(r chi.Router, cfg *core.Config, logger observability.Logger) *http.Server

type validateStartupFuncType = func(db *sql.DB, cfg *core.Config, logger observability.Logger) error

type registerHooksFuncType = func(lc fx.Lifecycle, srv *http.Server, db *sql.DB, logger observability.Logger)

var (
	routerProviderValue      routerProviderType      = newRouter
	httpServerProviderValue  httpServerProviderType  = newHTTPServer
	validateStartupFuncValue validateStartupFuncType = validateStartup
	registerHooksFuncValue   registerHooksFuncType   = registerHooks
)

type appSet struct {
	RouterProvider      routerProviderType
	HTTPServerProvider  httpServerProviderType
	ValidateStartupFunc validateStartupFuncType
	RegisterHooksFunc   registerHooksFuncType
}

// InitializeApp sets up application dependencies using Wire.
func InitializeApp() appSet {
	wire.Build(
		wire.Value(routerProviderValue),
		wire.Value(httpServerProviderValue),
		wire.Value(validateStartupFuncValue),
		wire.Value(registerHooksFuncValue),
		wire.Struct(new(appSet), "*"),
	)
	return appSet{}
}
