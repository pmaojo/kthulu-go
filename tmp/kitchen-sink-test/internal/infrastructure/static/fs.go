package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist/*
var content embed.FS

// GetFileSystem returns the embedded file system.
// It expects a "dist" directory to be present in the embedded content.
func GetFileSystem() fs.FS {
	f, err := fs.Sub(content, "dist")
	if err != nil {
		return content
	}
	return f
}

// ServeMiddleware handles serving static assets with pre-compression support (Brotli/Gzip).
// It checks Accept-Encoding headers and serves the corresponding .br or .gz file if it exists.
func ServeMiddleware(next http.Handler) http.Handler {
	fileServer := http.FileServer(http.FS(GetFileSystem()))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is for an API endpoint or explicitly not static, skip (but here we usually mount at specific prefix or fallback)
		// Assuming this is used as a fallback or specific route handler.

		urlPath := r.URL.Path
		if urlPath == "/" {
			urlPath = "index.html"
		}

		// Clean path to prevent directory traversal is handled by http.FS but good to be careful
		cleanPath := path.Clean(strings.TrimPrefix(urlPath, "/"))

		// Check for Brotli
		if strings.Contains(r.Header.Get("Accept-Encoding"), "br") {
			if f, err := GetFileSystem().Open(cleanPath + ".br"); err == nil {
				f.Close()
				w.Header().Set("Content-Encoding", "br")
				w.Header().Set("Content-Type", mimeTypeForFile(cleanPath))
				r.URL.Path += ".br"
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Check for Gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			if f, err := GetFileSystem().Open(cleanPath + ".gz"); err == nil {
				f.Close()
				w.Header().Set("Content-Encoding", "gzip")
				w.Header().Set("Content-Type", mimeTypeForFile(cleanPath))
				r.URL.Path += ".gz"
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Standard file server fallback
		fileServer.ServeHTTP(w, r)
	})
}

// mimeTypeForFile acts as a simple mime sniffer based on extension since serving compressed files
// might lose the extension-based content-type detection of http.FileServer
func mimeTypeForFile(f string) string {
	ext := path.Ext(f)
	switch ext {
	case ".html": return "text/html"
	case ".css": return "text/css"
	case ".js": return "application/javascript"
	case ".json": return "application/json"
	case ".png": return "image/png"
	case ".jpg", ".jpeg": return "image/jpeg"
	case ".svg": return "image/svg+xml"
	case ".ico": return "image/x-icon"
	}
	return "application/octet-stream"
}
