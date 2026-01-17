// @kthulu:project:tournament-api
// @kthulu:generated:true
// @kthulu:features:user,auth
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

	"tournament-api/internal/core"
	"tournament-api/internal/modules/auth"
	authAPI "tournament-api/internal/modules/auth/api"
	authCore "tournament-api/internal/modules/auth/core"
	"tournament-api/internal/modules/participant"
	participantAPI "tournament-api/internal/modules/participant/api"
	"tournament-api/internal/modules/tournament"
	tournamentAPI "tournament-api/internal/modules/tournament/api"
	"tournament-api/internal/modules/user"
	userAPI "tournament-api/internal/modules/user/api"
	userCore "tournament-api/internal/modules/user/core"
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
		user.Providers(),
		auth.Providers(), tournament.Providers(), participant.Providers(), fx.Invoke(func(lc fx.Lifecycle, router *mux.Router, userService userCore.UserService, authService authCore.AuthService, tournamentHandler *tournamentAPI.TournamentHandler, participantHandler *participantAPI.ParticipantHandler) {
			apiRouter := router.PathPrefix("/api/v1").Subrouter()

			// user routes
			userHandler := userAPI.NewUserHandler(userService)
			userHandler.RegisterRoutes(apiRouter)
			// auth routes
			authHandler := authAPI.NewAuthHandler(authService)
			authHandler.RegisterRoutes(apiRouter)

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
			tournamentHandler.RegisterRoutes(apiRouter)
			participantHandler.RegisterRoutes(apiRouter)
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
