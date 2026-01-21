// @kthulu:gth:routes
package gth

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/a-h/templ"

	"feature-tour/internal/views/pages"
	"feature-tour/internal/views/components"
	"feature-tour/internal/views/partials"

	// Module Core Imports
	
	
	
	
	
	
	productcore "feature-tour/internal/modules/product/core"
	
	
	
	invoicecore "feature-tour/internal/modules/invoice/core"
	
	
	
	mailcore "feature-tour/internal/modules/mail/core"
	
	
	
	cachecore "feature-tour/internal/modules/cache/core"
	
	
	
	storagecore "feature-tour/internal/modules/storage/core"
	
	
	
	schedulercore "feature-tour/internal/modules/scheduler/core"
	
	
	
	eventscore "feature-tour/internal/modules/events/core"
	
	
	
	
	usercore "feature-tour/internal/modules/user/core"
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
)

// RegisterRoutes registers the GTH (HTML) web routes
func RegisterRoutes(router *mux.Router, 
	userService usercore.UserService,
productService productcore.ProductService,
invoiceService invoicecore.InvoiceService,
mailService mailcore.MailService,
cacheService cachecore.CacheService,
storageService storagecore.StorageService,
schedulerService schedulercore.SchedulerService,
eventsService eventscore.EventsService,
) {
	// Public Landing Page
	router.HandleFunc("/", handleLandingPage).Methods("GET")

	router.HandleFunc("/admin", handleDashboard).Methods("GET")
	router.HandleFunc("/admin/", handleDashboard).Methods("GET")

	// Module Routes
	
	
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
	
	
	
	router.HandleFunc("/admin/invoice", func(w http.ResponseWriter, r *http.Request) {
		handleInvoicePage(w, r, invoiceService)
	}).Methods("GET")
	router.HandleFunc("/admin/invoice/new", handleInvoiceNew).Methods("GET")
	router.HandleFunc("/admin/invoice/search", func(w http.ResponseWriter, r *http.Request) {
		handleInvoiceSearch(w, r, invoiceService)
	}).Methods("GET")
	router.HandleFunc("/admin/invoice/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleInvoiceEdit(w, r, invoiceService)
	}).Methods("GET")
	router.HandleFunc("/admin/invoice/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleInvoiceDelete(w, r, invoiceService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/mail", func(w http.ResponseWriter, r *http.Request) {
		handleMailPage(w, r, mailService)
	}).Methods("GET")
	router.HandleFunc("/admin/mail/new", handleMailNew).Methods("GET")
	router.HandleFunc("/admin/mail/search", func(w http.ResponseWriter, r *http.Request) {
		handleMailSearch(w, r, mailService)
	}).Methods("GET")
	router.HandleFunc("/admin/mail/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleMailEdit(w, r, mailService)
	}).Methods("GET")
	router.HandleFunc("/admin/mail/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleMailDelete(w, r, mailService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/cache", func(w http.ResponseWriter, r *http.Request) {
		handleCachePage(w, r, cacheService)
	}).Methods("GET")
	router.HandleFunc("/admin/cache/new", handleCacheNew).Methods("GET")
	router.HandleFunc("/admin/cache/search", func(w http.ResponseWriter, r *http.Request) {
		handleCacheSearch(w, r, cacheService)
	}).Methods("GET")
	router.HandleFunc("/admin/cache/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleCacheEdit(w, r, cacheService)
	}).Methods("GET")
	router.HandleFunc("/admin/cache/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleCacheDelete(w, r, cacheService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/storage", func(w http.ResponseWriter, r *http.Request) {
		handleStoragePage(w, r, storageService)
	}).Methods("GET")
	router.HandleFunc("/admin/storage/new", handleStorageNew).Methods("GET")
	router.HandleFunc("/admin/storage/search", func(w http.ResponseWriter, r *http.Request) {
		handleStorageSearch(w, r, storageService)
	}).Methods("GET")
	router.HandleFunc("/admin/storage/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleStorageEdit(w, r, storageService)
	}).Methods("GET")
	router.HandleFunc("/admin/storage/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleStorageDelete(w, r, storageService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/scheduler", func(w http.ResponseWriter, r *http.Request) {
		handleSchedulerPage(w, r, schedulerService)
	}).Methods("GET")
	router.HandleFunc("/admin/scheduler/new", handleSchedulerNew).Methods("GET")
	router.HandleFunc("/admin/scheduler/search", func(w http.ResponseWriter, r *http.Request) {
		handleSchedulerSearch(w, r, schedulerService)
	}).Methods("GET")
	router.HandleFunc("/admin/scheduler/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleSchedulerEdit(w, r, schedulerService)
	}).Methods("GET")
	router.HandleFunc("/admin/scheduler/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleSchedulerDelete(w, r, schedulerService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/events", func(w http.ResponseWriter, r *http.Request) {
		handleEventsPage(w, r, eventsService)
	}).Methods("GET")
	router.HandleFunc("/admin/events/new", handleEventsNew).Methods("GET")
	router.HandleFunc("/admin/events/search", func(w http.ResponseWriter, r *http.Request) {
		handleEventsSearch(w, r, eventsService)
	}).Methods("GET")
	router.HandleFunc("/admin/events/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleEventsEdit(w, r, eventsService)
	}).Methods("GET")
	router.HandleFunc("/admin/events/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleEventsDelete(w, r, eventsService)
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



func handleInvoicePage(w http.ResponseWriter, r *http.Request, service invoicecore.InvoiceService) {
	// Fetch real items
	filter := invoicecore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListInvoices(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]invoicecore.Invoice, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.InvoiceTableRows(items))
		return
	}

	renderTempl(w, r, pages.InvoicesPage(items))
}

