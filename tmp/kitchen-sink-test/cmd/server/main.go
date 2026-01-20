// @kthulu:project:kitchen-sink-test
// @kthulu:generated:true
// @kthulu:features:auth,calendar,contact,inventory,invoice,organization,product,user,verifactu
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

	"kitchen-sink-test/internal/core"
 "kitchen-sink-test/internal/modules/auth"
 authCore "kitchen-sink-test/internal/modules/auth/core"
 authAPI "kitchen-sink-test/internal/modules/auth/api"
 "kitchen-sink-test/internal/modules/user"
 userCore "kitchen-sink-test/internal/modules/user/core"
 userAPI "kitchen-sink-test/internal/modules/user/api"
 "kitchen-sink-test/internal/modules/organization"
 organizationCore "kitchen-sink-test/internal/modules/organization/core"
 organizationAPI "kitchen-sink-test/internal/modules/organization/api"
 "kitchen-sink-test/internal/modules/contact"
 contactCore "kitchen-sink-test/internal/modules/contact/core"
 contactAPI "kitchen-sink-test/internal/modules/contact/api"
 "kitchen-sink-test/internal/modules/inventory"
 inventoryCore "kitchen-sink-test/internal/modules/inventory/core"
 inventoryAPI "kitchen-sink-test/internal/modules/inventory/api"
 "kitchen-sink-test/internal/modules/product"
 productCore "kitchen-sink-test/internal/modules/product/core"
 productAPI "kitchen-sink-test/internal/modules/product/api"
 "kitchen-sink-test/internal/modules/invoice"
 invoiceCore "kitchen-sink-test/internal/modules/invoice/core"
 invoiceAPI "kitchen-sink-test/internal/modules/invoice/api"
 "kitchen-sink-test/internal/modules/verifactu"
 verifactuCore "kitchen-sink-test/internal/modules/verifactu/core"
 verifactuAPI "kitchen-sink-test/internal/modules/verifactu/api"
 "kitchen-sink-test/internal/modules/calendar"
 calendarCore "kitchen-sink-test/internal/modules/calendar/core"
 calendarAPI "kitchen-sink-test/internal/modules/calendar/api"
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
		user.Providers(),
		contact.Providers(),
		inventory.Providers(),
		verifactu.Providers(),
		calendar.Providers(),
		organization.Providers(),
		product.Providers(),
		invoice.Providers(),

		fx.Invoke(func(lc fx.Lifecycle, router *mux.Router, authService authCore.AuthService, userService userCore.UserService, calendarService calendarCore.CalendarService, contactService contactCore.ContactService, invoiceService invoiceCore.InvoiceService, organizationService organizationCore.OrganizationService, inventoryService inventoryCore.InventoryService, productService productCore.ProductService, verifactuService verifactuCore.VerifactuService) {

			// Register GTH (HTML) web routes
			RegisterGTHRoutes(router, calendarService, contactService, inventoryService, invoiceService, organizationService, productService, userService, verifactuService, )

			// Register API routes
			apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// organization routes
	organizationHandler := organizationAPI.NewOrganizationHandler(organizationService)
	organizationHandler.RegisterRoutes(apiRouter)
	// inventory routes
	inventoryHandler := inventoryAPI.NewInventoryHandler(inventoryService)
	inventoryHandler.RegisterRoutes(apiRouter)
	// invoice routes
	invoiceHandler := invoiceAPI.NewInvoiceHandler(invoiceService)
	invoiceHandler.RegisterRoutes(apiRouter)
	// verifactu routes
	verifactuHandler := verifactuAPI.NewVerifactuHandler(verifactuService)
	verifactuHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authAPI.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userAPI.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
	// contact routes
	contactHandler := contactAPI.NewContactHandler(contactService)
	contactHandler.RegisterRoutes(apiRouter)
	// product routes
	productHandler := productAPI.NewProductHandler(productService)
	productHandler.RegisterRoutes(apiRouter)
	// calendar routes
	calendarHandler := calendarAPI.NewCalendarHandler(calendarService)
	calendarHandler.RegisterRoutes(apiRouter)

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
