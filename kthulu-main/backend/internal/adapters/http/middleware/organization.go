package middleware

import (
	"context"
	"net/http"
	"strconv"
)

// OrganizationIDKey is the context key for organization ID
const OrganizationIDKey ContextKey = "organization_id"

// OrganizationContextMiddleware extracts organization ID from header and stores it in context
func OrganizationContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if orgIDStr := r.Header.Get("X-Organization-ID"); orgIDStr != "" {
			if orgID, err := strconv.ParseUint(orgIDStr, 10, 32); err == nil {
				ctx := context.WithValue(r.Context(), OrganizationIDKey, uint(orgID))
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
