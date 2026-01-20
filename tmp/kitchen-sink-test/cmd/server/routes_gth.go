// @kthulu:gth:routes
package main

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/a-h/templ"

	"kitchen-sink-test/internal/views/pages"
	"kitchen-sink-test/internal/views/components"
	"kitchen-sink-test/internal/views/partials"

	// Module Core Imports
	
	
	
	
	calendarcore "kitchen-sink-test/internal/modules/calendar/core"
	
	
	
	contactcore "kitchen-sink-test/internal/modules/contact/core"
	
	
	
	inventorycore "kitchen-sink-test/internal/modules/inventory/core"
	
	
	
	invoicecore "kitchen-sink-test/internal/modules/invoice/core"
	
	
	
	
	
	productcore "kitchen-sink-test/internal/modules/product/core"
	
	
	
	
	
	verifactucore "kitchen-sink-test/internal/modules/verifactu/core"
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	organizationcore "kitchen-sink-test/internal/modules/organization/core"
	
	
	
	
	
	usercore "kitchen-sink-test/internal/modules/user/core"
	
	
	
	
)

// RegisterGTHRoutes registers the GTH (HTML) web routes
func RegisterGTHRoutes(router *mux.Router, 
	calendarService calendarcore.CalendarService,
contactService contactcore.ContactService,
inventoryService inventorycore.InventoryService,
invoiceService invoicecore.InvoiceService,
organizationService organizationcore.OrganizationService,
productService productcore.ProductService,
userService usercore.UserService,
verifactuService verifactucore.VerifactuService,
) {
	// Dashboard at root
	router.HandleFunc("/", handleDashboard).Methods("GET")
	router.HandleFunc("/admin", handleDashboard).Methods("GET")
	router.HandleFunc("/admin/", handleDashboard).Methods("GET")

	// Module Routes
	
	
	
	
	router.HandleFunc("/admin/calendar", func(w http.ResponseWriter, r *http.Request) {
		handleCalendarPage(w, r, calendarService)
	}).Methods("GET")
	router.HandleFunc("/admin/calendar/new", handleCalendarNew).Methods("GET")
	router.HandleFunc("/admin/calendar/search", func(w http.ResponseWriter, r *http.Request) {
		handleCalendarSearch(w, r, calendarService)
	}).Methods("GET")
	router.HandleFunc("/admin/calendar/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleCalendarEdit(w, r, calendarService)
	}).Methods("GET")
	router.HandleFunc("/admin/calendar/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleCalendarDelete(w, r, calendarService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/contact", func(w http.ResponseWriter, r *http.Request) {
		handleContactPage(w, r, contactService)
	}).Methods("GET")
	router.HandleFunc("/admin/contact/new", handleContactNew).Methods("GET")
	router.HandleFunc("/admin/contact/search", func(w http.ResponseWriter, r *http.Request) {
		handleContactSearch(w, r, contactService)
	}).Methods("GET")
	router.HandleFunc("/admin/contact/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleContactEdit(w, r, contactService)
	}).Methods("GET")
	router.HandleFunc("/admin/contact/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleContactDelete(w, r, contactService)
	}).Methods("GET")
	
	
	
	router.HandleFunc("/admin/inventory", func(w http.ResponseWriter, r *http.Request) {
		handleInventoryPage(w, r, inventoryService)
	}).Methods("GET")
	router.HandleFunc("/admin/inventory/new", handleInventoryNew).Methods("GET")
	router.HandleFunc("/admin/inventory/search", func(w http.ResponseWriter, r *http.Request) {
		handleInventorySearch(w, r, inventoryService)
	}).Methods("GET")
	router.HandleFunc("/admin/inventory/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleInventoryEdit(w, r, inventoryService)
	}).Methods("GET")
	router.HandleFunc("/admin/inventory/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleInventoryDelete(w, r, inventoryService)
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
	
	
	
	router.HandleFunc("/admin/organization", func(w http.ResponseWriter, r *http.Request) {
		handleOrganizationPage(w, r, organizationService)
	}).Methods("GET")
	router.HandleFunc("/admin/organization/new", handleOrganizationNew).Methods("GET")
	router.HandleFunc("/admin/organization/search", func(w http.ResponseWriter, r *http.Request) {
		handleOrganizationSearch(w, r, organizationService)
	}).Methods("GET")
	router.HandleFunc("/admin/organization/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleOrganizationEdit(w, r, organizationService)
	}).Methods("GET")
	router.HandleFunc("/admin/organization/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleOrganizationDelete(w, r, organizationService)
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
	
	
	
	router.HandleFunc("/admin/verifactu", func(w http.ResponseWriter, r *http.Request) {
		handleVerifactuPage(w, r, verifactuService)
	}).Methods("GET")
	router.HandleFunc("/admin/verifactu/new", handleVerifactuNew).Methods("GET")
	router.HandleFunc("/admin/verifactu/search", func(w http.ResponseWriter, r *http.Request) {
		handleVerifactuSearch(w, r, verifactuService)
	}).Methods("GET")
	router.HandleFunc("/admin/verifactu/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
		handleVerifactuEdit(w, r, verifactuService)
	}).Methods("GET")
	router.HandleFunc("/admin/verifactu/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		handleVerifactuDelete(w, r, verifactuService)
	}).Methods("GET")
	
	
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats := []pages.DashboardStats{
		{Label: "System Status", Value: "Online", Icon: "🟢", Color: "#dcfce7"},
	}
	renderTempl(w, r, pages.DashboardPage(stats))
}





