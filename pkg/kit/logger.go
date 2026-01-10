// @kthulu:core
package core

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the logging interface used throughout the application
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
	Sync() error
}

// zapLogger wraps zap.SugaredLogger to implement our Logger interface
type zapLogger struct {
	sugar *zap.SugaredLogger
}

// NewLoggerFromZap creates a Logger from a zap.Logger
func NewLoggerFromZap(logger *zap.Logger) Logger {
	return &zapLogger{sugar: logger.Sugar()}
}

func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.sugar.Debugw(msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.sugar.Infow(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.sugar.Warnw(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.sugar.Errorw(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.sugar.Fatalw(msg, fields...)
}

func (l *zapLogger) With(fields ...interface{}) Logger {
	return &zapLogger{sugar: l.sugar.With(fields...)}
}

func (l *zapLogger) Sync() error {
	return l.sugar.Sync()
}

// NewLogger builds a logger based on the configuration.
// It provides structured logging with appropriate settings for development and production.
func NewLogger(cfg *Config) (Logger, error) {
	zapLogger, err := NewZapLogger(cfg)
	if err != nil {
		return nil, err
	}
	return NewLoggerFromZap(zapLogger), nil
}

// NewZapLogger builds a zap logger based on the configuration.
// It provides structured logging with appropriate settings for development and production.
func NewZapLogger(cfg *Config) (*zap.Logger, error) {
	if cfg.IsDevelopment() {
		// Development logger with human-readable output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return config.Build()
	}

	// Production logger with JSON output
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	return config.Build()
}

// NewSugaredLogger creates a sugared logger for easier use with structured logging.
func NewSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

// LoggerConfig holds additional logging configuration
type LoggerConfig struct {
	Level       string `json:"level"`
	Development bool   `json:"development"`
	Encoding    string `json:"encoding"` // json or console
}

// NewLoggerWithConfig creates a logger with custom configuration
func NewLoggerWithConfig(cfg *Config, logCfg LoggerConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.IsDevelopment() || logCfg.Development {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Set log level
	if logCfg.Level != "" {
		level, err := zapcore.ParseLevel(logCfg.Level)
		if err != nil {
			return nil, err
		}
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	}

	// Set encoding
	if logCfg.Encoding != "" {
		zapConfig.Encoding = logCfg.Encoding
	}

	// Add caller information in development
	if cfg.IsDevelopment() {
		zapConfig.Development = true
		zapConfig.DisableStacktrace = false
	} else {
		zapConfig.DisableStacktrace = true
	}

	return zapConfig.Build()
}

// WithRequestContext adds request context fields to a logger
func WithRequestContext(logger *zap.Logger, requestID, method, path string) *zap.Logger {
	return logger.With(
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
	)
}

// WithUserContext adds user context fields to a logger
func WithUserContext(logger *zap.Logger, userID interface{}, email string) *zap.Logger {
	return logger.With(
		zap.Any("user_id", userID),
		zap.String("user_email", email),
	)
}

// GetZapLogger extracts the underlying zap.Logger from our Logger interface
// This is used for compatibility with existing functions that require *zap.Logger
func GetZapLogger(logger Logger) *zap.Logger {
	if zapLog, ok := logger.(*zapLogger); ok {
		return zapLog.sugar.Desugar()
	}
	// Fallback - create a new zap logger
	zapLog, _ := zap.NewProduction()
	return zapLog
}
