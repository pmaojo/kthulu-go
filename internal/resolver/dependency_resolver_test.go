package resolver

import (
	"slices"
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
)

func newResolver() *DependencyResolver {
	return NewDependencyResolver(&parser.ProjectAnalysis{
		Modules:      make(map[string]*parser.Module),
		Dependencies: []parser.Dependency{},
	})
}

func resolve(t *testing.T, modules ...string) *ResolutionPlan {
	t.Helper()
	plan, err := newResolver().ResolveDependencies(modules)
	if err != nil {
		t.Fatalf("ResolveDependencies(%v) returned an error: %v", modules, err)
	}
	return plan
}

// indexOf reports where module sits in the install order, or -1.
func indexOf(order []string, module string) int {
	return slices.Index(order, module)
}

func TestResolveExpandsTransitiveDependencies(t *testing.T) {
	// invoice needs user, organization, product and contact; those in turn
	// pull in auth, so the closure is wider than the direct rule.
	plan := resolve(t, "invoice")

	for _, want := range []string{"invoice", "user", "organization", "product", "contact", "auth"} {
		if !slices.Contains(plan.RequiredModules, want) {
			t.Errorf("%q missing from required modules %v", want, plan.RequiredModules)
		}
	}
}

func TestResolveLeafModuleHasOnlyItself(t *testing.T) {
	plan := resolve(t, "user")

	if len(plan.RequiredModules) != 1 || plan.RequiredModules[0] != "user" {
		t.Errorf("required modules = %v, want [user]", plan.RequiredModules)
	}
}

func TestResolveUnknownModuleIsTreatedAsLeaf(t *testing.T) {
	plan := resolve(t, "does-not-exist")

	if !slices.Contains(plan.RequiredModules, "does-not-exist") {
		t.Errorf("an unknown module should still be scheduled, got %v", plan.RequiredModules)
	}
}

func TestResolveDeduplicatesSharedDependencies(t *testing.T) {
	// payment and invoice both depend on user and organization.
	plan := resolve(t, "payment", "invoice")

	seen := map[string]int{}
	for _, module := range plan.RequiredModules {
		seen[module]++
	}
	for module, count := range seen {
		if count != 1 {
			t.Errorf("%q appears %d times in required modules", module, count)
		}
	}
}

func TestInstallOrderPutsEveryDependencyFirst(t *testing.T) {
	plan := resolve(t, "invoice", "payment", "verifactu", "calendar", "inventory", "oauthsso")

	rules := newResolver().rules
	for _, module := range plan.InstallOrder {
		position := indexOf(plan.InstallOrder, module)
		for _, dep := range rules[module] {
			depPosition := indexOf(plan.InstallOrder, dep)
			if depPosition == -1 {
				t.Errorf("%q depends on %q, which is not in the install order", module, dep)
				continue
			}
			if depPosition > position {
				t.Errorf("%q is installed at %d, after its dependant %q at %d",
					dep, depPosition, module, position)
			}
		}
	}
}

func TestInstallOrderCoversExactlyTheRequiredModules(t *testing.T) {
	plan := resolve(t, "invoice", "payment")

	if len(plan.InstallOrder) != len(plan.RequiredModules) {
		t.Fatalf("install order has %d entries, required modules %d",
			len(plan.InstallOrder), len(plan.RequiredModules))
	}
	for _, module := range plan.RequiredModules {
		if !slices.Contains(plan.InstallOrder, module) {
			t.Errorf("%q is required but never installed", module)
		}
	}
}

// Identical input must produce an identical plan: the generated project is
// scaffolded from these slices, so anything else makes the output unstable.
func TestResolutionIsDeterministic(t *testing.T) {
	first := resolve(t, "invoice", "payment", "calendar", "inventory")
	wantOrder := strings.Join(first.InstallOrder, ">")
	wantRequired := strings.Join(first.RequiredModules, ">")

	for i := 0; i < 50; i++ {
		plan := resolve(t, "invoice", "payment", "calendar", "inventory")

		if got := strings.Join(plan.InstallOrder, ">"); got != wantOrder {
			t.Fatalf("install order changed between runs:\n first: %s\n run %d: %s", wantOrder, i, got)
		}
		if got := strings.Join(plan.RequiredModules, ">"); got != wantRequired {
			t.Fatalf("required modules changed between runs:\n first: %s\n run %d: %s", wantRequired, i, got)
		}
	}
}

func TestResolutionIsIndependentOfRequestOrder(t *testing.T) {
	forwards := resolve(t, "invoice", "payment")
	backwards := resolve(t, "payment", "invoice")

	if !slices.Equal(forwards.InstallOrder, backwards.InstallOrder) {
		t.Errorf("asking for the same modules in another order changed the install order:\n%v\n%v",
			forwards.InstallOrder, backwards.InstallOrder)
	}
}

