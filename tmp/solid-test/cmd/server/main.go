// @kthulu:project:solid-test
// @kthulu:generated:true
// @kthulu:features:product,user,organization,auth
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"solid-test/internal/core"
 "solid-test/internal/modules/auth"
 authCore "solid-test/internal/modules/auth/core"
 authAPI "solid-test/internal/modules/auth/api"
 "solid-test/internal/modules/product"
 productCore "solid-test/internal/modules/product/core"
 productAPI "solid-test/internal/modules/product/api"
 "solid-test/internal/modules/user"
 userCore "solid-test/internal/modules/user/core"
 userAPI "solid-test/internal/modules/user/api"
 "solid-test/internal/modules/organization"
 organizationCore "solid-test/internal/modules/organization/core"
 organizationAPI "solid-test/internal/modules/organization/api"
)

type httpServer interface {
	Start() error
	Shutdown(context.Context) error
}

type realHTTPServer struct {
	server *http.Server
}

func newHTTPServer(handler http.Handler) httpServer {
	return &realHTTPServer{
		server: &http.Server{
			Addr:    ":8080",
			Handler: handler,
		},
	}
}

func (s *realHTTPServer) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *realHTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type noopHTTPServer struct{}

func (n *noopHTTPServer) Start() error {
	return nil
}

func (n *noopHTTPServer) Shutdown(context.Context) error {
	return nil
}

var serverBuilder = func(handler http.Handler) httpServer {
	if os.Getenv("KTHULU_TEST_MODE") == "1" {
		return &noopHTTPServer{}
	}
	return newHTTPServer(handler)
}

func main() {
	if err := runApplication(context.Background(), serverBuilder); err != nil {
		log.Fatal("Failed to start application:", err)
	}
}

func runApplication(ctx context.Context, builder func(http.Handler) httpServer) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := fx.New(
		// Core providers
		core.CoreRepositoryProviders(),
		fx.Provide(NewRouter),

		// Module providers
		auth.Providers(),
		product.Providers(),
		user.Providers(),
		organization.Providers(),

		fx.Invoke(func(lc fx.Lifecycle, router *mux.Router, productService productCore.ProductService, userService userCore.UserService, organizationService organizationCore.OrganizationService, authService authCore.AuthService) {

			// Register GTH (HTML) web routes
			RegisterGTHRoutes(router)

			// Register API routes
			apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// organization routes
	organizationHandler := organizationAPI.NewOrganizationHandler(organizationService)
	organizationHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authAPI.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)
	// product routes
	productHandler := productAPI.NewProductHandler(productService)
	productHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userAPI.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)

			server := builder(router)

			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					go func() {
						if err := server.Start(); err != nil {
							log.Println("server error:", err)
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
		}),
	)

	if err := app.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Stop(shutdownCtx)
}

func NewRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return router
}
