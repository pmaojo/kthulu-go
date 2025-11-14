// @kthulu:core
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	URL             string
	Driver          string // "sqlite" or "postgres"
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// JWTConfig holds JWT token configuration
type JWTConfig struct {
	Secret          string
	RefreshSecret   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// SMTPConfig holds email notification configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Enabled  bool
}

// FeatureFlagConfig holds configuration for feature flag providers
type FeatureFlagConfig struct {
	Provider    string
	URL         string
	APIKey      string
	Environment string
	AppName     string
}

// SentryConfig holds configuration for Sentry error tracking
type SentryConfig struct {
	DSN     string
	Enabled bool
}

// ObservabilityConfig holds tracing and metrics configuration.
// TraceExporter selects the tracing backend: "stdout" (default) or "jaeger".
// MetricsExporter selects the metrics backend (e.g., "prometheus").
type ObservabilityConfig struct {
	TraceSampleRate float64
	TraceExporter   string
	MetricsExporter string
}

// RateLimitConfig holds configuration for request rate limiting.
// Defaults: RequestsPerSecond=10, Burst=20.
type RateLimitConfig struct {
	// RequestsPerSecond is the steady rate of requests allowed per second (default 10).
	RequestsPerSecond float64
	// Burst is the maximum number of requests that can be allowed in a burst (default 20).
	Burst int
}

// Config holds application-wide configuration values.
type Config struct {
	Version          string
	Database         DatabaseConfig
	Server           ServerConfig
	JWT              JWTConfig
	SMTP             SMTPConfig
	FeatureFlags     FeatureFlagConfig
	Sentry           SentryConfig
	Observability    ObservabilityConfig
	Env              string
	AuthProvider     string
	Modules          []string
	VerifactuSIFCode string // Two-character SIF code for VeriFactu
	VerifactuMode    string
	RateLimit        RateLimitConfig
}

const databaseURLEnv = "DATABASE_URL"

