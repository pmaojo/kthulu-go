// @kthulu:module:access
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// RoleKey is the context key for user role
const RoleKey ContextKey = "user_role"

// RBACMiddleware creates a middleware that enforces role-based access control
func RBACMiddleware(roleRepo repository.RoleRepository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			logger := GetSugaredLogger(r.Context())

			// Get user ID from context (should be set by auth middleware)
			userID, err := GetUserID(r.Context())
			if err != nil {
				logger.Warn("RBAC middleware called without authenticated user")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Load user's role and store in context
			role, err := roleRepo.FindByUserID(r.Context(), userID)
			if err != nil {
				logger.Errorw("Failed to load user role", "userId", userID, "error", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), RoleKey, role)
			logger.Infow("RBAC middleware passed", "userId", userID, "role", role.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roleRepo repository.RoleRepository, requiredRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			logger := GetSugaredLogger(r.Context())

			// Get user ID from context (should be set by auth middleware)
			userID, err := GetUserID(r.Context())
			if err != nil {
				logger.Warn("RequireRole middleware called without authenticated user")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Load the user's role from repository
			role, err := roleRepo.FindByUserID(r.Context(), userID)
			if err != nil {
				logger.Errorw("Failed to load user role", "userId", userID, "error", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Compare with required role
			if !strings.EqualFold(role.Name, requiredRole) {
				logger.Warnw("Role requirement not met", "userId", userID, "requiredRole", requiredRole, "userRole", role.Name)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Store role in context for downstream handlers
			ctx := context.WithValue(r.Context(), RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission creates a middleware that requires a specific permission
func RequirePermission(roleRepo repository.RoleRepository, resource, action string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			logger := GetSugaredLogger(r.Context())

			// Get user ID from context (should be set by auth middleware)
			userID, err := GetUserID(r.Context())
			if err != nil {
				logger.Warn("RequirePermission middleware called without authenticated user")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Load the user's role with permissions
			role, err := roleRepo.FindByUserID(r.Context(), userID)
			if err != nil {
				logger.Errorw("Failed to load user role for permission check", "userId", userID, "error", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Verify permission
			if !role.HasPermission(resource, action) {
				logger.Warnw("Permission requirement not met", "userId", userID, "role", role.Name, "resource", resource, "action", action)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Store role in context for downstream handlers
			ctx := context.WithValue(r.Context(), RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RoleScopeMiddleware creates a middleware that validates X-Role-Scope header
func RoleScopeMiddleware(roleRepo repository.RoleRepository) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get logger from context
			logger := GetSugaredLogger(r.Context())

			// Get user ID from context (should be set by auth middleware)
			userID, err := GetUserID(r.Context())
			if err != nil {
				logger.Warn("RoleScope middleware called without authenticated user")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check for X-Role-Scope header
			roleScope := r.Header.Get("X-Role-Scope")
			if roleScope == "" {
				logger.Infow("No role scope specified", "userId", userID)
				next.ServeHTTP(w, r)
				return
			}

			// Parse role scope (format: "role:resource:action" or "role")
			parts := strings.Split(roleScope, ":")
			if len(parts) != 1 && len(parts) != 3 {
				logger.Warnw("Invalid role scope format", "userId", userID, "roleScope", roleScope)
				http.Error(w, "Invalid role scope format", http.StatusBadRequest)
				return
			}

			requiredRole := parts[0]

			// Load user's role
			role, err := roleRepo.FindByUserID(r.Context(), userID)
			if err != nil {
				logger.Errorw("Failed to load user role for scope validation", "userId", userID, "error", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Verify role name matches required role
			if !strings.EqualFold(role.Name, requiredRole) {
				logger.Warnw("Role scope mismatch", "userId", userID, "expectedRole", requiredRole, "userRole", role.Name)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// If resource and action specified, verify permission
			if len(parts) == 3 {
				resource := parts[1]
				action := parts[2]
				if !role.HasPermission(resource, action) {
					logger.Warnw("Role scope permission denied", "userId", userID, "role", role.Name, "resource", resource, "action", action)
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			// Store role in context for downstream handlers
			ctx := context.WithValue(r.Context(), RoleKey, role)
			logger.Infow("Role scope validated", "userId", userID, "roleScope", roleScope)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserRole extracts the user role from context
func GetUserRole(ctx context.Context) (*domain.Role, error) {
	if role, ok := ctx.Value(RoleKey).(*domain.Role); ok {
		return role, nil
	}
	return nil, domain.ErrRoleNotFound
}