func TestCalculateInstallOrderReportsCycles(t *testing.T) {
	r := newResolver()
	r.rules["a"] = []string{"b"}
	r.rules["b"] = []string{"a"}

	_, err := r.calculateInstallOrder([]string{"a", "b"})
	if err == nil {
		t.Fatal("a cycle between a and b must be reported")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("error = %q, want it to name the circular dependency", err)
	}
}

func TestCalculateInstallOrderIgnoresDependenciesOutsideTheSet(t *testing.T) {
	r := newResolver()

	// auth depends on user, but user was not requested here.
	order, err := r.calculateInstallOrder([]string{"auth"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 1 || order[0] != "auth" {
		t.Errorf("order = %v, want [auth]", order)
	}
}

func TestCalculateInstallOrderOnEmptySet(t *testing.T) {
	order, err := newResolver().calculateInstallOrder(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("order = %v, want empty", order)
	}
}

func TestPlanWithoutACoreModuleIsWarnedAbout(t *testing.T) {
	r := newResolver()
	plan := &ResolutionPlan{
		RequiredModules: []string{"analytics"},
		Conflicts:       []ConflictInfo{},
		Warnings:        []string{},
	}

	r.detectConflicts(plan)

	if len(plan.Warnings) == 0 {
		t.Fatal("a plan with neither user nor auth should raise a warning")
	}
	if !strings.Contains(strings.Join(plan.Warnings, " "), "core authentication") {
		t.Errorf("warning does not mention the missing core modules: %v", plan.Warnings)
	}
}

func TestPlanWithACoreModuleIsNotWarnedAbout(t *testing.T) {
	r := newResolver()
	plan := &ResolutionPlan{
		RequiredModules: []string{"user", "analytics"},
		Conflicts:       []ConflictInfo{},
		Warnings:        []string{},
	}

	r.detectConflicts(plan)

	for _, warning := range plan.Warnings {
		if strings.Contains(warning, "core authentication") {
			t.Errorf("user is present, so no core-module warning is due: %q", warning)
		}
	}
}

// The core check asks only whether user or auth is in the plan; it does not
// verify that each module's own rule dependencies made it in. A hand-built
// plan can therefore be internally inconsistent and still pass unremarked.
func TestDetectConflictsDoesNotValidatePerModuleDependencies(t *testing.T) {
	r := newResolver()
	plan := &ResolutionPlan{
		RequiredModules: []string{"auth"}, // auth's rule needs user, absent here
		Conflicts:       []ConflictInfo{},
		Warnings:        []string{},
	}

	r.detectConflicts(plan)

	if len(plan.Conflicts) != 0 {
		t.Errorf("no conflict is raised for a missing rule dependency today, got %+v", plan.Conflicts)
	}
}

func TestCompletePlanHasNoMissingDependencyConflicts(t *testing.T) {
	plan := resolve(t, "invoice", "payment")

	for _, conflict := range plan.Conflicts {
		t.Errorf("a fully resolved plan should carry no conflicts, got %+v", conflict)
	}
}

func TestSuggestOptionalModulesSkipsAlreadyRequiredOnes(t *testing.T) {
	// user suggests notification and audit; ask for one of them up front.
	plan := resolve(t, "user", "audit")

	for _, optional := range plan.OptionalModules {
		if slices.Contains(plan.RequiredModules, optional) {
			t.Errorf("%q is already required, so it should not also be suggested", optional)
		}
	}
	if !slices.Contains(plan.OptionalModules, "notification") {
		t.Errorf("notification should still be suggested for user, got %v", plan.OptionalModules)
	}
}

func TestGetModuleInfoOnAnUnknownModule(t *testing.T) {
	_, err := newResolver().GetModuleInfo("nope")
	if err == nil {
		t.Error("an unknown module should be reported as an error, not an empty ModuleInfo")
	}
}

func TestGetModuleInfoDescribesAKnownModule(t *testing.T) {
	r := NewDependencyResolver(&parser.ProjectAnalysis{
		Modules: map[string]*parser.Module{
			"user": {Name: "user"},
		},
		Dependencies: []parser.Dependency{},
	})

	info, err := r.GetModuleInfo("user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "user" {
		t.Errorf("info.Name = %q, want user", info.Name)
	}
	if info.Description == "" || info.Category == "" {
		t.Errorf("module info should carry a description and a category, got %+v", info)
	}
}

func TestContainsHelper(t *testing.T) {
	haystack := []string{"user", "auth"}

	if !contains(haystack, "auth") {
		t.Error("contains should find an element that is present")
	}
	if contains(haystack, "invoice") {
		t.Error("contains should not find an element that is absent")
	}
	if contains(nil, "user") {
		t.Error("contains on an empty slice is always false")
	}
}
