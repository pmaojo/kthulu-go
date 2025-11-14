// @kthulu:security:resource=admin,roles=admin,actions=read,write,delete
// @kthulu:security:level=high,audit=true
package admin

import (
	"context"

	"github.com/gin-gonic/gin"
)

// @kthulu:security:resource=users,roles=admin,manager,actions=read,write
// AdminHandler provides administrative operations
type AdminHandler struct {
	userService UserService
}

// @kthulu:security:resource=users,roles=admin,actions=delete,condition=user.status!=active
// DeleteUser deletes a user (requires admin role and user must not be active)
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Admin-only operation with conditional security
	if err := h.userService.DeleteUser(context.Background(), userID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "User deleted successfully"})
}

// @kthulu:security:resource=system,roles=admin,actions=read
// GetSystemStats returns system statistics (admin only)
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.getSystemStats()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}

// @kthulu:security:resource=audit,roles=admin,auditor,actions=read
// GetAuditLogs returns audit logs (admin and auditor roles)
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	logs, err := h.getAuditLogs()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, logs)
}

func (h *AdminHandler) getSystemStats() (map[string]interface{}, error) {
	return map[string]interface{}{
		"users_count":     1250,
		"active_sessions": 89,
		"system_health":   "healthy",
	}, nil
}

func (h *AdminHandler) getAuditLogs() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"timestamp": "2025-11-14T10:00:00Z", "action": "user.login", "user_id": "user123"},
		{"timestamp": "2025-11-14T10:05:00Z", "action": "user.logout", "user_id": "user123"},
	}, nil
}

type UserService interface {
	DeleteUser(ctx context.Context, userID string) error
}
