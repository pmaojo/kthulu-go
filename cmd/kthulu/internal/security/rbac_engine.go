package security

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// RBACEngine provides enterprise role-based access control
type RBACEngine struct {
	policies    map[string]*SecurityPolicy
	roles       map[string]*Role
	permissions map[string]*Permission
	cache       *RBACCache
	mutex       sync.RWMutex
	config      *RBACConfig
}

// RBACConfig configures the RBAC engine
type RBACConfig struct {
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	AuditEnabled       bool          `json:"audit_enabled"`
	StrictMode         bool          `json:"strict_mode"`
	DefaultDenyPolicy  bool          `json:"default_deny_policy"`
	HierarchicalRoles  bool          `json:"hierarchical_roles"`
	ContextualSecurity bool          `json:"contextual_security"`
}

// SecurityPolicy represents a security policy derived from @kthulu:security tags
type SecurityPolicy struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Module        string                 `json:"module"`
	Resource      string                 `json:"resource"`
	Actions       []string               `json:"actions"`
	Conditions    map[string]interface{} `json:"conditions"`
	RequiredRoles []string               `json:"required_roles"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Role represents a security role
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []string     `json:"permissions"`
	ParentRoles []string     `json:"parent_roles"`
	Level       int          `json:"level"`
	Metadata    RoleMetadata `json:"metadata"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Permission represents a fine-grained permission
type Permission struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Scope       string                 `json:"scope"`
	Conditions  map[string]interface{} `json:"conditions"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
}

// RoleMetadata contains additional role information
type RoleMetadata struct {
	Department    string            `json:"department"`
	BusinessUnit  string            `json:"business_unit"`
	SecurityLevel string            `json:"security_level"`
	Tags          []string          `json:"tags"`
	Attributes    map[string]string `json:"attributes"`
}

// AccessRequest represents an authorization request
type AccessRequest struct {
	Subject   string                 `json:"subject"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Context   map[string]interface{} `json:"context"`
	UserRoles []string               `json:"user_roles"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
}

// AccessResult represents the result of an authorization check
type AccessResult struct {
	Allowed         bool                   `json:"allowed"`
	Reason          string                 `json:"reason"`
	AppliedPolicies []string               `json:"applied_policies"`
	Conditions      map[string]interface{} `json:"conditions"`
	AuditLog        *AuditEntry            `json:"audit_log,omitempty"`
	CacheHit        bool                   `json:"cache_hit"`
	ProcessingTime  time.Duration          `json:"processing_time"`
}

// AuditEntry represents a security audit log entry
type AuditEntry struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Subject    string                 `json:"subject"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Result     bool                   `json:"result"`
	Reason     string                 `json:"reason"`
	Context    map[string]interface{} `json:"context"`
	PolicyID   string                 `json:"policy_id"`
	SessionID  string                 `json:"session_id"`
	RemoteAddr string                 `json:"remote_addr"`
	UserAgent  string                 `json:"user_agent"`
}

// RBACCache provides caching for authorization decisions
type RBACCache struct {
	decisions map[string]*CachedDecision
	mutex     sync.RWMutex
	ttl       time.Duration
	maxSize   int
}

// CachedDecision represents a cached authorization decision
type CachedDecision struct {
	Result    *AccessResult `json:"result"`
	ExpiresAt time.Time     `json:"expires_at"`
	HitCount  int           `json:"hit_count"`
	CreatedAt time.Time     `json:"created_at"`
}

// NewRBACEngine creates a new RBAC engine
func NewRBACEngine(config *RBACConfig) *RBACEngine {
	if config == nil {
		config = &RBACConfig{
			CacheEnabled:       true,
			CacheTTL:           300 * time.Second, // 5 minutes
			AuditEnabled:       true,
			StrictMode:         true,
			DefaultDenyPolicy:  true,
			HierarchicalRoles:  true,
			ContextualSecurity: true,
		}
	}

	cache := &RBACCache{
		decisions: make(map[string]*CachedDecision),
		ttl:       config.CacheTTL,
		maxSize:   1000, // Default cache size
	}

	return &RBACEngine{
		policies:    make(map[string]*SecurityPolicy),
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
		cache:       cache,
		config:      config,
	}
}

// CheckAccess performs authorization check
func (e *RBACEngine) CheckAccess(ctx context.Context, request *AccessRequest) (*AccessResult, error) {
	startTime := time.Now()

	// Generate cache key
	cacheKey := e.generateCacheKey(request)

	// Check cache first
	if e.config.CacheEnabled {
		if cached := e.cache.Get(cacheKey); cached != nil && !cached.IsExpired() {
			cached.HitCount++
			result := cached.Result
			result.CacheHit = true
			result.ProcessingTime = time.Since(startTime)
			return result, nil
		}
	}

	// Perform authorization logic
	result := &AccessResult{
		Allowed:         false,
		AppliedPolicies: []string{},
		Conditions:      make(map[string]interface{}),
		CacheHit:        false,
		ProcessingTime:  0,
	}

	// Apply default deny policy
	if e.config.DefaultDenyPolicy {
		result.Reason = "Default deny policy - no explicit allow found"
	}

	// Check user roles and permissions
	allowed, reason, policies := e.evaluateAccess(request)
	result.Allowed = allowed
	result.Reason = reason
	result.AppliedPolicies = policies

	// Create audit entry if enabled
	if e.config.AuditEnabled {
		auditEntry := e.createAuditEntry(request, result)
		result.AuditLog = auditEntry
	}

	result.ProcessingTime = time.Since(startTime)

	// Cache the result
	if e.config.CacheEnabled {
		e.cache.Set(cacheKey, result)
	}

	return result, nil
}

