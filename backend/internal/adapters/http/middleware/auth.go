// @kthulu:module:auth
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	auth "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/auth"
)

// UserIDKey is the context key for user ID
const UserIDKey ContextKey = "user_id"

// AuthMiddleware creates a middleware that validates JWT tokens and extracts user information
func AuthMiddleware(tokenManager core.TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			logger := GetSugaredLogger(r.Context())

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check for Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Warn("Invalid Authorization header format")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]

			// Validate the access token
			claims, err := tokenManager.ValidateAccessToken(tokenStr)
			if err != nil {
				logger.Warnw("Invalid access token", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate DPoP proof binding if present
			if err := auth.ValidateDPoP(r, tokenStr); err != nil {
				logger.Warnw("Invalid DPoP proof", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims
			userIDFloat, ok := claims["sub"].(float64)
			if !ok {
				logger.Warn("Invalid user ID in token claims")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userID := uint(userIDFloat)

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)

			logger.Infow("User authenticated", "userId", userID)

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts the user ID from context
func GetUserID(ctx context.Context) (uint, error) {
	if userID, ok := ctx.Value(UserIDKey).(uint); ok {
		return userID, nil
	}
	return 0, domain.ErrUserNotFound
}

// RequireAuth is a convenience function that creates an auth middleware
func RequireAuth(tokenManager core.TokenManager) func(next http.Handler) http.Handler {
	return AuthMiddleware(tokenManager)
}
