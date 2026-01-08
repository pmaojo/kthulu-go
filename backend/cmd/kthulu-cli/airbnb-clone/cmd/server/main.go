// @kthulu:project:airbnb-clone
// @kthulu:generated:true
// @kthulu:features:user,auth,property,booking,review
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

	"airbnb-clone/internal/core"
 "airbnb-clone/internal/modules/auth"
 authDomain "airbnb-clone/internal/modules/auth/domain"
 authHandlers "airbnb-clone/internal/modules/auth/handlers"
 "airbnb-clone/internal/modules/property"
 propertyDomain "airbnb-clone/internal/modules/property/domain"
 propertyHandlers "airbnb-clone/internal/modules/property/handlers"
 "airbnb-clone/internal/modules/booking"
 bookingDomain "airbnb-clone/internal/modules/booking/domain"
 bookingHandlers "airbnb-clone/internal/modules/booking/handlers"
 "airbnb-clone/internal/modules/review"
 reviewDomain "airbnb-clone/internal/modules/review/domain"
 reviewHandlers "airbnb-clone/internal/modules/review/handlers"
 "airbnb-clone/internal/modules/user"
 userDomain "airbnb-clone/internal/modules/user/domain"
 userHandlers "airbnb-clone/internal/modules/user/handlers"
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
		review.Providers(),
		user.Providers(),
		auth.Providers(),
		property.Providers(),
		booking.Providers(),

		fx.Invoke(func(lc fx.Lifecycle, router *mux.Router, authService authDomain.AuthService, propertyService propertyDomain.PropertyService, bookingService bookingDomain.BookingService, reviewService reviewDomain.ReviewService, userService userDomain.UserService) {
			apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// property routes
	propertyHandler := propertyHandlers.NewPropertyHandler(propertyService)
	propertyHandler.RegisterRoutes(apiRouter)
	// booking routes
	bookingHandler := bookingHandlers.NewBookingHandler(bookingService)
	bookingHandler.RegisterRoutes(apiRouter)
	// review routes
	reviewHandler := reviewHandlers.NewReviewHandler(reviewService)
	reviewHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userHandlers.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authHandlers.NewAuthHandler(authService)
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
