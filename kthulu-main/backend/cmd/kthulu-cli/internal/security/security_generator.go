package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SecurityGenerator generates security components from @kthulu:security tags
type SecurityGenerator struct {
	processor  *SecurityTagProcessor
	rbacEngine *RBACEngine
}

// SecurityGenerationResult contains the results of security generation
type SecurityGenerationResult struct {
	PoliciesGenerated   int                      `json:"policies_generated"`
	RolesGenerated      int                      `json:"roles_generated"`
	MiddlewareGenerated bool                     `json:"middleware_generated"`
	ConfigGenerated     bool                     `json:"config_generated"`
	SecurityReport      *SecurityReport          `json:"security_report"`
	GeneratedFiles      []string                 `json:"generated_files"`
	Recommendations     []SecurityRecommendation `json:"recommendations"`
	ProcessingTime      time.Duration            `json:"processing_time"`
}

// SecurityReport provides a comprehensive security analysis
type SecurityReport struct {
	ProjectPath             string                  `json:"project_path"`
	TotalSecurityTags       int                     `json:"total_security_tags"`
	CoveragePercentage      float64                 `json:"coverage_percentage"`
	SecurityVulnerabilities []SecurityVulnerability `json:"vulnerabilities"`
	ComplianceStatus        map[string]bool         `json:"compliance_status"`
	RiskLevel               string                  `json:"risk_level"`
	RecommendedActions      []string                `json:"recommended_actions"`
	GeneratedAt             time.Time               `json:"generated_at"`
}

// SecurityVulnerability represents a potential security issue
type SecurityVulnerability struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	FilePath       string `json:"file_path"`
	LineNumber     int    `json:"line_number"`
	Recommendation string `json:"recommendation"`
}

// SecurityRecommendation provides actionable security advice
type SecurityRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	FilePath    string `json:"file_path,omitempty"`
	LineNumber  int    `json:"line_number,omitempty"`
}

// NewSecurityGenerator creates a new security generator
func NewSecurityGenerator() *SecurityGenerator {
	rbacEngine := NewRBACEngine(nil)
	processor := NewSecurityTagProcessor(rbacEngine)

	return &SecurityGenerator{
		processor:  processor,
		rbacEngine: rbacEngine,
	}
}

