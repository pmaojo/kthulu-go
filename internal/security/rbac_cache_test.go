package security

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// Subject and resource arrive from request data. Joining them with a plain
// separator let a caller who controls either one land on another request's
// cache entry: "alice:orders" + "read" and "alice" + "orders:read" produced
// the same key.
func TestCacheKeyIsNotForgeableThroughSeparators(t *testing.T) {
	tests := []struct {
		name string
		a, b *AccessRequest
	}{
		{
			name: "subject absorbs the resource",
			a:    &AccessRequest{Subject: "alice", Resource: "orders:read", Action: "read"},
			b:    &AccessRequest{Subject: "alice:orders", Resource: "read", Action: "read"},
		},
		{
			name: "action absorbs the role list",
			a:    &AccessRequest{Subject: "u", Resource: "r", Action: "read", UserRoles: []string{"clerk"}},
			b:    &AccessRequest{Subject: "u", Resource: "r", Action: "read:clerk"},
		},
		{
			name: "one role spelling two",
			a:    &AccessRequest{Subject: "u", Resource: "r", Action: "read", UserRoles: []string{"clerk", "auditor"}},
			b:    &AccessRequest{Subject: "u", Resource: "r", Action: "read", UserRoles: []string{"clerk,auditor"}},
		},
		{
			name: "context key absorbs its own value",
			a:    &AccessRequest{Subject: "u", Resource: "r", Action: "read", Context: map[string]interface{}{"mfa": "true"}},
			b:    &AccessRequest{Subject: "u", Resource: "r", Action: "read", Context: map[string]interface{}{"mfa=true": ""}},
		},
	}

	e := testEngine(nil)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, other := e.generateCacheKey(tc.a), e.generateCacheKey(tc.b); got == other {
				t.Errorf("two different requests share cache key %q", got)
			}
		})
	}
}

// Roles authorize as a set, so two spellings of the same set must answer from
// one entry rather than re-evaluating and storing a second.
func TestCacheKeyIgnoresRoleOrderAndRepetition(t *testing.T) {
	e := testEngine(nil)
	base := &AccessRequest{Subject: "u", Resource: "r", Action: "read", UserRoles: []string{"clerk", "auditor"}}

	for _, roles := range [][]string{
		{"auditor", "clerk"},
		{"clerk", "auditor", "clerk"},
	} {
		other := &AccessRequest{Subject: "u", Resource: "r", Action: "read", UserRoles: roles}
		if e.generateCacheKey(base) != e.generateCacheKey(other) {
			t.Errorf("roles %v should share a cache entry with %v", roles, base.UserRoles)
		}
	}
}

// A decision cached under the old rule set outlives the rule that produced it
// unless the write invalidates it. Five minutes of granted access after a
// policy is tightened is the default TTL, not a rounding error.
func TestTighteningAPolicyTakesEffectImmediately(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "orders", Name: "orders", Resource: "orders",
		Actions: []string{"read"}, RequiredRoles: []string{"clerk"},
	})
	req := &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"},
	}
	if !allow(t, e, req).Allowed {
		t.Fatal("clerk should be allowed before the policy is tightened")
	}

	e.AddPolicy(&SecurityPolicy{
		ID: "orders", Name: "orders", Resource: "orders",
		Actions: []string{"read"}, RequiredRoles: []string{"admin"},
	})

	if res := allow(t, e, req); res.Allowed {
		t.Errorf("clerk still allowed after the policy required admin (cache hit: %v)", res.CacheHit)
	}
}

// Adding a role changes who inherits what, so it invalidates decisions too.
func TestAddingARoleInvalidatesCachedDenials(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "vault", Name: "vault", Resource: "vault",
		Actions: []string{"read"}, RequiredRoles: []string{"admin"},
	})
	req := &AccessRequest{
		Subject: "u", Resource: "vault", Action: "read", UserRoles: []string{"senior"},
	}
	if allow(t, e, req).Allowed {
		t.Fatal("senior is not admin yet")
	}

	e.AddRole(&Role{ID: "senior", Name: "senior", ParentRoles: []string{"admin"}})

	if !allow(t, e, req).Allowed {
		t.Error("senior inherits admin once the role exists; the stale denial was served instead")
	}
}

// maxSize was never enforced: cleanup only drops expired entries, so a cache
// whose entries are all live grew without limit. The request context is part
// of the key, so anyone who can vary it can mint entries.
func TestCacheStaysWithinItsSizeLimit(t *testing.T) {
	e := testEngine(nil)
	e.cache.maxSize = 50

	for i := 0; i < 500; i++ {
		allow(t, e, &AccessRequest{
			Subject: "u", Resource: "orders", Action: "read",
			Context: map[string]interface{}{"nonce": i},
		})
	}

	if got := e.cache.Len(); got > e.cache.maxSize {
		t.Errorf("cache holds %d decisions, limit is %d", got, e.cache.maxSize)
	}
}

