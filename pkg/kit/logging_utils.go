// @kthulu:core
package core

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerFields is a type alias for structured logging fields
type LoggerFields map[string]interface{}

// LogWithFields logs a message with structured fields
func LogWithFields(logger *zap.Logger, level zapcore.Level, msg string, fields LoggerFields) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	logger.Log(level, msg, zapFields...)
}

// LogError logs an error with context
func LogError(ctx context.Context, logger *zap.Logger, msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	logger.Error(msg, allFields...)
}

// LogInfo logs an info message with context
func LogInfo(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// LogDebug logs a debug message with context
func LogDebug(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// LogWarn logs a warning message with context
func LogWarn(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// TimedOperation logs the duration of an operation
func TimedOperation(logger *zap.Logger, operationName string, fn func() error) error {
	start := time.Now()
	logger.Info("Starting operation", zap.String("operation", operationName))

	err := fn()
	duration := time.Since(start)

	if err != nil {
		logger.Error("Operation failed",
			zap.String("operation", operationName),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		logger.Info("Operation completed",
			zap.String("operation", operationName),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// DatabaseOperationLogger creates a logger specifically for database operations
func DatabaseOperationLogger(logger *zap.Logger, operation, table string) *zap.Logger {
	return logger.With(
		zap.String("component", "database"),
		zap.String("operation", operation),
		zap.String("table", table),
	)
}

// BusinessLogicLogger creates a logger specifically for business logic
func BusinessLogicLogger(logger *zap.Logger, useCase, operation string) *zap.Logger {
	return logger.With(
		zap.String("component", "business_logic"),
		zap.String("use_case", useCase),
		zap.String("operation", operation),
	)
}

// HTTPLogger creates a logger specifically for HTTP operations
func HTTPLogger(logger *zap.Logger, method, path string) *zap.Logger {
	return logger.With(
		zap.String("component", "http"),
		zap.String("method", method),
		zap.String("path", path),
	)
}