func handleInvoiceNew(w http.ResponseWriter, r *http.Request) {
	item := invoicecore.Invoice{}
	renderTempl(w, r, components.InvoiceFormModal(&item, false))
}

func handleInvoiceEdit(w http.ResponseWriter, r *http.Request, service invoicecore.InvoiceService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetInvoiceByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.InvoiceFormModal(item, true))
}

func handleInvoiceDelete(w http.ResponseWriter, r *http.Request, service invoicecore.InvoiceService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetInvoiceByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Invoice", uint(id), "/api/v1/invoice/" + vars["id"]))
}

func handleInvoiceSearch(w http.ResponseWriter, r *http.Request, service invoicecore.InvoiceService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := invoicecore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListInvoices(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]invoicecore.Invoice, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.InvoiceTableRows(items))
}



func handleMailPage(w http.ResponseWriter, r *http.Request, service mailcore.MailService) {
	// Fetch real items
	filter := mailcore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListMails(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]mailcore.Mail, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.MailTableRows(items))
		return
	}

	renderTempl(w, r, pages.MailsPage(items))
}

func handleMailNew(w http.ResponseWriter, r *http.Request) {
	item := mailcore.Mail{}
	renderTempl(w, r, components.MailFormModal(&item, false))
}

func handleMailEdit(w http.ResponseWriter, r *http.Request, service mailcore.MailService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetMailByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.MailFormModal(item, true))
}

func handleMailDelete(w http.ResponseWriter, r *http.Request, service mailcore.MailService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetMailByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Mail", uint(id), "/api/v1/mail/" + vars["id"]))
}

func handleMailSearch(w http.ResponseWriter, r *http.Request, service mailcore.MailService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := mailcore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListMails(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]mailcore.Mail, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.MailTableRows(items))
}



func handleCachePage(w http.ResponseWriter, r *http.Request, service cachecore.CacheService) {
	// Fetch real items
	filter := cachecore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListCaches(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]cachecore.Cache, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.CacheTableRows(items))
		return
	}

	renderTempl(w, r, pages.CachesPage(items))
}

