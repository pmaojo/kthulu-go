// @kthulu:core
package common

// Database drivers
const (
	SQLiteDriver   = "sqlite"
	PostgresDriver = "postgres"
	SQLite3Dialect = "sqlite3"
)

// Sort orders
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Default sort fields
const (
	DefaultSortField = "created_at"
)

// Health check paths
const (
	HealthzPath = "/healthz"
)

// Status values
const (
	StatusHealthy = "healthy"
)

// Security levels
const (
	SecurityLevelLow    = "LOW"
	SecurityLevelMedium = "MEDIUM"
	SecurityLevelHigh   = "HIGH"
)