func handleCalendarPage(w http.ResponseWriter, r *http.Request, service calendarcore.CalendarService) {
	// Fetch real items
	filter := calendarcore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListCalendars(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]calendarcore.Calendar, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.CalendarTableRows(items))
		return
	}

	renderTempl(w, r, pages.CalendarsPage(items))
}

func handleCalendarNew(w http.ResponseWriter, r *http.Request) {
	item := calendarcore.Calendar{}
	renderTempl(w, r, components.CalendarFormModal(&item, false))
}

func handleCalendarEdit(w http.ResponseWriter, r *http.Request, service calendarcore.CalendarService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetCalendarByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.CalendarFormModal(item, true))
}

func handleCalendarDelete(w http.ResponseWriter, r *http.Request, service calendarcore.CalendarService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetCalendarByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Calendar", uint(id), "/api/v1/calendar/" + vars["id"]))
}

func handleCalendarSearch(w http.ResponseWriter, r *http.Request, service calendarcore.CalendarService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := calendarcore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListCalendars(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]calendarcore.Calendar, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.CalendarTableRows(items))
}



func handleContactPage(w http.ResponseWriter, r *http.Request, service contactcore.ContactService) {
	// Fetch real items
	filter := contactcore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListContacts(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]contactcore.Contact, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.ContactTableRows(items))
		return
	}

	renderTempl(w, r, pages.ContactsPage(items))
}

func handleContactNew(w http.ResponseWriter, r *http.Request) {
	item := contactcore.Contact{}
	renderTempl(w, r, components.ContactFormModal(&item, false))
}

func handleContactEdit(w http.ResponseWriter, r *http.Request, service contactcore.ContactService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetContactByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ContactFormModal(item, true))
}

func handleContactDelete(w http.ResponseWriter, r *http.Request, service contactcore.ContactService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetContactByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Contact", uint(id), "/api/v1/contact/" + vars["id"]))
}

func handleContactSearch(w http.ResponseWriter, r *http.Request, service contactcore.ContactService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := contactcore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListContacts(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]contactcore.Contact, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.ContactTableRows(items))
}



func handleInventoryPage(w http.ResponseWriter, r *http.Request, service inventorycore.InventoryService) {
	// Fetch real items
	filter := inventorycore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListInventories(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]inventorycore.Inventory, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.InventoryTableRows(items))
		return
	}

	renderTempl(w, r, pages.InventoriesPage(items))
}