// evaluateAccess performs the core authorization logic
func (e *RBACEngine) evaluateAccess(request *AccessRequest) (bool, string, []string) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	appliedPolicies := []string{}

	// Check if user has any of the required roles for the resource
	for _, policy := range e.policies {
		if e.policyMatches(policy, request) {
			appliedPolicies = append(appliedPolicies, policy.ID)

			// Check if user has any required roles
			if e.userHasRequiredRole(request.UserRoles, policy.RequiredRoles) {
				// Check additional conditions
				if e.evaluateConditions(policy.Conditions, request.Context) {
					return true, fmt.Sprintf("Access granted by policy %s", policy.Name), appliedPolicies
				}
			}
		}
	}

	if len(appliedPolicies) > 0 {
		return false, "User lacks required roles or conditions not met", appliedPolicies
	}

	return false, "No applicable security policies found", appliedPolicies
}

// policyMatches checks if a policy applies to the request
func (e *RBACEngine) policyMatches(policy *SecurityPolicy, request *AccessRequest) bool {
	// Check resource match
	if policy.Resource != "*" && !e.resourceMatches(policy.Resource, request.Resource) {
		return false
	}

	// Check action match
	if len(policy.Actions) > 0 {
		actionMatched := false
		for _, action := range policy.Actions {
			if action == "*" || action == request.Action {
				actionMatched = true
				break
			}
		}
		if !actionMatched {
			return false
		}
	}

	return true
}

// resourceMatches checks if a resource pattern matches the request resource
func (e *RBACEngine) resourceMatches(pattern, resource string) bool {
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(resource, prefix)
	}

	return pattern == resource
}

// userHasRequiredRole checks if user has any of the required roles
func (e *RBACEngine) userHasRequiredRole(userRoles, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true // No specific roles required
	}

	userRoleSet := make(map[string]bool)
	for _, role := range userRoles {
		userRoleSet[role] = true

		// Check hierarchical roles if enabled
		if e.config.HierarchicalRoles {
			if roleObj, exists := e.roles[role]; exists {
				for _, parentRole := range roleObj.ParentRoles {
					userRoleSet[parentRole] = true
				}
			}
		}
	}

	for _, requiredRole := range requiredRoles {
		if userRoleSet[requiredRole] {
			return true
		}
	}

	return false
}

// evaluateConditions checks if conditions are met
func (e *RBACEngine) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, expectedValue := range conditions {
		contextValue, exists := context[key]
		if !exists {
			return false
		}

		if contextValue != expectedValue {
			return false
		}
	}

	return true
}

// generateCacheKey creates a cache key for the request
func (e *RBACEngine) generateCacheKey(request *AccessRequest) string {
	rolesStr := strings.Join(request.UserRoles, ",")
	return fmt.Sprintf("rbac:%s:%s:%s:%s", request.Subject, request.Resource, request.Action, rolesStr)
}

// createAuditEntry creates an audit log entry
func (e *RBACEngine) createAuditEntry(request *AccessRequest, result *AccessResult) *AuditEntry {
	return &AuditEntry{
		ID:        fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Subject:   request.Subject,
		Action:    request.Action,
		Resource:  request.Resource,
		Result:    result.Allowed,
		Reason:    result.Reason,
		Context:   request.Context,
	}
}

// AddPolicy adds a security policy to the engine
func (e *RBACEngine) AddPolicy(policy *SecurityPolicy) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	policy.UpdatedAt = time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}

	e.policies[policy.ID] = policy
}

// AddRole adds a role to the engine
func (e *RBACEngine) AddRole(role *Role) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now()
	}

	e.roles[role.ID] = role
}

// AddPermission adds a permission to the engine
func (e *RBACEngine) AddPermission(permission *Permission) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if permission.CreatedAt.IsZero() {
		permission.CreatedAt = time.Now()
	}

	e.permissions[permission.ID] = permission
}

// GetStats returns RBAC engine statistics
func (e *RBACEngine) GetStats() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	return map[string]interface{}{
		"total_policies":    len(e.policies),
		"total_roles":       len(e.roles),
		"total_permissions": len(e.permissions),
		"cache_size":        len(e.cache.decisions),
		"config":            e.config,
	}
}

// Cache methods
func (c *RBACCache) Get(key string) *CachedDecision {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	decision, exists := c.decisions[key]
	if !exists {
		return nil
	}

	return decision
}

func (c *RBACCache) Set(key string, result *AccessResult) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Clean up expired entries if cache is full
	if len(c.decisions) >= c.maxSize {
		c.cleanup()
	}

	c.decisions[key] = &CachedDecision{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
		HitCount:  0,
	}
}

func (c *RBACCache) cleanup() {
	now := time.Now()
	for key, decision := range c.decisions {
		if decision.ExpiresAt.Before(now) {
			delete(c.decisions, key)
		}
	}
}

func (d *CachedDecision) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}