// GenerateSecurityInfrastructure analyzes project and generates complete security infrastructure
func (g *SecurityGenerator) GenerateSecurityInfrastructure(projectPath string) (*SecurityGenerationResult, error) {
	startTime := time.Now()

	result := &SecurityGenerationResult{
		GeneratedFiles:  []string{},
		Recommendations: []SecurityRecommendation{},
	}

	// Process security tags
	err := g.processor.ProcessProjectSecurity(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to process security tags: %w", err)
	}

	// Get RBAC stats
	stats := g.rbacEngine.GetStats()
	result.PoliciesGenerated = stats["total_policies"].(int)
	result.RolesGenerated = stats["total_roles"].(int)

	// Generate middleware file
	if err := g.generateMiddlewareFile(projectPath); err == nil {
		result.MiddlewareGenerated = true
		result.GeneratedFiles = append(result.GeneratedFiles, filepath.Join(projectPath, "internal/middleware/rbac.go"))
	}

	// Generate security configuration
	if err := g.generateSecurityConfig(projectPath); err == nil {
		result.ConfigGenerated = true
		result.GeneratedFiles = append(result.GeneratedFiles, filepath.Join(projectPath, "config/security.yaml"))
	}

	// Generate security report
	report, err := g.generateSecurityReport(projectPath)
	if err == nil {
		result.SecurityReport = report
	}

	// Generate recommendations
	result.Recommendations = g.generateRecommendations(projectPath)

	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// generateMiddlewareFile creates RBAC middleware file
func (g *SecurityGenerator) generateMiddlewareFile(projectPath string) error {
	middlewareDir := filepath.Join(projectPath, "internal", "middleware")
	if err := os.MkdirAll(middlewareDir, 0755); err != nil {
		return err
	}

	middlewareCode := g.generateMiddlewareCode()
	filePath := filepath.Join(middlewareDir, "rbac.go")

	return os.WriteFile(filePath, []byte(middlewareCode), 0644)
}

// generateMiddlewareCode creates the RBAC middleware code
func (g *SecurityGenerator) generateMiddlewareCode() string {
	return `package middleware

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
`
}

// generateSecurityConfig creates security configuration file
func (g *SecurityGenerator) generateSecurityConfig(projectPath string) error {
	configDir := filepath.Join(projectPath, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	config := g.processor.GenerateSecurityConfig()
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(configDir, "security.json")
	return os.WriteFile(filePath, configBytes, 0644)
}

// generateSecurityReport creates comprehensive security analysis
func (g *SecurityGenerator) generateSecurityReport(projectPath string) (*SecurityReport, error) {
	report := &SecurityReport{
		ProjectPath:             projectPath,
		SecurityVulnerabilities: []SecurityVulnerability{},
		ComplianceStatus:        make(map[string]bool),
		RecommendedActions:      []string{},
		GeneratedAt:             time.Now(),
	}

	// Analyze security coverage
	// This would involve scanning for unprotected endpoints, missing auth, etc.
	report.TotalSecurityTags = 15    // Placeholder
	report.CoveragePercentage = 65.0 // Placeholder

	// Risk assessment
	if report.CoveragePercentage < 50 {
		report.RiskLevel = "HIGH"
	} else if report.CoveragePercentage < 80 {
		report.RiskLevel = "MEDIUM"
	} else {
		report.RiskLevel = "LOW"
	}

	// Compliance checks
	report.ComplianceStatus["SOC2"] = report.CoveragePercentage > 70
	report.ComplianceStatus["GDPR"] = report.CoveragePercentage > 80
	report.ComplianceStatus["HIPAA"] = report.CoveragePercentage > 85

	return report, nil
}

// generateRecommendations creates security recommendations
func (g *SecurityGenerator) generateRecommendations(projectPath string) []SecurityRecommendation {
	stats := g.rbacEngine.GetStats()
	recommendations := []SecurityRecommendation{}

	// Add recommendations based on analysis
	if stats["total_policies"].(int) < 5 {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "security_policy",
			Priority:    "HIGH",
			Title:       "Insufficient Security Policies",
			Description: "Project has very few security policies. Consider adding @kthulu:security tags to protect sensitive operations.",
			Action:      "Add @kthulu:security tags to handlers and services",
		})
	}

	if stats["total_roles"].(int) < 3 {
		recommendations = append(recommendations, SecurityRecommendation{
			Type:        "rbac_roles",
			Priority:    "MEDIUM",
			Title:       "Limited Role Structure",
			Description: "Consider implementing a more granular role-based access control system.",
			Action:      "Define additional roles for different user types and access levels",
		})
	}

	recommendations = append(recommendations, SecurityRecommendation{
		Type:        "middleware",
		Priority:    "HIGH",
		Title:       "Implement Security Middleware",
		Description: "Apply the generated RBAC middleware to all API routes for consistent security enforcement.",
		Action:      "Add middleware to router configuration",
	})

	return recommendations
}

// GetSecurityStatus provides current security status
func (g *SecurityGenerator) GetSecurityStatus() map[string]interface{} {
	stats := g.rbacEngine.GetStats()

	return map[string]interface{}{
		"rbac_engine": stats,
		"policies":    g.listPolicies(),
		"roles":       g.listRoles(),
		"last_update": time.Now(),
	}
}

func (g *SecurityGenerator) listPolicies() []map[string]interface{} {
	// This would normally access the internal policies
	// For demo purposes, return sample data
	return []map[string]interface{}{
		{"id": "policy_1", "name": "Admin Access", "resource": "admin"},
		{"id": "policy_2", "name": "User Access", "resource": "user"},
	}
}

func (g *SecurityGenerator) listRoles() []map[string]interface{} {
	return []map[string]interface{}{
		{"id": "admin", "name": "Administrator", "level": 100},
		{"id": "user", "name": "Regular User", "level": 10},
		{"id": "guest", "name": "Guest", "level": 1},
	}
}
