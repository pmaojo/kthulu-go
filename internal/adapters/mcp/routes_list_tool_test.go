package mcp_test

import (
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

const routesGinFixture = `package server

import "github.com/gin-gonic/gin"

func SetupRouter(r *gin.Engine) {
	r.GET("/api/orders", OrderHandler.FindAll)
	r.POST("/api/orders", OrderHandler.Create)
	r.GET("/api/orders/:id", OrderHandler.FindByID)
	r.PATCH("/api/orders/:id/status", authMiddleware(OrderHandler.Update))
	r.DELETE("/api/orders/:id", adminMiddleware(authMiddleware(OrderHandler.Delete)))
}
`

const routesChiFixture = `package server

import "github.com/go-chi/chi/v5"

func SetupRouter(r chi.Router) {
	r.Get("/api/products", ProductHandler.List)
	r.Post("/api/products", ProductHandler.Create)
}
`

func TestRoutesListFindsGinRoutes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesGinFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)
	require.NotEmpty(t, routes)

	byPath := routesByPath(routes)
	require.Contains(t, byPath, "/api/orders")
	require.Contains(t, byPath, "/api/orders/:id")
	require.Contains(t, byPath, "/api/orders/:id/status")

	// Verify methods are parsed correctly.
	var methods []string
	for _, r := range routes {
		methods = append(methods, r.Method)
	}
	require.Contains(t, methods, "GET")
	require.Contains(t, methods, "POST")
	require.Contains(t, methods, "PATCH")
	require.Contains(t, methods, "DELETE")
}

func TestRoutesListDetectsMiddleware(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesGinFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)

	var middlewares []string
	for _, r := range routes {
		middlewares = append(middlewares, r.Middleware)
	}
	// At least one route has authMiddleware.
	combined := strings.Join(middlewares, " ")
	require.Contains(t, combined, "authMiddleware")
}

func TestRoutesListChiRoutes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesChiFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)
	require.NotEmpty(t, routes)

	byPath := routesByPath(routes)
	require.Contains(t, byPath, "/api/products")
}

func TestRoutesListModuleFilter(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesGinFixture)
	writeTestFile(t, dir, "internal/products/router.go", routesChiFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)

	filtered := mcp.FilterRoutesByModule(routes, "orders")
	require.NotEmpty(t, filtered)
	for _, r := range filtered {
		require.True(t,
			strings.Contains(strings.ToLower(r.Module), "orders") ||
				strings.Contains(strings.ToLower(r.Path), "orders"),
			"expected route to match 'orders': %+v", r)
	}

	// Products should not appear.
	for _, r := range filtered {
		require.NotContains(t, r.Path, "/api/products")
	}
}

func TestRoutesListFormatsAsTable(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesGinFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)

	table := mcp.FormatRoutesTable(routes)
	require.Contains(t, table, "METHOD")
	require.Contains(t, table, "PATH")
	require.Contains(t, table, "HANDLER")
	require.Contains(t, table, "GET")
	require.Contains(t, table, "/api/orders")
}

func TestRoutesListFormatsAsJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "internal/server/router.go", routesGinFixture)

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)

	output, err := mcp.FormatRoutesJSON(routes)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(strings.TrimSpace(output), "["), "expected JSON array, got: %s", output)
	require.Contains(t, output, `"method"`)
	require.Contains(t, output, `"path"`)
}

func TestRoutesListNoRoutes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	routes, err := mcp.DiscoverRoutes(dir)
	require.NoError(t, err)
	require.Empty(t, routes)
}

// routesByPath returns a set of paths for easy lookup.
func routesByPath(routes []mcp.RouteEntry) map[string]bool {
	m := make(map[string]bool, len(routes))
	for _, r := range routes {
		m[r.Path] = true
	}
	return m
}
