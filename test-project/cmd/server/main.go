// @kthulu:project:test-project
// @kthulu:generated:true
// @kthulu:features:user,product
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"test-project/internal/core"
	"test-project/internal/adapters/http/modules/product"
	productDomain "test-project/internal/adapters/http/modules/product/domain"
	productHandlers "test-project/internal/adapters/http/modules/product/handlers"
	"test-project/internal/adapters/http/modules/organization"
	organizationDomain "test-project/internal/adapters/http/modules/organization/domain"
	organizationHandlers "test-project/internal/adapters/http/modules/organization/handlers"
	"test-project/internal/adapters/http/modules/auth"
	authDomain "test-project/internal/adapters/http/modules/auth/domain"
	authHandlers "test-project/internal/adapters/http/modules/auth/handlers"
	"test-project/internal/adapters/http/modules/user"
	userDomain "test-project/internal/adapters/http/modules/user/domain"
	userHandlers "test-project/internal/adapters/http/modules/user/handlers"
)

func main() {
	ctx := context.Background()

	app := fx.New(
		// Core providers
		core.CoreRepositoryProviders(),

		// Module providers
		organization.Providers(),
		auth.Providers(),
		user.Providers(),
		product.Providers(),

		// HTTP server
		fx.Invoke(func(lc fx.Lifecycle, userService userDomain.UserService, productService productDomain.ProductService, organizationService organizationDomain.OrganizationService, authService authDomain.AuthService) {
			router := setupRoutes(userService, productService, organizationService, authService)
			server := &http.Server{
				Addr:    ":8080",
				Handler: router,
			}

			lc.Append(fx.Hook{
				OnStart: func(context.Context) error {
					log.Println("Starting server on :8080")
					go server.ListenAndServe()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Stopping server")
					return server.Shutdown(ctx)
				},
			})
		}),
	)

	// Start application
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop application
	if err := app.Stop(ctx); err != nil {
		log.Fatal("Failed to stop application:", err)
	}

	log.Println("Server stopped")
}

func setupRoutes(userService userDomain.UserService, productService productDomain.ProductService, organizationService organizationDomain.OrganizationService, authService authDomain.AuthService) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Add module routes here
	// user routes
	userHandler := userHandlers.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
	// product routes
	productHandler := productHandlers.NewProductHandler(productService)
	productHandler.RegisterRoutes(apiRouter)
	// organization routes
	organizationHandler := organizationHandlers.NewOrganizationHandler(organizationService)
	organizationHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authHandlers.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)

	return router
}