func TestEvictionRemovesTheOldestEntriesFirst(t *testing.T) {
	c := &RBACCache{decisions: map[string]*CachedDecision{}, ttl: time.Minute, maxSize: 3}
	for _, key := range []string{"a", "b", "c"} {
		c.Set(key, &AccessResult{Allowed: true})
		time.Sleep(time.Millisecond)
	}

	c.Set("d", &AccessResult{Allowed: true})

	if c.Hit("a") != nil {
		t.Error("the oldest entry should have been evicted")
	}
	for _, key := range []string{"c", "d"} {
		if c.Hit(key) == nil {
			t.Errorf("%q is among the newest entries and should still be cached", key)
		}
	}
}

// With caching on, every request still has to be audited in its own right.
// Handing back the first request's entry left one audit record standing for
// every later request that hit the same decision.
func TestEachCacheHitIsAuditedOnItsOwn(t *testing.T) {
	e := testEngine(nil)
	req := &AccessRequest{Subject: "alice", Resource: "orders", Action: "read"}

	first := allow(t, e, req)
	if first.AuditLog == nil {
		t.Fatal("the first request must be audited")
	}
	time.Sleep(2 * time.Millisecond)

	second := allow(t, e, req)
	if !second.CacheHit {
		t.Fatal("the second identical request should come from the cache")
	}
	if second.AuditLog == nil {
		t.Fatal("a cached decision is still an access, and must be audited")
	}
	if second.AuditLog.ID == first.AuditLog.ID {
		t.Error("the cache hit replayed the first request's audit entry")
	}
	if second.AuditLog == first.AuditLog {
		t.Error("audit entries are shared between callers; concurrent writers would race")
	}
	if second.AuditLog.Result != first.AuditLog.Result {
		t.Error("the audited outcome must match the decision that was served")
	}
}

func TestAuditIsNotResurrectedWhenDisabled(t *testing.T) {
	e := testEngine(&RBACConfig{CacheEnabled: true, CacheTTL: time.Minute, AuditEnabled: false})
	req := &AccessRequest{Subject: "alice", Resource: "orders", Action: "read"}

	allow(t, e, req)
	if second := allow(t, e, req); second.AuditLog != nil {
		t.Error("audit is off; a cache hit must not carry an entry")
	}
}

// The returned decision belongs to the caller. A shallow copy still shared the
// slice and the map underneath it.
func TestWritingIntoAReturnedDecisionDoesNotReachTheCache(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "orders", Name: "orders", Resource: "orders",
		Actions: []string{"read"}, RequiredRoles: []string{"clerk"},
	})
	req := &AccessRequest{
		Subject: "alice", Resource: "orders", Action: "read", UserRoles: []string{"clerk"},
	}

	first := allow(t, e, req)
	if len(first.AppliedPolicies) == 0 {
		t.Fatal("the matching policy should be reported")
	}
	first.AppliedPolicies[0] = "tampered"
	first.Conditions["tampered"] = true

	second := allow(t, e, req)
	if second.AppliedPolicies[0] == "tampered" {
		t.Error("AppliedPolicies is shared with the cached decision")
	}
	if _, found := second.Conditions["tampered"]; found {
		t.Error("Conditions is shared with the cached decision")
	}
}

// GetStats read the cache map while holding only the engine's lock, which does
// not cover it. Run with -race.
func TestStatsAndAccessChecksDoNotRace(t *testing.T) {
	e := testEngine(nil)
	e.AddPolicy(&SecurityPolicy{
		ID: "orders", Name: "orders", Resource: "orders", Actions: []string{"read"},
	})

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				if stats := e.GetStats(); stats["cache_size"] == nil {
					t.Error("stats must report the cache size")
				}
			case 1:
				e.AddRole(&Role{ID: fmt.Sprintf("r%d", i), Name: "r"})
			default:
				_, err := e.CheckAccess(context.Background(), &AccessRequest{
					Subject: fmt.Sprintf("u%d", i), Resource: "orders", Action: "read",
				})
				if err != nil {
					t.Errorf("CheckAccess: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()
}

// A role cycle is a configuration mistake, not a reason to hang the request.
func TestCyclicRolesTerminate(t *testing.T) {
	e := testEngine(nil)
	e.AddRole(&Role{ID: "a", Name: "a", ParentRoles: []string{"b"}})
	e.AddRole(&Role{ID: "b", Name: "b", ParentRoles: []string{"a"}})
	e.AddPolicy(&SecurityPolicy{
		ID: "vault", Name: "vault", Resource: "vault",
		Actions: []string{"read"}, RequiredRoles: []string{"admin"},
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
		t.Fatal("role expansion did not terminate on a cycle")
	}
}

// CheckAccess takes a context; a caller that has given up should not be made
// to wait for a decision it will discard.
func TestCancelledContextIsRefused(t *testing.T) {
	e := testEngine(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := e.CheckAccess(ctx, &AccessRequest{Subject: "u", Resource: "r", Action: "read"})
	if err == nil {
		t.Fatalf("a cancelled context should be reported, got result %+v", res)
	}
	if res != nil {
		t.Error("no decision should be returned alongside the error")
	}
}
