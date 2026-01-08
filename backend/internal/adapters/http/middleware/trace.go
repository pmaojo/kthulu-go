package middleware

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDMiddleware stores the current trace ID in the request context.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
