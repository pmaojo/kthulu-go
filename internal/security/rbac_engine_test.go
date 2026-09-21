package security

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func testEngine(cfg *RBACConfig) *RBACEngine {
	if cfg == nil {
		cfg = &RBACConfig{
			CacheEnabled:      true,
			CacheTTL:          time.Minute,
			AuditEnabled:      true,
			DefaultDenyPolicy: true,
			HierarchicalRoles: true,
		}
	}
	return NewRBACEngine(cfg)
}

func allow(t *testing.T, e *RBACEngine, req *AccessRequest) *AccessResult {
	t.Helper()
	res, err := e.CheckAccess(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckAccess returned an error: %v", err)
	}
	return res
}

func TestNewRBACEngineDefaults(t *testing.T) {
	e := NewRBACEngine(nil)

	if !e.config.CacheEnabled || !e.config.DefaultDenyPolicy || !e.config.HierarchicalRoles {
		t.Errorf("nil config should enable caching, default-deny and hierarchical roles, got %+v", e.config)
	}
	if e.config.CacheTTL != 300*time.Second {
		t.Errorf("default cache TTL = %v, want 5m", e.config.CacheTTL)
	}
	if e.cache == nil || e.policies == nil || e.roles == nil || e.permissions == nil {
		t.Error("engine must be constructed with its maps and cache ready to use")
	}
}

func TestCheckAccessDeniesWhenNoPolicyApplies(t *testing.T) {
	res := allow(t, testEngine(nil), &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"admin"},
	})

	if res.Allowed {
		t.Error("an engine with no policies must deny")
	}
	if res.Reason != "No applicable security policies found" {
		t.Errorf("reason = %q, want the no-policy reason", res.Reason)
	}
	if len(res.AppliedPolicies) != 0 {
		t.Errorf("AppliedPolicies = %v, want empty", res.AppliedPolicies)
	}
}

func TestCheckAccessGrantsOnMatchingPolicy(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "orders-read", Name: "orders read", Resource: "orders",
		Actions: []string{"read"}, RequiredRoles: []string{"clerk"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"},
	})

	if !res.Allowed {
		t.Fatalf("expected allow, got deny: %s", res.Reason)
	}
	if len(res.AppliedPolicies) != 1 || res.AppliedPolicies[0] != "orders-read" {
		t.Errorf("AppliedPolicies = %v, want [orders-read]", res.AppliedPolicies)
	}
}

func TestCheckAccessDeniesWhenRoleMissing(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "orders-read", Name: "orders read", Resource: "orders",
		Actions: []string{"read"}, RequiredRoles: []string{"clerk"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "mallory", Resource: "orders", Action: "read", UserRoles: []string{"guest"},
	})

	if res.Allowed {
		t.Error("a user without the required role must be denied")
	}
	if res.Reason != "User lacks required roles or conditions not met" {
		t.Errorf("reason = %q, want the missing-role reason", res.Reason)
	}
	if len(res.AppliedPolicies) != 1 {
		t.Errorf("a policy that matched the resource should still be reported as applied, got %v", res.AppliedPolicies)
	}
}

func TestPolicyMatching(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		actions  []string
		reqRes   string
		reqAct   string
		want     bool
	}{
		{"exact resource and action", "orders", []string{"read"}, "orders", "read", true},
		{"action not listed", "orders", []string{"read"}, "orders", "delete", false},
		{"wildcard action", "orders", []string{"*"}, "orders", "delete", true},
		{"wildcard resource", "*", []string{"read"}, "anything", "read", true},
		{"prefix resource", "orders/*", []string{"read"}, "orders/42", "read", true},
		{"prefix does not match other tree", "orders/*", []string{"read"}, "invoices/42", "read", false},
		{"different resource", "orders", []string{"read"}, "invoices", "read", false},
		{"no actions listed matches any action", "orders", nil, "orders", "delete", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := testEngine(nil)
			e.AddPolicy(&SecurityPolicy{
				ID: "p", Name: "p", Resource: tc.resource, Actions: tc.actions,
			})

			res := allow(t, e, &AccessRequest{
				Subject: "u", Resource: tc.reqRes, Action: tc.reqAct,
			})

			if res.Allowed != tc.want {
				t.Errorf("allowed = %v, want %v (reason %q)", res.Allowed, tc.want, res.Reason)
			}
		})
	}
}

