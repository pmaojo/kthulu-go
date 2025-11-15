package security

import (
	"fmt"
	"go/token"
	"strings"
	"time"

	kthuluParser "github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
)

// SecurityTagProcessor processes @kthulu:security tags and generates RBAC policies
type SecurityTagProcessor struct {
	rbacEngine *RBACEngine
	fileSet    *token.FileSet
}

// SecurityTagInfo represents security information extracted from tags
type SecurityTagInfo struct {
	FilePath    string                 `json:"file_path"`
	LineNumber  int                    `json:"line_number"`
	TagType     string                 `json:"tag_type"`
	Resource    string                 `json:"resource"`
	Actions     []string               `json:"actions"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	Conditions  map[string]interface{} `json:"conditions"`
	Attributes  map[string]string      `json:"attributes"`
}

// NewSecurityTagProcessor creates a new security tag processor
func NewSecurityTagProcessor(rbacEngine *RBACEngine) *SecurityTagProcessor {
	return &SecurityTagProcessor{
		rbacEngine: rbacEngine,
		fileSet:    token.NewFileSet(),
	}
}

// ProcessProjectSecurity analyzes a project and generates security policies from @kthulu:security tags
func (p *SecurityTagProcessor) ProcessProjectSecurity(projectPath string) error {
	// Use the advanced parser to get all tags
	integration := kthuluParser.NewAdvancedIntegration()
	analysis, _, _, err := integration.AnalyzeProjectWithInsights(projectPath)
	if err != nil {
		return fmt.Errorf("failed to analyze project: %w", err)
	}

	// Process each tag to extract security information
	securityTags := p.extractSecurityTags(analysis.Tags)

	// Generate security policies from extracted tags
	policies := p.generateSecurityPolicies(securityTags)

	// Add policies to RBAC engine
	for _, policy := range policies {
		p.rbacEngine.AddPolicy(policy)
	}

	// Generate default roles based on discovered security patterns
	roles := p.generateDefaultRoles(securityTags)
	for _, role := range roles {
		p.rbacEngine.AddRole(role)
	}

	return nil
}

// extractSecurityTags filters and processes security-related tags
func (p *SecurityTagProcessor) extractSecurityTags(tags []kthuluParser.Tag) []*SecurityTagInfo {
	var securityTags []*SecurityTagInfo

	for _, tag := range tags {
		// Process @kthulu:security tags
		if strings.HasPrefix(string(tag.Type), "security") ||
			strings.Contains(tag.Content, "security") {

			securityInfo := &SecurityTagInfo{
				LineNumber: tag.Line,
				TagType:    string(tag.Type),
				Attributes: tag.Attributes,
				Conditions: make(map[string]interface{}),
			}

			// Parse security-specific attributes
			p.parseSecurityAttributes(securityInfo, tag)

			securityTags = append(securityTags, securityInfo)
		}

		// Also process other security-relevant tags
		if p.isSecurityRelevant(tag) {
			securityInfo := p.extractImplicitSecurity(tag)
			if securityInfo != nil {
				securityTags = append(securityTags, securityInfo)
			}
		}
	}

	return securityTags
}

// parseSecurityAttributes parses security-specific tag attributes
func (p *SecurityTagProcessor) parseSecurityAttributes(info *SecurityTagInfo, tag kthuluParser.Tag) {
	for key, value := range tag.Attributes {
		switch key {
		case "role", "roles":
			info.Roles = p.parseStringList(value)
		case "permission", "permissions":
			info.Permissions = p.parseStringList(value)
		case "action", "actions":
			info.Actions = p.parseStringList(value)
		case "resource":
			info.Resource = value
		case "condition":
			// Parse condition expressions
			conditions := p.parseConditions(value)
			for k, v := range conditions {
				info.Conditions[k] = v
			}
		case "level":
			info.Conditions["security_level"] = value
		case "scope":
			info.Conditions["scope"] = value
		}
	}

	// Extract resource from tag value if not explicitly set
	if info.Resource == "" {
		info.Resource = p.inferResource(tag)
	}
}

// isSecurityRelevant checks if a tag has security implications
func (p *SecurityTagProcessor) isSecurityRelevant(tag kthuluParser.Tag) bool {
	securityKeywords := []string{
		"admin", "auth", "user", "permission", "role",
		"private", "protected", "secure", "sensitive",
		"token", "session", "credential", "secret",
	}

	content := strings.ToLower(tag.Content + " " + tag.Value)
	for _, keyword := range securityKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	return false
}

// extractImplicitSecurity extracts security information from non-security tags
func (p *SecurityTagProcessor) extractImplicitSecurity(tag kthuluParser.Tag) *SecurityTagInfo {
	content := strings.ToLower(tag.Content + " " + tag.Value)

	info := &SecurityTagInfo{
		LineNumber: tag.Line,
		TagType:    "implicit_security",
		Attributes: make(map[string]string),
		Conditions: make(map[string]interface{}),
	}

	// Infer security requirements based on content
	if strings.Contains(content, "admin") {
		info.Roles = []string{"admin"}
		info.Resource = p.inferResource(tag)
		info.Actions = []string{"read", "write", "delete"}
	} else if strings.Contains(content, "user") {
		info.Roles = []string{"user"}
		info.Resource = p.inferResource(tag)
		info.Actions = []string{"read"}
	} else if strings.Contains(content, "auth") {
		info.Roles = []string{"authenticated"}
		info.Resource = p.inferResource(tag)
		info.Actions = []string{"read"}
	}

	// Only return if we found meaningful security info
	if len(info.Roles) > 0 || len(info.Permissions) > 0 {
		return info
	}

	return nil
}

// generateSecurityPolicies creates security policies from extracted tag information
func (p *SecurityTagProcessor) generateSecurityPolicies(securityTags []*SecurityTagInfo) []*SecurityPolicy {
	var policies []*SecurityPolicy
	policyID := 1

	for _, tagInfo := range securityTags {
		// Skip if no actionable security information
		if len(tagInfo.Roles) == 0 && len(tagInfo.Permissions) == 0 {
			continue
		}

		policy := &SecurityPolicy{
			ID:            fmt.Sprintf("policy_%d", policyID),
			Name:          fmt.Sprintf("Policy for %s", tagInfo.Resource),
			Description:   fmt.Sprintf("Auto-generated policy from security tag at line %d", tagInfo.LineNumber),
			Resource:      tagInfo.Resource,
			Actions:       tagInfo.Actions,
			RequiredRoles: tagInfo.Roles,
			Conditions:    tagInfo.Conditions,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Set default actions if none specified
		if len(policy.Actions) == 0 {
			policy.Actions = []string{"read"}
		}

		// Set default resource if none specified
		if policy.Resource == "" {
			policy.Resource = "default"
		}

		policies = append(policies, policy)
		policyID++
	}

	return policies
}

// generateDefaultRoles creates default roles based on discovered patterns
func (p *SecurityTagProcessor) generateDefaultRoles(securityTags []*SecurityTagInfo) []*Role {
	roleMap := make(map[string]*Role)

	// Standard enterprise roles
	standardRoles := map[string]*Role{
		"admin": {
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system access",
			Permissions: []string{"*"},
			Level:       100,
		},
		"user": {
			ID:          "user",
			Name:        "Regular User",
			Description: "Standard user access",
			Permissions: []string{"read", "write:own"},
			ParentRoles: []string{},
			Level:       10,
		},
		"authenticated": {
			ID:          "authenticated",
			Name:        "Authenticated User",
			Description: "Basic authenticated access",
			Permissions: []string{"read:public"},
			Level:       5,
		},
		"guest": {
			ID:          "guest",
			Name:        "Guest User",
			Description: "Anonymous access",
			Permissions: []string{"read:public:limited"},
			Level:       1,
		},
	}

	// Start with standard roles
	for id, role := range standardRoles {
		role.CreatedAt = time.Now()
		roleMap[id] = role
	}

	// Extract additional roles from tags
	for _, tagInfo := range securityTags {
		for _, roleName := range tagInfo.Roles {
			if _, exists := roleMap[roleName]; !exists {
				role := &Role{
					ID:          roleName,
					Name:        strings.Title(roleName),
					Description: fmt.Sprintf("Auto-generated role: %s", roleName),
					Permissions: []string{"read"},
					Level:       20, // Default level
					CreatedAt:   time.Now(),
				}
				roleMap[roleName] = role
			}
		}
	}

	// Convert map to slice
	var roles []*Role
	for _, role := range roleMap {
		roles = append(roles, role)
	}

	return roles
}

// Helper functions
func (p *SecurityTagProcessor) parseStringList(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (p *SecurityTagProcessor) parseConditions(value string) map[string]interface{} {
	conditions := make(map[string]interface{})

	// Simple condition parsing: key=value,key2=value2
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		if kv := strings.Split(pair, "="); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			conditions[key] = val
		}
	}

	return conditions
}

func (p *SecurityTagProcessor) inferResource(tag kthuluParser.Tag) string {
	// Try to infer resource from tag value or type
	if tag.Value != "" {
		return tag.Value
	}

	if string(tag.Type) != "" {
		return string(tag.Type)
	}

	return "default"
}

// Enterprise integration methods

// GenerateMiddleware creates HTTP middleware for RBAC enforcement
func (p *SecurityTagProcessor) GenerateMiddleware() string {
	return `
// Auto-generated RBAC middleware
func RBACMiddleware(rbacEngine *security.RBACEngine) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Extract user information from context
		userID := c.GetString("user_id")
		userRoles := c.GetStringSlice("user_roles")
		
		if userID == "" {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Create access request
		request := &security.AccessRequest{
			Subject:   userID,
			Resource:  c.Request.URL.Path,
			Action:    strings.ToLower(c.Request.Method),
			UserRoles: userRoles,
			Context: map[string]interface{}{
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"method":     c.Request.Method,
			},
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		}

		// Check access
		result, err := rbacEngine.CheckAccess(c.Request.Context(), request)
		if err != nil {
			c.JSON(500, gin.H{"error": "Authorization check failed"})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(403, gin.H{
				"error":  "Access denied",
				"reason": result.Reason,
			})
			c.Abort()
			return
		}

		// Add authorization info to context
		c.Set("access_result", result)
		c.Next()
	})
}
`
}

// GenerateSecurityConfig creates configuration for security settings
func (p *SecurityTagProcessor) GenerateSecurityConfig() map[string]interface{} {
	return map[string]interface{}{
		"rbac": map[string]interface{}{
			"enabled":             true,
			"cache_enabled":       true,
			"cache_ttl":           "5m",
			"audit_enabled":       true,
			"strict_mode":         true,
			"default_deny_policy": true,
			"hierarchical_roles":  true,
			"contextual_security": true,
		},
		"audit": map[string]interface{}{
			"enabled":        true,
			"log_level":      "INFO",
			"retention_days": 90,
			"storage_type":   "database",
		},
		"session": map[string]interface{}{
			"timeout":       "30m",
			"secure_cookie": true,
			"same_site":     "strict",
		},
	}
}
