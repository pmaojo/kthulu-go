// @kthulu:gth:routes
package gth

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/a-h/templ"

	"vercel-test/internal/views/pages"
	"vercel-test/internal/views/components"
	"vercel-test/internal/views/partials"

	// Module Core Imports
	
	
	productcore "vercel-test/internal/modules/product/core"
	
	
	
	
	
	
	
	
	usercore "vercel-test/internal/modules/user/core"
	
	
)

// RegisterRoutes registers the GTH (HTML) web routes
func RegisterRoutes(router *mux.Router, 
	productService productcore.ProductService,
userService usercore.UserService,
) {
	// Public Landing Page
	router.HandleFunc("/", handleLandingPage).Methods("GET")

	router.HandleFunc("/admin", handleDashboard).Methods("GET")
	router.HandleFunc("/admin/", handleDashboard).Methods("GET")

	// Module Routes
	
	
	router.HandleFunc("/admin/product", func(w http.ResponseWriter, r *http.Request) {
		handleProductPage(w, r, productService)
	}).Methods("GET")
	router.HandleFunc("/admin/product/new", handleProductNew).Methods("GET")
	router.HandleFunc("/admin/product/search", func(w http.ResponseWriter, r *http.Request) {
		handleProductSearch(w, r, productService)
	}).Methods("GET")
	router.HandleFunc("/admin/product/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleProductEdit(w, r, productService)
	}).Methods("GET")
	router.HandleFunc("/admin/product/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleProductDelete(w, r, productService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/user", func(w http.ResponseWriter, r *http.Request) {
		handleUserPage(w, r, userService)
	}).Methods("GET")
	router.HandleFunc("/admin/user/new", handleUserNew).Methods("GET")
	router.HandleFunc("/admin/user/search", func(w http.ResponseWriter, r *http.Request) {
		handleUserSearch(w, r, userService)
	}).Methods("GET")
	router.HandleFunc("/admin/user/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleUserEdit(w, r, userService)
	}).Methods("GET")
	router.HandleFunc("/admin/user/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleUserDelete(w, r, userService)
	}).Methods("GET")
	
	
}

func handleLandingPage(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, r, pages.LandingPage())
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats := []pages.DashboardStats{
		{Label: "System Status", Value: "Online", Icon: "🟢", Color: "#dcfce7"},
	}
	renderTempl(w, r, pages.DashboardPage(stats))
}



func handleProductPage(w http.ResponseWriter, r *http.Request, service productcore.ProductService) {
	// Fetch real items
	filter := productcore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListProducts(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]productcore.Product, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.ProductTableRows(items))
		return
	}

	renderTempl(w, r, pages.ProductsPage(items))
}

func handleProductNew(w http.ResponseWriter, r *http.Request) {
	item := productcore.Product{}
	renderTempl(w, r, components.ProductFormModal(&item, false))
}

func handleProductEdit(w http.ResponseWriter, r *http.Request, service productcore.ProductService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetProductByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ProductFormModal(item, true))
}

func handleProductDelete(w http.ResponseWriter, r *http.Request, service productcore.ProductService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetProductByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Product", uint(id), "/api/v1/product/" + vars["id"]))
}

func handleProductSearch(w http.ResponseWriter, r *http.Request, service productcore.ProductService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := productcore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListProducts(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]productcore.Product, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.ProductTableRows(items))
}



func handleUserPage(w http.ResponseWriter, r *http.Request, service usercore.UserService) {
	// Fetch real items
	filter := usercore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListUsers(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]usercore.User, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.UserTableRows(items))
		return
	}

	renderTempl(w, r, pages.UsersPage(items))
}

func handleUserNew(w http.ResponseWriter, r *http.Request) {
	item := usercore.User{}
	renderTempl(w, r, components.UserFormModal(&item, false))
}

func handleUserEdit(w http.ResponseWriter, r *http.Request, service usercore.UserService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetUserByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.UserFormModal(item, true))
}

func handleUserDelete(w http.ResponseWriter, r *http.Request, service usercore.UserService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetUserByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("User", uint(id), "/api/v1/user/" + vars["id"]))
}

func handleUserSearch(w http.ResponseWriter, r *http.Request, service usercore.UserService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := usercore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListUsers(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]usercore.User, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.UserTableRows(items))
}



func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
