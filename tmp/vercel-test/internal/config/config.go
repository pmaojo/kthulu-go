package config

import (
    "os"
)

type DatabaseConfig struct {
    URL    string
    Driver string
}

type Config struct {
    Database DatabaseConfig
    Env      string
}

func NewConfig() (*Config, error) {
    dbDriver := getEnv("DB_DRIVER", "sqlite")
    dbURL := os.Getenv("DATABASE_URL")

    // Default DSN construction if not provided (Development Defaults)
    if dbURL == "" {
        
        dbURL = getEnv("SQLITE_PATH", "data/vercel-test.db")
        
    }

    return &Config{
        Database: DatabaseConfig{
            URL:    dbURL,
            Driver: dbDriver,
        },
        Env: getEnv("ENV", "development"),
    }, nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