// A bare prefix pattern matches on the string, so "order*" also covers
// "orders-archive". Pinned so the pattern language cannot widen silently.
func TestResourcePrefixIsStringPrefixNotPathSegment(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{ID: "p", Name: "p", Resource: "order*", Actions: []string{"read"}})

	for _, resource := range []string{"orders", "order-archive", "orders/42"} {
		res := allow(t, e, &AccessRequest{Subject: "u", Resource: resource, Action: "read"})
		if !res.Allowed {
			t.Errorf("resource %q should match pattern order*", resource)
		}
	}

	res := allow(t, e, &AccessRequest{Subject: "u", Resource: "invoices", Action: "read"})
	if res.Allowed {
		t.Error("invoices must not match pattern order*")
	}
}

func TestEmptyRequiredRolesAllowsAnyRole(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{ID: "public", Name: "public", Resource: "docs", Actions: []string{"read"}})

	res := allow(t, e, &AccessRequest{Subject: "anon", Resource: "docs", Action: "read"})
	if !res.Allowed {
		t.Errorf("a policy with no required roles should allow, got %q", res.Reason)
	}
}

func TestHierarchicalRolesGrantThroughDirectParent(t *testing.T) {
	e := testEngine(nil)
	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "*",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"senior"},
	})
	if !res.Allowed {
		t.Errorf("senior inherits admin through its parent role, got %q", res.Reason)
	}
}

// Parent lookup walks the whole chain: junior -> senior -> admin reaches
// admin. Declaring a hierarchy and then honouring only its first step refused
// access the hierarchy says the user has.
func TestHierarchicalRolesAreOneLevelByDefault(t *testing.T) {
	e := testEngine(nil)
	e.AddRole(&Role{ID: "junior", Name: "junior", ParentRoles: []string{"senior"}})
	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "*",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"junior"},
	})
	if res.Allowed {
		t.Error("RoleHierarchyDepth defaults to one level; a grandparent role must not be granted")
	}
}

// RoleHierarchyDepth is a deployment's own choice, not a fixed policy this
// engine imposes: -1 opts into the full chain, however deep.
func TestRoleHierarchyDepthUnlimitedReachesTheWholeChain(t *testing.T) {
	e := testEngine(&RBACConfig{
		CacheEnabled: false, HierarchicalRoles: true, RoleHierarchyDepth: -1,
	})
	e.AddRole(&Role{ID: "junior", Name: "junior", ParentRoles: []string{"senior"}})
	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "*",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"junior"},
	})
	if !res.Allowed {
		t.Errorf("RoleHierarchyDepth: -1 must reach the whole chain, got %q", res.Reason)
	}
}

// A configured depth of N reaches exactly N levels up, not more.
func TestRoleHierarchyDepthStopsAtTheConfiguredLevel(t *testing.T) {
	e := testEngine(&RBACConfig{
		CacheEnabled: false, HierarchicalRoles: true, RoleHierarchyDepth: 2,
	})
	e.AddRole(&Role{ID: "junior", Name: "junior", ParentRoles: []string{"senior"}})
	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})
	e.AddRole(&Role{ID: "admin", Name: "admin", ParentRoles: []string{"root"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "vault",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})
	e.AddPolicy(&SecurityPolicy{
		ID: "root-only", Name: "root only", Resource: "vault-of-vaults",
		Actions: []string{"*"}, RequiredRoles: []string{"root"},
	})

	granted := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"junior"},
	})
	if !granted.Allowed {
		t.Errorf("depth 2 must reach admin (junior -> senior -> admin), got %q", granted.Reason)
	}

	denied := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault-of-vaults", Action: "read", UserRoles: []string{"junior"},
	})
	if denied.Allowed {
		t.Error("depth 2 must not reach root, a third level up")
	}
}

// A cycle must not hang the walk regardless of the configured depth,
// including the unlimited (-1) setting.
func TestRoleHierarchyCycleTerminatesEvenWhenUnlimited(t *testing.T) {
	e := testEngine(&RBACConfig{
		CacheEnabled: false, HierarchicalRoles: true, RoleHierarchyDepth: -1,
	})
	e.AddRole(&Role{ID: "a", Name: "a", ParentRoles: []string{"b"}})
	e.AddRole(&Role{ID: "b", Name: "b", ParentRoles: []string{"a"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "*",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})

	done := make(chan bool, 1)
	go func() {
		done <- allow(t, e, &AccessRequest{
			Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"a"},
		}).Allowed
	}()

	select {
	case allowed := <-done:
		if allowed {
			t.Error("neither a nor b reaches admin")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("role expansion did not terminate on a cycle with unlimited depth")
	}
}

func TestHierarchicalRolesDisabled(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: false, HierarchicalRoles: false})
	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "admin-only", Name: "admin only", Resource: "*",
		Actions: []string{"*"}, RequiredRoles: []string{"admin"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"senior"},
	})
	if res.Allowed {
		t.Error("with HierarchicalRoles off, a parent role must not be inherited")
	}
}

