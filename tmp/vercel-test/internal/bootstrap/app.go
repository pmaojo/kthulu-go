package bootstrap

import (
	"net/http"
	
	"github.com/gorilla/mux"
	"go.uber.org/fx"

	"vercel-test/internal/core"
 "vercel-test/internal/modules/organization"
 organizationCore "vercel-test/internal/modules/organization/core"
 organizationAPI "vercel-test/internal/modules/organization/api"
 "vercel-test/internal/modules/auth"
 authCore "vercel-test/internal/modules/auth/core"
 authAPI "vercel-test/internal/modules/auth/api"
 "vercel-test/internal/modules/product"
 productCore "vercel-test/internal/modules/product/core"
 productAPI "vercel-test/internal/modules/product/api"
 "vercel-test/internal/modules/user"
 userCore "vercel-test/internal/modules/user/core"
 userAPI "vercel-test/internal/modules/user/api"

	"vercel-test/internal/views"
	"vercel-test/internal/adapters/http/gth"

)

// AppOptions returns the common Fx options for the application
func AppOptions() []fx.Option {
	return []fx.Option{
		// Core providers
		core.CoreRepositoryProviders(),
		fx.Provide(NewRouter),

		// Module providers
		organization.Providers(),
		auth.Providers(),
		product.Providers(),
		user.Providers(),

		// Route Registration
		fx.Invoke(RegisterRoutes),
	}
}

// RegisterRoutes registers all application routes
func RegisterRoutes(router *mux.Router, productService productCore.ProductService, userService userCore.UserService, organizationService organizationCore.OrganizationService, authService authCore.AuthService) {

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
