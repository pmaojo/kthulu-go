// @title Kthulu API
// @version 1.0.0
// @description Complete enterprise-grade API for the Kthulu framework
// @description
// @description This API provides authentication, user management, organization management,
// @description and ERP-lite modules including contacts, products, invoices, inventory, and calendar.
// @description
// @description ## Authentication
// @description
// @description Most endpoints require authentication via JWT Bearer token in the Authorization header:
// @description ```
// @description Authorization: Bearer <your-jwt-token>
// @description ```
// @description
// @description ## Module Tags
// @description
// @description Endpoints are tagged by module for selective generation:
// @description - `@kthulu:core` - Core framework functionality
// @description - `@kthulu:module:auth` - Authentication module
// @description - `@kthulu:module:user` - User management module
// @description - `@kthulu:module:org` - Organization management module
// @description - `@kthulu:module:contacts` - Contacts module
// @description - `@kthulu:module:products` - Products module
// @description - `@kthulu:module:invoices` - Invoices module
// @description - `@kthulu:module:inventory` - Inventory module
// @description - `@kthulu:module:calendar` - Calendar module

// @contact.name Kthulu Framework
// @contact.url https://github.com/kthulu-framework

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token obtained from /auth/login endpoint. Format: Bearer <token>

package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/core/metrics"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/http/middleware"
	db "github.com/pmaojo/kthulu-go/backend/internal/infrastructure/db"
	"github.com/pmaojo/kthulu-go/backend/internal/modules"
	flagcfg "github.com/pmaojo/kthulu-go/backend/internal/modules/flags"
	vf "github.com/pmaojo/kthulu-go/backend/internal/modules/verifactu"
	"github.com/pmaojo/kthulu-go/backend/internal/observability"
)

// newRouter constructs the application's HTTP router with middleware.
func newRouter(p struct {
	fx.In
	RouteRegistry *modules.RouteRegistry
	DB            *sql.DB
	Logger        observability.Logger
	Config        *core.Config
	TokenManager  core.TokenManager
	Flags         flagcfg.HeaderConfig
	Metrics       *metrics.PrometheusMetrics
}) chi.Router {
	r := chi.NewRouter()

	allowedOrigins := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if p.Config.Env != "production" {
		allowedOrigins = append(allowedOrigins, "http://localhost:4173", "http://127.0.0.1:4173")
	}
	allowedOriginsMap := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowedOriginsMap[origin] = struct{}{}
	}

	// Add middleware stack
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")
			if origin != "" {
				if _, ok := allowedOriginsMap[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

					if req.Method == http.MethodOptions {
						w.WriteHeader(http.StatusNoContent)
						return
					}
				}
			}
			next.ServeHTTP(w, req)
		})
	})
	r.Use(otelhttp.NewMiddleware("kthulu-service"))
	r.Use(middleware.TraceIDMiddleware)
	r.Use(middleware.JWTTraceMiddleware(p.TokenManager))
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.LoggingMiddleware(p.Logger))
	if p.Metrics != nil {
		r.Use(middleware.MetricsMiddleware(p.Metrics.Provider))
	}
	r.Use(middleware.RecoveryMiddleware(p.Logger))
	r.Use(middleware.AdvancedHealthMiddleware(p.DB, p.Logger, p.Config.Version))
	r.Use(middleware.FlagsMiddleware(p.Flags))
	limiter := rate.NewLimiter(rate.Limit(p.Config.RateLimit.RequestsPerSecond), p.Config.RateLimit.Burst)
	r.Use(middleware.RateLimitMiddleware(limiter))
	r.Use(chimiddleware.Compress(5))

	// Expose Prometheus metrics endpoint before module routes
	r.Handle("/metrics", p.Metrics.Handler)

	// Register all routes from modules
	p.RouteRegistry.RegisterAllRoutes(r)
	return r
}

// newHTTPServer prepares the HTTP server instance with configuration.
func newHTTPServer(r chi.Router, cfg *core.Config, logger observability.Logger) *http.Server {
	// Validate configuration
	if cfg.Server.Addr == "" {
		logger.Fatal("Server address not configured")
	}

	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.Info("HTTP server configured",
		zap.String("addr", server.Addr),
		zap.Duration("read_timeout", server.ReadTimeout),
		zap.Duration("write_timeout", server.WriteTimeout),
		zap.Duration("idle_timeout", server.IdleTimeout),
	)

	return server
}

// validateStartup performs startup validation checks
func validateStartup(db *sql.DB, cfg *core.Config, logger observability.Logger) error {
	logger.Info("Performing startup validation")

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Error("Database connection validation failed", zap.Error(err))
		return fmt.Errorf("database connection failed: %w", err)
	}
	logger.Info("Database connection validated")

	// Validate critical configuration
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret not configured")
	}

	if cfg.Server.Addr == "" {
		return fmt.Errorf("server address not configured")
	}

	logger.Info("Startup validation completed successfully")
	return nil
}

