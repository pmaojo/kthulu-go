package middleware

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"backend/core"
)

// JWTTraceMiddleware extracts the jti from the JWT and attaches it to the current span.
func JWTTraceMiddleware(tokenManager core.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenStr := parts[1]
					if claims, err := tokenManager.ValidateAccessToken(tokenStr); err == nil {
						if jti, ok := claims["jti"].(string); ok {
							span := trace.SpanFromContext(r.Context())
							if span.SpanContext().IsValid() {
								span.SetAttributes(attribute.String("jwt.jti", jti))
							}
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
