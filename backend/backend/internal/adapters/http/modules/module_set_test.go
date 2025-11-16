// @kthulu:core
package modules

import (
	"testing"

	"go.uber.org/fx"
)

func TestModuleSetBuilder(t *testing.T) {
	t.Run("empty module set", func(t *testing.T) {
		registry := NewRegistry()
		moduleSet := NewModuleSetBuilder(registry).Build()

		if len(moduleSet.GetRegisteredModules()) != 0 {
			t.Errorf("Expected empty module set, got %d modules", len(moduleSet.GetRegisteredModules()))
		}
	})

	t.Run("with core modules", func(t *testing.T) {
		registry := NewRegistry()
		RegisterBuiltinModules(registry)
		moduleSet := NewModuleSetBuilder(registry).WithCoreModules().Build()

		expectedModules := coreModuleNames
		registeredModules := moduleSet.GetRegisteredModules()

		if len(registeredModules) != len(expectedModules) {
			t.Errorf("Expected %d modules, got %d", len(expectedModules), len(registeredModules))
		}

		for _, expected := range expectedModules {
			if !moduleSet.IsRegistered(expected) {
				t.Errorf("Expected module %s to be registered", expected)
			}
		}
	})

	t.Run("with ERP modules", func(t *testing.T) {
		registry := NewRegistry()
		RegisterBuiltinModules(registry)
		moduleSet := NewModuleSetBuilder(registry).WithERPModules().Build()

		expectedModules := erpModuleNames
		registeredModules := moduleSet.GetRegisteredModules()

		if len(registeredModules) != len(expectedModules) {
			t.Errorf("Expected %d modules, got %d. Registered: %v", len(expectedModules), len(registeredModules), registeredModules)
		}

		for _, expected := range expectedModules {
			if !moduleSet.IsRegistered(expected) {
				t.Errorf("Expected module %s to be registered", expected)
			}
		}
	})

	t.Run("with all modules", func(t *testing.T) {
		registry := NewRegistry()
		RegisterBuiltinModules(registry)
		moduleSet := NewModuleSetBuilder(registry).WithAllModules().Build()

		// Should have all builtin modules
		expectedCount := len(BuiltinModules)
		registeredModules := moduleSet.GetRegisteredModules()

		if len(registeredModules) != expectedCount {
			t.Errorf("Expected %d modules, got %d. Registered: %v", expectedCount, len(registeredModules), registeredModules)
		}
	})

	t.Run("register and deregister", func(t *testing.T) {
		moduleSet := NewModuleSet(NewRegistry())

		// Register a test module
		testModule := fx.Options()
		moduleSet.Register("test", testModule)

		if !moduleSet.IsRegistered("test") {
			t.Error("Expected test module to be registered")
		}

		// Deregister the module
		moduleSet.Deregister("test")

		if moduleSet.IsRegistered("test") {
			t.Error("Expected test module to be deregistered")
		}
	})
}

func TestDefaultModuleSet(t *testing.T) {
	registry := NewRegistry()
	RegisterBuiltinModules(registry)
	moduleSet := DefaultModuleSet(registry)

	// Should have all builtin modules
	expectedCount := len(BuiltinModules)
	registeredModules := moduleSet.GetRegisteredModules()

	if len(registeredModules) != expectedCount {
		t.Errorf("Expected %d modules in default set, got %d. Registered: %v", expectedCount, len(registeredModules), registeredModules)
	}

	// Verify specific modules are present
	for _, module := range coreModuleNames {
		if !moduleSet.IsRegistered(module) {
			t.Errorf("Expected core module %s to be in default set", module)
		}
	}

	for _, module := range erpModuleNames {
		if !moduleSet.IsRegistered(module) {
			t.Errorf("Expected ERP module %s to be in default set", module)
		}
	}
}

func TestModuleSetProviderSelection(t *testing.T) {
	registry := NewRegistry()
	RegisterBuiltinModules(registry)
	moduleSet := NewModuleSet(registry)

	t.Run("product module only", func(t *testing.T) {
		providers := moduleSet.requiredProviderKeys([]string{"product"})
		if !containsProvider(providers, providerProductRepo) {
			t.Fatalf("expected product provider, got %v", providers)
		}
		if containsProvider(providers, providerContactRepo) {
			t.Fatalf("unexpected contact provider for product module: %v", providers)
		}
	})

	t.Run("inventory module pulls dependent product repo", func(t *testing.T) {
		providers := moduleSet.requiredProviderKeys([]string{"inventory"})
		if !containsProvider(providers, providerInventoryRepo) {
			t.Fatalf("expected inventory provider, got %v", providers)
		}
		if !containsProvider(providers, providerProductRepo) {
			t.Fatalf("expected product provider for inventory module, got %v", providers)
		}
	})
}

func containsProvider(providers []string, target string) bool {
	for _, provider := range providers {
		if provider == target {
			return true
		}
	}
	return false
}
