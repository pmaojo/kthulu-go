// @kthulu:core
package adapterhttp

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// StaticHandler serves the frontend application and static assets
type StaticHandler struct {
	logger    *zap.Logger
	staticDir string
	indexFile string
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(logger *zap.Logger) *StaticHandler {
	return &StaticHandler{
		logger:    logger,
		staticDir: "public",
		indexFile: "index.html",
	}
}

// RegisterRoutes registers the static file serving routes
func (h *StaticHandler) RegisterRoutes(r chi.Router) {
	h.logger.Info("Registering static file routes", zap.String("staticDir", h.staticDir))

	// Serve static assets (CSS, JS, images, etc.)
	r.Handle("/assets/*", h.serveStaticFiles())
	r.Handle("/favicon.ico", h.serveStaticFiles())
	r.Handle("/robots.txt", h.serveStaticFiles())
	r.Handle("/manifest.json", h.serveStaticFiles())

	// Catch-all route for SPA (Single Page Application)
	// This must be registered last to avoid conflicts with API routes
	r.NotFound(h.serveSPA())
}

// serveStaticFiles creates a handler for static assets
func (h *StaticHandler) serveStaticFiles() http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(h.staticDir))

	return func(w http.ResponseWriter, r *http.Request) {
		// Security: prevent directory traversal
		if strings.Contains(r.URL.Path, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Check if file exists
		filePath := filepath.Join(h.staticDir, r.URL.Path)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		// Set appropriate cache headers for static assets
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			// Cache static assets for 1 year (they should be versioned)
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// Cache other static files for 1 hour
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		// Serve the file
		fileServer.ServeHTTP(w, r)
	}
}

// serveSPA creates a handler for Single Page Application routing
func (h *StaticHandler) serveSPA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only serve SPA for GET requests to non-API routes
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Skip API routes - they should be handled by API handlers
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Skip health check and other system routes
		if strings.HasPrefix(r.URL.Path, "/health") ||
			strings.HasPrefix(r.URL.Path, "/metrics") ||
			strings.HasPrefix(r.URL.Path, "/docs") {
			http.NotFound(w, r)
			return
		}

		// Serve index.html for all other routes (SPA routing)
		indexPath := filepath.Join(h.staticDir, h.indexFile)

		// Check if index.html exists
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			h.logger.Warn("Frontend not built - index.html not found",
				zap.String("path", indexPath),
				zap.String("url", r.URL.Path))

			// Serve a helpful message if frontend is not built
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Kthulu - Frontend Not Built</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .logo { font-size: 2em; font-weight: bold; color: #2563eb; margin-bottom: 20px; }
        .message { color: #374151; line-height: 1.6; }
        .code { background: #f3f4f6; padding: 15px; border-radius: 4px; font-family: monospace; margin: 15px 0; }
        .api-link { color: #2563eb; text-decoration: none; }
        .api-link:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🐙 Kthulu Framework</div>
        <div class="message">
            <h2>Frontend Not Built</h2>
            <p>The Kthulu API is running successfully, but the frontend application hasn't been built yet.</p>
            
            <h3>To build the frontend:</h3>
            <div class="code">
                cd frontend<br>
                npm install<br>
                npm run build
            </div>
            
            <h3>API Documentation:</h3>
            <p>You can access the API documentation at: <a href="/docs" class="api-link">/docs</a></p>
            
            <h3>Health Check:</h3>
            <p>API health status: <a href="/health" class="api-link">/health</a></p>
        </div>
    </div>
</body>
</html>
			`))
			return
		}

		// Set headers for SPA
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Serve index.html
		http.ServeFile(w, r, indexPath)

		h.logger.Debug("Served SPA route",
			zap.String("path", r.URL.Path),
			zap.String("userAgent", r.Header.Get("User-Agent")))
	}
}

// SetStaticDir allows customizing the static files directory
func (h *StaticHandler) SetStaticDir(dir string) {
	h.staticDir = dir
	h.logger.Info("Static directory updated", zap.String("staticDir", dir))
}

// SetIndexFile allows customizing the index file name
func (h *StaticHandler) SetIndexFile(file string) {
	h.indexFile = file
	h.logger.Info("Index file updated", zap.String("indexFile", file))
}
