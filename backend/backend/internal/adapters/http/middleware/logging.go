// @kthulu:core
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/pmaojo/kthulu-go/backend/internal/observability"
)

var randRead = rand.Read

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// LoggerKey is the context key for logger
	LoggerKey ContextKey = "logger"
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
)

// LoggingMiddleware creates a middleware that logs HTTP requests with correlation IDs
func LoggingMiddleware(logger observability.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get or generate request ID
			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = generateRequestID(logger)
			}

			// Create logger with request ID
			requestLogger := logger.With(
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			// Add request ID and logger to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			ctx = context.WithValue(ctx, LoggerKey, requestLogger)

			// Include trace_id in logger if present in context
			if traceID := GetTraceID(r.Context()); traceID != "" {
				requestLogger = requestLogger.With(zap.String("trace_id", traceID))
				ctx = context.WithValue(ctx, TraceIDKey, traceID)
			}

			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Log request start
			requestLogger.Info("Request started")

			// Process request
			next.ServeHTTP(ww, r)

			// Log request completion
			duration := time.Since(start)
			requestLogger.Info("Request completed",
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", duration),
			)
		})
	}
}

// GetRequestID extracts the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetLogger extracts the logger from context
func GetLogger(ctx context.Context) observability.Logger {
	if logger, ok := ctx.Value(LoggerKey).(observability.Logger); ok {
		return logger
	}
	return nil
}

// GetSugaredLogger extracts a sugared logger from context
func GetSugaredLogger(ctx context.Context) *zap.SugaredLogger {
	l := GetLogger(ctx)
	if l == nil {
		return zap.NewNop().Sugar()
	}
	return observability.GetZapLogger(l).Sugar()
}

// GetTraceID extracts the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// generateRequestID creates a random request ID
// If random source fails, it falls back to a UUID and logs the error
func generateRequestID(logger observability.Logger) string {
	bytes := make([]byte, 8)
	if _, err := randRead(bytes); err != nil {
		logger.Error("failed to generate random request ID", zap.Error(err))
		return uuid.NewString()
	}
	return hex.EncodeToString(bytes)
}