func handleInventoryNew(w http.ResponseWriter, r *http.Request) {
	item := inventorycore.Inventory{}
	renderTempl(w, r, components.InventoryFormModal(&item, false))
}

func handleInventoryEdit(w http.ResponseWriter, r *http.Request, service inventorycore.InventoryService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetInventoryByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.InventoryFormModal(item, true))
}

func handleInventoryDelete(w http.ResponseWriter, r *http.Request, service inventorycore.InventoryService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetInventoryByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Inventory", uint(id), "/api/v1/inventory/" + vars["id"]))
}

func handleInventorySearch(w http.ResponseWriter, r *http.Request, service inventorycore.InventoryService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := inventorycore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListInventories(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]inventorycore.Inventory, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.InventoryTableRows(items))
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



func handleOrganizationPage(w http.ResponseWriter, r *http.Request, service organizationcore.OrganizationService) {
	// Fetch real items
	filter := organizationcore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListOrganizations(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]organizationcore.Organization, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.OrganizationTableRows(items))
		return
	}

	renderTempl(w, r, pages.OrganizationsPage(items))
}

func handleOrganizationNew(w http.ResponseWriter, r *http.Request) {
	item := organizationcore.Organization{}
	renderTempl(w, r, components.OrganizationFormModal(&item, false))
}

func handleOrganizationEdit(w http.ResponseWriter, r *http.Request, service organizationcore.OrganizationService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetOrganizationByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.OrganizationFormModal(item, true))
}

func handleOrganizationDelete(w http.ResponseWriter, r *http.Request, service organizationcore.OrganizationService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetOrganizationByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Organization", uint(id), "/api/v1/organization/" + vars["id"]))
}

func handleOrganizationSearch(w http.ResponseWriter, r *http.Request, service organizationcore.OrganizationService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := organizationcore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListOrganizations(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]organizationcore.Organization, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.OrganizationTableRows(items))
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



func handleVerifactuPage(w http.ResponseWriter, r *http.Request, service verifactucore.VerifactuService) {
	// Fetch real items
	filter := verifactucore.SearchFilter{
		Limit: 100,
	}
	itemsPtr, err := service.ListVerifactus(filter)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}
	
	// Convert pointers to values for template
	items := make([]verifactucore.Verifactu, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	
	if r.Header.Get("HX-Target") == "table-body" {
		renderTempl(w, r, partials.VerifactuTableRows(items))
		return
	}

	renderTempl(w, r, pages.VerifactusPage(items))
}

func handleVerifactuNew(w http.ResponseWriter, r *http.Request) {
	item := verifactucore.Verifactu{}
	renderTempl(w, r, components.VerifactuFormModal(&item, false))
}

func handleVerifactuEdit(w http.ResponseWriter, r *http.Request, service verifactucore.VerifactuService) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	item, err := service.GetVerifactuByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.VerifactuFormModal(item, true))
}

func handleVerifactuDelete(w http.ResponseWriter, r *http.Request, service verifactucore.VerifactuService) {
	vars := mux.Vars(r)
	// We just need the ID for the confirmation modal, validating existence is optional but good practice
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = service.GetVerifactuByID(uint(id))
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	renderTempl(w, r, components.ConfirmDelete("Verifactu", uint(id), "/api/v1/verifactu/" + vars["id"]))
}

func handleVerifactuSearch(w http.ResponseWriter, r *http.Request, service verifactucore.VerifactuService) {
	query := r.URL.Query().Get("sort") // Using sort param for now as per previous search implementation
	if q := r.URL.Query().Get("q"); q != "" {
		query = q
	}

	filter := verifactucore.SearchFilter{
		Query: query,
		Limit: 100,
	}
	itemsPtr, err := service.ListVerifactus(filter)
	if err != nil {
		http.Error(w, "Failed to search items", http.StatusInternalServerError)
		return
	}

	items := make([]verifactucore.Verifactu, len(itemsPtr))
	for i, item := range itemsPtr {
		items[i] = *item
	}
	renderTempl(w, r, partials.VerifactuTableRows(items))
}



func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
