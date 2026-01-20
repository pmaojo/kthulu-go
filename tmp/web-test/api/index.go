// @kthulu:vercel:entrypoint
package handler

import (
	"net/http"
	"os"
	"sync"
	
	"web-test/internal/infrastructure/server"
	"web-test/internal/infrastructure/database"
)

var (
	app  http.Handler
	once sync.Once
)

func init() {
	once.Do(func() {
		// Auto-migrate if enabled
		if os.Getenv("RUN_MIGRATIONS") == "true" {
			db, err := database.NewDB()
			if err == nil {
				database.AutoMigrate(db)
			}
		}
		
		// Initialize app handler
		app = server.NewHandler()
	})
}

// Handler is the Vercel serverless function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