func TestConditionsGateAccess(t *testing.T) {
	tests := []struct {
		name       string
		conditions map[string]interface{}
		reqContext map[string]interface{}
		want       bool
	}{
		{"no conditions", nil, nil, true},
		{"condition met", map[string]interface{}{"mfa": true}, map[string]interface{}{"mfa": true}, true},
		{"condition not met", map[string]interface{}{"mfa": true}, map[string]interface{}{"mfa": false}, false},
		{"context key absent", map[string]interface{}{"mfa": true}, map[string]interface{}{}, false},
		{"nil context", map[string]interface{}{"mfa": true}, nil, false},
		{
			"every condition must hold",
			map[string]interface{}{"mfa": true, "region": "eu"},
			map[string]interface{}{"mfa": true, "region": "us"},
			false,
		},
		{
			"extra context keys are ignored",
			map[string]interface{}{"mfa": true},
			map[string]interface{}{"mfa": true, "device": "laptop"},
			true,
		},
		{
			"slice values compare by content, not identity",
			map[string]interface{}{"tenants": []string{"a", "b"}},
			map[string]interface{}{"tenants": []string{"a", "b"}},
			true,
		},
		{
			"slice values with different content are rejected",
			map[string]interface{}{"tenants": []string{"a", "b"}},
			map[string]interface{}{"tenants": []string{"a"}},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := testEngine(&RBACConfig{CacheEnabled: false})
			e.AddPolicy(&SecurityPolicy{
				ID: "p", Name: "p", Resource: "vault", Actions: []string{"read"},
				Conditions: tc.conditions,
			})

			res := allow(t, e, &AccessRequest{
				Subject: "u", Resource: "vault", Action: "read", Context: tc.reqContext,
			})
			if res.Allowed != tc.want {
				t.Errorf("allowed = %v, want %v", res.Allowed, tc.want)
			}
		})
	}
}

// A condition holding a map or slice must not take the authorization path
// down with it: comparing those with == panics at runtime.
func TestConditionWithNonComparableValueDoesNotPanic(t *testing.T) {
	for _, value := range []interface{}{
		[]string{"a"},
		map[string]interface{}{"k": "v"},
		[]interface{}{1, 2},
	} {
		e := testEngine(&RBACConfig{CacheEnabled: false})
		e.AddPolicy(&SecurityPolicy{
			ID: "p", Name: "p", Resource: "doc", Actions: []string{"read"},
			Conditions: map[string]interface{}{"scope": value},
		})

		res, err := e.CheckAccess(context.Background(), &AccessRequest{
			Subject: "u", Resource: "doc", Action: "read",
			Context: map[string]interface{}{"scope": value},
		})
		if err != nil {
			t.Fatalf("CheckAccess(%T) returned an error: %v", value, err)
		}
		if !res.Allowed {
			t.Errorf("condition %T equal to the context value should allow", value)
		}
	}
}

// The cache key covers the context, so a decision made under a satisfying
// context is not replayed for a request that does not satisfy it.
func TestCacheDoesNotReplayDecisionsAcrossContexts(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "mfa-only", Name: "mfa only", Resource: "vault", Actions: []string{"read"},
		RequiredRoles: []string{"staff"},
		Conditions:    map[string]interface{}{"mfa": true},
	})

	granted := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"staff"},
		Context: map[string]interface{}{"mfa": true},
	})
	if !granted.Allowed {
		t.Fatalf("request satisfying the condition should be allowed, got %q", granted.Reason)
	}

	denied := allow(t, e, &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"staff"},
		Context: map[string]interface{}{"mfa": false},
	})
	if denied.Allowed {
		t.Errorf("the same subject without mfa must be denied (cache hit = %v)", denied.CacheHit)
	}
}

func TestCacheServesRepeatedIdenticalRequests(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "p", Name: "p", Resource: "orders", Actions: []string{"read"},
		RequiredRoles: []string{"clerk"},
	})
	req := &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"},
	}

	first := allow(t, e, req)
	if first.CacheHit {
		t.Error("the first decision cannot be a cache hit")
	}

	second := allow(t, e, req)
	if !second.CacheHit {
		t.Error("an identical request should be served from the cache")
	}
	if second.Allowed != first.Allowed {
		t.Errorf("cached decision = %v, want %v", second.Allowed, first.Allowed)
	}
}

func TestCacheDisabledAlwaysReevaluates(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: false})
	e.AddPolicy(&SecurityPolicy{ID: "p", Name: "p", Resource: "orders", Actions: []string{"read"}})
	req := &AccessRequest{Subject: "alice", Resource: "orders", Action: "read"}

	allow(t, e, req)
	second := allow(t, e, req)
	if second.CacheHit {
		t.Error("no request may report a cache hit when caching is disabled")
	}
}

