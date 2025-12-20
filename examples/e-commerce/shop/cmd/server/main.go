// @kthulu:project:shop
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

	"shop/internal/core"
 "shop/internal/adapters/http/modules/auth"
 authDomain "shop/internal/adapters/http/modules/auth/domain"
 authHandlers "shop/internal/adapters/http/modules/auth/handlers"
 "shop/internal/adapters/http/modules/user"
 userDomain "shop/internal/adapters/http/modules/user/domain"
 userHandlers "shop/internal/adapters/http/modules/user/handlers"

 "shop/internal/adapters/http/modules/products"
 productsDomain "shop/internal/adapters/http/modules/products/domain"
 productsHandlers "shop/internal/adapters/http/modules/products/handlers"
 "shop/internal/adapters/http/modules/orders"
 ordersDomain "shop/internal/adapters/http/modules/orders/domain"
 ordersHandlers "shop/internal/adapters/http/modules/orders/handlers"
 "shop/internal/adapters/http/modules/payments"
 paymentsDomain "shop/internal/adapters/http/modules/payments/domain"
 paymentsHandlers "shop/internal/adapters/http/modules/payments/handlers"
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
		products.Providers(),
		orders.Providers(),
		payments.Providers(),

fx.Invoke(func(
	lc fx.Lifecycle,
	userService userDomain.UserService,
	authService authDomain.AuthService,
	productsService productsDomain.ProductsService,
	ordersService ordersDomain.OrdersService,
	paymentsService paymentsDomain.PaymentsService,
) {
router := setupRoutes()
apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// user routes
	userHandler := userHandlers.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authHandlers.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)

	// products routes
	productsHandler := productsHandlers.NewProductsHandler(productsService)
	// Manually register routes since generated handlers don't have RegisterRoutes yet
	apiRouter.HandleFunc("/products", productsHandler.Create).Methods("POST")
	apiRouter.HandleFunc("/products", productsHandler.List).Methods("GET")
	apiRouter.HandleFunc("/products/{id}", productsHandler.GetByID).Methods("GET")

	// orders routes
	ordersHandler := ordersHandlers.NewOrdersHandler(ordersService)
	apiRouter.HandleFunc("/orders", ordersHandler.Create).Methods("POST")
	apiRouter.HandleFunc("/orders", ordersHandler.List).Methods("GET")
	apiRouter.HandleFunc("/orders/{id}", ordersHandler.GetByID).Methods("GET")

	// payments routes
	paymentsHandler := paymentsHandlers.NewPaymentsHandler(paymentsService)
	apiRouter.HandleFunc("/payments", paymentsHandler.Create).Methods("POST")
	apiRouter.HandleFunc("/payments", paymentsHandler.List).Methods("GET")
	apiRouter.HandleFunc("/payments/{id}", paymentsHandler.GetByID).Methods("GET")

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
