package core

import (
	"fmt"
	"log"
	"os"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

func CoreRepositoryProviders() fx.Option {
        return fx.Options(
                fx.Provide(NewDatabase),
        )
}

func NewDatabase() (*gorm.DB, error) {
if os.Getenv("KTHULU_TEST_MODE") == "1" {
log.Println("Using in-memory SQLite database for tests")
return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
}
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "postgres"),
			getEnv("DB_NAME", "my-kthulu-app"),
		)
		log.Printf("Connecting to PostgreSQL at %s:%s/%s", getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"), getEnv("DB_NAME", "my-kthulu-app"))
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})

}

func getEnv(key, fallback string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return fallback
}