// A policy added after a decision was cached must take effect.
func TestCachedDenyIsNotServedAfterTTL(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: true, CacheTTL: 20 * time.Millisecond})
	req := &AccessRequest{Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"}}

	if res := allow(t, e, req); res.Allowed {
		t.Fatal("expected the first decision to deny")
	}

	e.AddPolicy(&SecurityPolicy{
		ID: "p", Name: "p", Resource: "orders", Actions: []string{"read"},
		RequiredRoles: []string{"clerk"},
	})
	time.Sleep(40 * time.Millisecond)

	if res := allow(t, e, req); !res.Allowed {
		t.Errorf("once the cached deny expired the new policy should apply, got %q", res.Reason)
	}
}

func TestAuditEntryRecordsTheDecision(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: false, AuditEnabled: true})
	e.AddPolicy(&SecurityPolicy{
		ID: "p", Name: "p", Resource: "orders", Actions: []string{"read"},
		RequiredRoles: []string{"clerk"},
	})

	res := allow(t, e, &AccessRequest{
		Subject: "mallory", Resource: "orders", Action: "read", UserRoles: []string{"guest"},
	})

	if res.AuditLog == nil {
		t.Fatal("audit is enabled, so the result must carry an audit entry")
	}
	if res.AuditLog.Subject != "mallory" || res.AuditLog.Resource != "orders" || res.AuditLog.Action != "read" {
		t.Errorf("audit entry does not describe the request: %+v", res.AuditLog)
	}
	if res.AuditLog.Result != res.Allowed {
		t.Errorf("audit entry result = %v, want %v", res.AuditLog.Result, res.Allowed)
	}
}

func TestAuditDisabledLeavesNoEntry(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: false, AuditEnabled: false})

	res := allow(t, e, &AccessRequest{Subject: "u", Resource: "orders", Action: "read"})
	if res.AuditLog != nil {
		t.Error("no audit entry should be produced when audit is disabled")
	}
}

func TestAddRoleAndPermissionAreVisibleInStats(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{ID: "p1", Name: "p1", Resource: "*"})
	e.AddRole(&Role{ID: "r1", Name: "r1"})
	e.AddPermission(&Permission{ID: "perm1", Name: "perm1", Resource: "orders", Action: "read"})

	stats := e.GetStats()
	for key, want := range map[string]int{"total_policies": 1, "total_roles": 1, "total_permissions": 1} {
		if got, ok := stats[key].(int); !ok || got != want {
			t.Errorf("stats[%q] = %v, want %d", key, stats[key], want)
		}
	}
}

func TestCachedDecisionExpiry(t *testing.T) {
	fresh := &CachedDecision{ExpiresAt: time.Now().Add(time.Minute)}
	if fresh.IsExpired() {
		t.Error("a decision expiring in the future is not expired")
	}

	stale := &CachedDecision{ExpiresAt: time.Now().Add(-time.Minute)}
	if !stale.IsExpired() {
		t.Error("a decision whose deadline has passed is expired")
	}
}

// CheckAccess is reachable from every request handler, so concurrent callers
// must not race. Run this file with -race to make the check meaningful.
func TestCheckAccessIsSafeUnderConcurrency(t *testing.T) {
	e := testEngine(nil)
	for i := 0; i < 5; i++ {
		e.AddPolicy(&SecurityPolicy{
			ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("p%d", i),
			Resource: fmt.Sprintf("res%d", i), Actions: []string{"read"},
			RequiredRoles: []string{"clerk"},
		})
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, err := e.CheckAccess(context.Background(), &AccessRequest{
				Subject: fmt.Sprintf("u%d", i%3), Resource: fmt.Sprintf("res%d", i%5),
				Action: "read", UserRoles: []string{"clerk"},
			})
			if err != nil {
				t.Errorf("CheckAccess returned an error: %v", err)
				return
			}
			if !res.Allowed {
				t.Errorf("request %d should be allowed, got %q", i, res.Reason)
			}
		}(i)
	}
	wg.Wait()
}

// The decision handed back belongs to the caller; writing to it must not
// change what the next request reads out of the cache.
func TestMutatingAReturnedResultDoesNotPoisonTheCache(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "p", Name: "p", Resource: "orders", Actions: []string{"read"},
		RequiredRoles: []string{"clerk"},
	})
	req := &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"},
	}

	first := allow(t, e, req)
	if !first.Allowed {
		t.Fatalf("expected allow, got %q", first.Reason)
	}

	first.Allowed = false
	first.Reason = "tampered"

	second := allow(t, e, req)
	if !second.Allowed || second.Reason == "tampered" {
		t.Errorf("cached decision was altered through the returned result: allowed=%v reason=%q",
			second.Allowed, second.Reason)
	}
}