func handleCacheNew(w http.ResponseWriter, r *http.Request) {
	item := cachecore.Cache{}
	renderTempl(w, r, components.CacheFormModal(&item, false))
}

func handleCacheEdit(w http.ResponseWriter, r *http.Request, service cachecore.CacheService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetCacheByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.CacheFormModal(item, true))
}

func handleCacheDelete(w http.ResponseWriter, r *http.Request, service cachecore.CacheService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetCacheByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Cache", uint(id), "/api/v1/cache/" + vars["id"]))
}

func handleCacheSearch(w http.ResponseWriter, r *http.Request, service cachecore.CacheService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := cachecore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListCaches(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]cachecore.Cache, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.CacheTableRows(items))
}



func handleStoragePage(w http.ResponseWriter, r *http.Request, service storagecore.StorageService) {
	// Fetch real items
	filter := storagecore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListStorages(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]storagecore.Storage, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.StorageTableRows(items))
		return
	}

	renderTempl(w, r, pages.StoragesPage(items))
}

func handleStorageNew(w http.ResponseWriter, r *http.Request) {
	item := storagecore.Storage{}
	renderTempl(w, r, components.StorageFormModal(&item, false))
}

func handleStorageEdit(w http.ResponseWriter, r *http.Request, service storagecore.StorageService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetStorageByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.StorageFormModal(item, true))
}

func handleStorageDelete(w http.ResponseWriter, r *http.Request, service storagecore.StorageService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetStorageByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Storage", uint(id), "/api/v1/storage/" + vars["id"]))
}

func handleStorageSearch(w http.ResponseWriter, r *http.Request, service storagecore.StorageService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := storagecore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListStorages(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]storagecore.Storage, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.StorageTableRows(items))
}



func handleSchedulerPage(w http.ResponseWriter, r *http.Request, service schedulercore.SchedulerService) {
	// Fetch real items
	filter := schedulercore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListSchedulers(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]schedulercore.Scheduler, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.SchedulerTableRows(items))
		return
	}

	renderTempl(w, r, pages.SchedulersPage(items))
}

func handleSchedulerNew(w http.ResponseWriter, r *http.Request) {
	item := schedulercore.Scheduler{}
	renderTempl(w, r, components.SchedulerFormModal(&item, false))
}

func handleSchedulerEdit(w http.ResponseWriter, r *http.Request, service schedulercore.SchedulerService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetSchedulerByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.SchedulerFormModal(item, true))
}

func handleSchedulerDelete(w http.ResponseWriter, r *http.Request, service schedulercore.SchedulerService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetSchedulerByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Scheduler", uint(id), "/api/v1/scheduler/" + vars["id"]))
}

func handleSchedulerSearch(w http.ResponseWriter, r *http.Request, service schedulercore.SchedulerService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := schedulercore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListSchedulers(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]schedulercore.Scheduler, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.SchedulerTableRows(items))
}



func handleEventsPage(w http.ResponseWriter, r *http.Request, service eventscore.EventsService) {
	// Fetch real items
	filter := eventscore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListEvents(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]eventscore.Events, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.EventsTableRows(items))
		return
	}

	renderTempl(w, r, pages.EventsPage(items))
}

func handleEventsNew(w http.ResponseWriter, r *http.Request) {
	item := eventscore.Events{}
	renderTempl(w, r, components.EventsFormModal(&item, false))
}

func handleEventsEdit(w http.ResponseWriter, r *http.Request, service eventscore.EventsService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetEventsByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.EventsFormModal(item, true))
}

func handleEventsDelete(w http.ResponseWriter, r *http.Request, service eventscore.EventsService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetEventsByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Events", uint(id), "/api/v1/events/" + vars["id"]))
}

func handleEventsSearch(w http.ResponseWriter, r *http.Request, service eventscore.EventsService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := eventscore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListEvents(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]eventscore.Events, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.EventsTableRows(items))
}



func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
