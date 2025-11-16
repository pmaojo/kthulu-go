// @kthulu:project:my-new-project
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

"my-new-project/internal/core"
 "my-new-project/internal/adapters/http/modules/user"
 userDomain "my-new-project/internal/adapters/http/modules/user/domain"
 userHandlers "my-new-project/internal/adapters/http/modules/user/handlers"
 "my-new-project/internal/adapters/http/modules/auth"
 authDomain "my-new-project/internal/adapters/http/modules/auth/domain"
 authHandlers "my-new-project/internal/adapters/http/modules/auth/handlers"
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

// Module providers
		user.Providers(),
		auth.Providers(),

fx.Invoke(func(lc fx.Lifecycle, userService userDomain.UserService, authService authDomain.AuthService) {
router := setupRoutes()
apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// auth routes
	authHandler := authHandlers.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userHandlers.NewUserHandler(userService)
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

func setupRoutes() *mux.Router {
router := mux.NewRouter()

router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
_, _ = w.Write([]byte("OK"))
})

return router
}
