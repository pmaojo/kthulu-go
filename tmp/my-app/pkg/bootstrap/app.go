package bootstrap

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"my-app/internal/core"
	"my-app/internal/infrastructure/database"
 "my-app/internal/modules/user"
 userCore "my-app/internal/modules/user/core"
 userAPI "my-app/internal/modules/user/api"
 "my-app/internal/modules/organization"
 organizationCore "my-app/internal/modules/organization/core"
 organizationAPI "my-app/internal/modules/organization/api"
 "my-app/internal/modules/auth"
 authCore "my-app/internal/modules/auth/core"
 authAPI "my-app/internal/modules/auth/api"
 "my-app/internal/modules/product"
 productCore "my-app/internal/modules/product/core"
 productAPI "my-app/internal/modules/product/api"

	"my-app/internal/adapters/http/gth"

)

// AppOptions returns the common Fx options for the application
func AppOptions() []fx.Option {
	return []fx.Option{
		// Core providers
		core.CoreRepositoryProviders(),
		fx.Provide(NewRouter),

		// Module providers
		product.Providers(),
		user.Providers(),
		organization.Providers(),
		auth.Providers(),

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
func RegisterRoutes(router *mux.Router, organizationService organizationCore.OrganizationService, authService authCore.AuthService, productService productCore.ProductService, userService userCore.UserService) {

	// Register GTH (HTML) web routes
	gth.RegisterRoutes(router, productService, userService, )

	// Register API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// product routes
	productHandler := productAPI.NewProductHandler(productService)
	productHandler.RegisterRoutes(apiRouter)
	// user routes
	userHandler := userAPI.NewUserHandler(userService)
	userHandler.RegisterRoutes(apiRouter)
	// organization routes
	organizationHandler := organizationAPI.NewOrganizationHandler(organizationService)
	organizationHandler.RegisterRoutes(apiRouter)
	// auth routes
	authHandler := authAPI.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiRouter)
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