// NewConfig loads environment variables and constructs Config.
// It validates all required configuration and provides sensible defaults.
func NewConfig() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Overload(".env.local")
	if err := loadVaultEnv(); err != nil {
		return nil, err
	}
	config := &Config{}
	config.Version = getEnvWithDefault("SERVICE_VERSION", Version)

	// Environment
	config.Env = getEnvWithDefault("ENV", "development")
	config.AuthProvider = getEnvWithDefault("AUTH_PROVIDER", "auth")

	// Active modules configuration (comma-separated list)
	if modules := os.Getenv("MODULES"); modules != "" {
		config.Modules = strings.Split(modules, ",")
		for i := range config.Modules {
			config.Modules[i] = strings.TrimSpace(config.Modules[i])
		}
	}
	// VeriFactu configuration
	config.VerifactuSIFCode = getEnvWithDefault("VERIFACTU_SIF_CODE", "01")
	config.VerifactuMode = getEnvWithDefault("VERIFACTU_MODE", "queued")

	// Feature flag configuration
	config.FeatureFlags = FeatureFlagConfig{
		Provider:    getEnvWithDefault("FF_PROVIDER", ""),
		URL:         getEnvWithDefault("FF_URL", ""),
		APIKey:      getEnvWithDefault("FF_API_KEY", ""),
		Environment: getEnvWithDefault("FF_ENVIRONMENT", ""),
		AppName:     getEnvWithDefault("FF_APP_NAME", "kthulu"),
	}

	// Sentry configuration
	sentryEnabled, _ := strconv.ParseBool(getEnvWithDefault("SENTRY_ENABLED", "false"))
	config.Sentry = SentryConfig{
		DSN:     getEnvWithDefault("SENTRY_DSN", ""),
		Enabled: sentryEnabled,
	}

	// Observability configuration
	traceSampleRate, err := strconv.ParseFloat(getEnvWithDefault("TRACE_SAMPLE_RATE", "1"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid TRACE_SAMPLE_RATE: %w", err)
	}
	config.Observability = ObservabilityConfig{
		TraceSampleRate: traceSampleRate,
		TraceExporter:   strings.ToLower(getEnvWithDefault("TRACE_EXPORTER", "stdout")),
		MetricsExporter: getEnvWithDefault("METRICS_EXPORTER", "prometheus"),
	}

	// Rate limiting configuration
	rps, err := strconv.ParseFloat(getEnvWithDefault("RATE_LIMIT_RPS", "10"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_RPS: %w", err)
	}

	burst, err := strconv.Atoi(getEnvWithDefault("RATE_LIMIT_BURST", "20"))
	if err != nil {
		return nil, fmt.Errorf("invalid RATE_LIMIT_BURST: %w", err)
	}

	config.RateLimit = RateLimitConfig{
		RequestsPerSecond: rps,
		Burst:             burst,
	}

	// Database configuration - Optimal: SQLite by default
	dbDriver := getEnvWithDefault("DB_DRIVER", "sqlite")
	var dbURL string

	switch dbDriver {
	case "sqlite":
		dbURL = getEnvWithDefault(databaseURLEnv, "./kthulu.db")
	case "postgres":
		dbURL = os.Getenv(databaseURLEnv)
		if dbURL == "" {
			return nil, errors.New("DATABASE_URL is required when using postgres driver")
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s (supported: sqlite, postgres)", dbDriver)
	}

	maxOpenConns, err := strconv.Atoi(getEnvWithDefault("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := strconv.Atoi(getEnvWithDefault("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetime, err := time.ParseDuration(getEnvWithDefault("DB_CONN_MAX_LIFETIME", "1h"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %w", err)
	}

	config.Database = DatabaseConfig{
		URL:             dbURL,
		Driver:          dbDriver,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}

	// Server configuration
	readTimeout, err := time.ParseDuration(getEnvWithDefault("SERVER_READ_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := time.ParseDuration(getEnvWithDefault("SERVER_WRITE_TIMEOUT", "15s"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_WRITE_TIMEOUT: %w", err)
	}

	idleTimeout, err := time.ParseDuration(getEnvWithDefault("SERVER_IDLE_TIMEOUT", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_IDLE_TIMEOUT: %w", err)
	}

	config.Server = ServerConfig{
		Addr:         getEnvWithDefault("HTTP_ADDR", ":8080"),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		return nil, errors.New("JWT_REFRESH_SECRET is required")
	}

	accessTokenTTL, err := time.ParseDuration(getEnvWithDefault("JWT_ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTokenTTL, err := time.ParseDuration(getEnvWithDefault("JWT_REFRESH_TOKEN_TTL", "7d"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_TTL: %w", err)
	}

	config.JWT = JWTConfig{
		Secret:          jwtSecret,
		RefreshSecret:   jwtRefreshSecret,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}

	// SMTP configuration
	smtpEnabled, _ := strconv.ParseBool(getEnvWithDefault("SMTP_ENABLED", "false"))
	smtpPort, err := strconv.Atoi(getEnvWithDefault("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	config.SMTP = SMTPConfig{
		Host:     getEnvWithDefault("SMTP_HOST", "localhost"),
		Port:     smtpPort,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     getEnvWithDefault("SMTP_FROM", "noreply@kthulu.local"),
		Enabled:  smtpEnabled,
	}

	// Validate SMTP configuration if enabled
	if config.SMTP.Enabled {
		if config.SMTP.Host == "" {
			return nil, errors.New("SMTP_HOST is required when SMTP is enabled")
		}
		if config.SMTP.Username == "" {
			return nil, errors.New("SMTP_USERNAME is required when SMTP is enabled")
		}
		if config.SMTP.Password == "" {
			return nil, errors.New("SMTP_PASSWORD is required when SMTP is enabled")
		}
	}

	return config, nil
}

// getEnvWithDefault returns the value of the environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// loadVaultEnv loads secrets from HashiCorp Vault if VAULT_SECRET_PATH is set.
// It expects the `vault` CLI and parses its JSON output.
func loadVaultEnv() error {
	path := os.Getenv("VAULT_SECRET_PATH")
	if path == "" {
		return nil
	}
	if _, err := exec.LookPath("vault"); err != nil {
		return fmt.Errorf("vault CLI not found: %w", err)
	}
	cmd := exec.Command("vault", "kv", "get", "-format=json", path)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("vault fetch: %w", err)
	}
	var resp struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		return fmt.Errorf("vault parse: %w", err)
	}
	for k, v := range resp.Data.Data {
		os.Setenv(k, v)
	}
	return nil
}