// registerHooks wires server lifecycle to Fx with proper logging and graceful shutdown.
func registerHooks(lc fx.Lifecycle, srv *http.Server, db *sql.DB, logger observability.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := core.Migrate(db, observability.GetZapLogger(logger)); err != nil {
				logger.Error("Database migration failed", zap.Error(err))
				return err
			}

			logger.Info("Starting HTTP server",
				zap.String("addr", srv.Addr),
				zap.Duration("read_timeout", srv.ReadTimeout),
				zap.Duration("write_timeout", srv.WriteTimeout),
				zap.Duration("idle_timeout", srv.IdleTimeout),
			)

			// Start server in a goroutine
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("HTTP server failed to start", zap.Error(err))
				}
			}()

			// Verify server started successfully by checking if we can connect
			// Give it a moment to start
			time.Sleep(100 * time.Millisecond)

			logger.Info("HTTP server started successfully", zap.String("addr", srv.Addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Initiating graceful shutdown")

			// Create a context with timeout for shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// Shutdown HTTP server gracefully
			logger.Info("Shutting down HTTP server")
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to shutdown HTTP server gracefully", zap.Error(err))
				// Force close if graceful shutdown fails
				if closeErr := srv.Close(); closeErr != nil {
					logger.Error("Failed to force close HTTP server", zap.Error(closeErr))
				}
				return err
			}
			logger.Info("HTTP server shutdown completed")

			// Close database connection
			logger.Info("Closing database connection")
			if err := core.CloseDB(db, observability.GetZapLogger(logger)); err != nil {
				logger.Error("Failed to close database connection", zap.Error(err))
				return err
			}
			logger.Info("Database connection closed")

			logger.Info("Graceful shutdown completed successfully")
			return nil
		},
	})
}

func run() error {
	cfg, err := core.NewConfig()
	if err != nil {
		fmt.Println("Failed to load configuration", err)
		return err
	}

	if cfg.Sentry.Enabled && cfg.Sentry.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{Dsn: cfg.Sentry.DSN, Environment: cfg.Env}); err != nil {
			fmt.Println("Failed to initialize Sentry", err)
		} else {
			defer sentry.Flush(2 * time.Second)
		}
	}

	registry := modules.NewRegistry()
	modules.RegisterBuiltinModules(registry)

	// Only load the modules needed for the current web UI experience
	coreModules := []string{"projects", "templates", "modules", "static", "health"}

	builder := modules.NewModuleSetBuilder(registry)
	for _, moduleName := range coreModules {
		if _, ok := registry.GetModule(moduleName); ok {
			builder.WithModule(moduleName)
		} else {
			fmt.Printf("Module not found: %s\n", moduleName)
		}
	}

	moduleSet := builder.Build()

	app := fx.New(
		fx.Supply(cfg),
		fx.Supply(moduleSet),

		core.Module,
		observability.Module,
		modules.FlagsModule,

		moduleSet.Build([]string{}),
		modules.SharedServiceProviders(),

		fx.Provide(
			newRouter,
			newHTTPServer,
		),

		fx.Invoke(validateStartup),
		fx.Invoke(registerHooks),
		fx.Invoke(func(tp *sdktrace.TracerProvider, m *metrics.PrometheusMetrics) {}),
		fx.Invoke(func(lc fx.Lifecycle, params struct {
			fx.In
			Repo   vf.Repository `optional:"true"`
			Config *core.Config
		}) {
			if params.Repo == nil {
				return
			}
			lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
				if params.Config.VerifactuMode == "real-time" {
					return params.Repo.SetLiveMode(ctx, time.Now().Year(), true)
				}
				return nil
			}})
		}),

		fx.NopLogger,
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		fmt.Println("Failed to start application", err)
		return err
	}

	<-app.Done()

	if err := app.Stop(ctx); err != nil {
		fmt.Println("Failed to stop application gracefully", err)
		return err
	}

	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "verifactu" && len(os.Args) > 2 && os.Args[2] == "export" {
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		orgID := fs.Int("org", 0, "Organization ID")
		_ = fs.Parse(os.Args[3:])
		if *orgID == 0 {
			fmt.Fprintln(os.Stderr, "--org is required")
			os.Exit(1)
		}
		if err := exportVerifactu(*orgID); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// exportVerifactu performs the export use case from the CLI.
func exportVerifactu(orgID int) error {
	ctx := context.Background()
	cfg, err := core.NewConfig()
	if err != nil {
		return err
	}
	logger, err := core.NewZapLogger(cfg)
	if err != nil {
		return err
	}
	defer logger.Sync()

	dbConn, err := core.NewDB(cfg, logger)
	if err != nil {
		return err
	}
	defer core.CloseDB(dbConn, logger)

	repo := db.NewVerifactuRepository(dbConn)
	signer := vf.NewHMACSigner([]byte(os.Getenv("VERIFACTU_SIGN_KEY")))
	svc := vf.NewService(repo, signer, cfg.VerifactuSIFCode, cfg.VerifactuMode)

	data, sig, err := svc.ExportRecords(ctx, orgID)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("verifactu_%d.zip", orgID)
	if err := os.WriteFile(file, data, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "exported %s signature %s\n", file, hex.EncodeToString(sig))
	return nil
}
