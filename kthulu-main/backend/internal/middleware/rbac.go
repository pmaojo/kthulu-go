package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"backend/cmd/kthulu-cli/internal/security"
)

// RBACMiddleware provides enterprise role-based access control
func RBACMiddleware(rbacEngine *security.RBACEngine) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract user information from context/headers
		userID := extractUserID(c)
		userRoles := extractUserRoles(c)
		
		if userID == "" {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Create access request
		request := &security.AccessRequest{
			Subject:   userID,
			Resource:  normalizeResource(c.Request.URL.Path),
			Action:    strings.ToLower(c.Request.Method),
			UserRoles: userRoles,
			Context: map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"query":      c.Request.URL.RawQuery,
			},
			Timestamp: time.Now(),
			RequestID: generateRequestID(),
		}

		// Check access
		result, err := rbacEngine.CheckAccess(c.Request.Context(), request)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Authorization check failed",
				"trace": err.Error(),
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(403, gin.H{
				"error":           "Access denied",
				"reason":          result.Reason,
				"applied_policies": result.AppliedPolicies,
				"request_id":      request.RequestID,
			})
			c.Abort()
			return
		}

		// Add authorization info to context for downstream handlers
		c.Set("access_result", result)
		c.Set("user_id", userID)
		c.Set("user_roles", userRoles)
		c.Set("request_id", request.RequestID)
		
		c.Next()
	})
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})
}

// AuditMiddleware logs security events
func AuditMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		// Log security-relevant events
		if c.Writer.Status() == 403 || c.Writer.Status() == 401 {
			logSecurityEvent(c, start)
		}
	})
}

// Helper functions
func extractUserID(c *gin.Context) string {
	// Try various sources
	if userID := c.GetString("user_id"); userID != "" {
		return userID
	}
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	// Extract from JWT or session
	return ""
}

func extractUserRoles(c *gin.Context) []string {
	if roles := c.GetStringSlice("user_roles"); len(roles) > 0 {
		return roles
	}
	if rolesHeader := c.GetHeader("X-User-Roles"); rolesHeader != "" {
		return strings.Split(rolesHeader, ",")
	}
	return []string{"guest"}
}

func normalizeResource(path string) string {
	// Convert REST paths to resource names
	// /api/v1/users/123 -> users
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[2] // Return resource name
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "default"
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func logSecurityEvent(c *gin.Context, start time.Time) {
	// Implement security event logging
	log.WithFields(log.Fields{
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"status":      c.Writer.Status(),
		"ip":          c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"duration":    time.Since(start),
		"user_id":     c.GetString("user_id"),
		"request_id":  c.GetString("request_id"),
	}).Warn("Security event")
}
