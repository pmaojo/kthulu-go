// @kthulu:project:workshop-api
// @kthulu:generated:true
// @kthulu:features:users
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

	"github.com/example/workshop-api/internal/adapters/http/modules/bookings"
	"github.com/example/workshop-api/internal/adapters/http/modules/cars"
	"github.com/example/workshop-api/internal/adapters/http/modules/customers"
	"github.com/example/workshop-api/internal/adapters/http/modules/services"
	"github.com/example/workshop-api/internal/adapters/http/modules/users"
	usersDomain "github.com/example/workshop-api/internal/adapters/http/modules/users/domain"
	usersHandlers "github.com/example/workshop-api/internal/adapters/http/modules/users/handlers"
	"github.com/example/workshop-api/internal/core"
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

		// Core providers
		fx.Provide(setupRoutes),

		// Module providers
		users.Providers(), customers.Providers(), cars.Providers(), services.Providers(), bookings.Providers(),
		fx.Invoke(func(lc fx.Lifecycle, router *mux.Router, usersService usersDomain.UsersService) {
			apiRouter := router.PathPrefix("/api/v1").Subrouter()

			// users routes
			usersHandler := usersHandlers.NewUsersHandler(usersService)
			usersHandler.RegisterRoutes(apiRouter)

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

func setupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return router
}
