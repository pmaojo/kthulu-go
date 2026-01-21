package bootstrap

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"feature-tour/internal/core"
	"feature-tour/internal/infrastructure/database"
 "feature-tour/internal/modules/user"
 userCore "feature-tour/internal/modules/user/core"
 userAPI "feature-tour/internal/modules/user/api"
 "feature-tour/internal/modules/auth"
 authCore "feature-tour/internal/modules/auth/core"
 authAPI "feature-tour/internal/modules/auth/api"
 "feature-tour/internal/modules/organization"
 organizationCore "feature-tour/internal/modules/organization/core"
 organizationAPI "feature-tour/internal/modules/organization/api"
 "feature-tour/internal/modules/contact"
 contactCore "feature-tour/internal/modules/contact/core"
 contactAPI "feature-tour/internal/modules/contact/api"
 "feature-tour/internal/modules/cache"
 cacheCore "feature-tour/internal/modules/cache/core"
 cacheAPI "feature-tour/internal/modules/cache/api"
 "feature-tour/internal/modules/storage"
 storageCore "feature-tour/internal/modules/storage/core"
 storageAPI "feature-tour/internal/modules/storage/api"
 "feature-tour/internal/modules/product"
 productCore "feature-tour/internal/modules/product/core"
 productAPI "feature-tour/internal/modules/product/api"
 "feature-tour/internal/modules/invoice"
 invoiceCore "feature-tour/internal/modules/invoice/core"
 invoiceAPI "feature-tour/internal/modules/invoice/api"
 "feature-tour/internal/modules/mail"
 mailCore "feature-tour/internal/modules/mail/core"
 mailAPI "feature-tour/internal/modules/mail/api"
 "feature-tour/internal/modules/scheduler"
 schedulerCore "feature-tour/internal/modules/scheduler/core"
 schedulerAPI "feature-tour/internal/modules/scheduler/api"
 "feature-tour/internal/modules/events"
 eventsCore "feature-tour/internal/modules/events/core"
 eventsAPI "feature-tour/internal/modules/events/api"

	"feature-tour/internal/adapters/http/gth"

)

// AppOptions returns the common Fx options for the application
func AppOptions() []fx.Option {
	return []fx.Option{
		// Core providers
		core.CoreRepositoryProviders(),
		fx.Provide(NewRouter),

		// Module providers
		product.Providers(),
		cache.Providers(),
		scheduler.Providers(),
		auth.Providers(),
		organization.Providers(),
		invoice.Providers(),
		contact.Providers(),
		mail.Providers(),
		storage.Providers(),
		events.Providers(),
		user.Providers(),

		// Route Registration
		fx.Invoke(RegisterRoutes),
	}
}

// RunMigrations runs database migrations if enabled via env var
func RunMigrations() {
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		db, err := database.NewDB()
		if err == nil {
			database.AutoMigrate(db)
		}
	}
}

// RegisterRoutes registers all application routes
func RegisterRoutes(router *mux.Router, productService productCore.ProductService, contactService contactCore.ContactService, mailService mailCore.MailService, storageService storageCore.StorageService, schedulerService schedulerCore.SchedulerService, userService userCore.UserService, authService authCore.AuthService, organizationService organizationCore.OrganizationService, invoiceService invoiceCore.InvoiceService, cacheService cacheCore.CacheService, eventsService eventsCore.EventsService) {

	// Register GTH (HTML) web routes
	gth.RegisterRoutes(router, userService, productService, invoiceService, mailService, cacheService, storageService, schedulerService, eventsService, )

	// Register API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// product routes
	productHandler := productAPI.NewProductHandler(productService)
	productHandler.RegisterRoutes(apiRouter)
	// invoice routes
	invoiceHandler := invoiceAPI.NewInvoiceHandler(invoiceService)
	invoiceHandler.RegisterRoutes(apiRouter)
	// mail routes
	mailHandler := mailAPI.NewMailHandler(mailService)
	mailHandler.RegisterRoutes(apiRouter)
	// scheduler routes
	schedulerHandler := schedulerAPI.NewSchedulerHandler(schedulerService)
	schedulerHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authAPI.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)
	// organization routes
	organizationHandler := organizationAPI.NewOrganizationHandler(organizationService)
	organizationHandler.RegisterRoutes(apiRouter)
	// contact routes
	contactHandler := contactAPI.NewContactHandler(contactService)
	contactHandler.RegisterRoutes(apiRouter)
	// cache routes
	cacheHandler := cacheAPI.NewCacheHandler(cacheService)
	cacheHandler.RegisterRoutes(apiRouter)
	// storage routes
	storageHandler := storageAPI.NewStorageHandler(storageService)
	storageHandler.RegisterRoutes(apiRouter)
	// events routes
	eventsHandler := eventsAPI.NewEventsHandler(eventsService)
	eventsHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userAPI.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
}

// NewRouter creates a new mux router with health check
func NewRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return router
}
