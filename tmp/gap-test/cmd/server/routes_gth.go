// @kthulu:gth:routes
package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/a-h/templ"

	"gap-test/internal/views/layouts"
	"gap-test/internal/views/pages"
)

// RegisterGTHRoutes registers the GTH (HTML) web routes
func RegisterGTHRoutes(router *mux.Router) {
	// Dashboard at root
	router.HandleFunc("/", handleDashboard).Methods("GET")
	router.HandleFunc("/admin", handleDashboard).Methods("GET")
	router.HandleFunc("/admin/", handleDashboard).Methods("GET")
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats := []pages.DashboardStats{
		{Label: "Total Items", Value: "0", Icon: "📦", Color: "#e0f2fe"},
		{Label: "Active", Value: "0", Icon: "✓", Color: "#dcfce7"},
		{Label: "Pending", Value: "0", Icon: "⏳", Color: "#fef3c7"},
	}
	renderTempl(w, r, pages.DashboardPage(stats))
}

func renderTempl(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

// IsHTMXRequest checks if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
