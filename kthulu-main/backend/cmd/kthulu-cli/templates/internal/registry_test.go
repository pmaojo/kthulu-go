package modules

import (
	"testing"

	"go.uber.org/fx"
)

func TestRegistry_RegisterGetAndDeregister(t *testing.T) {
	registry := NewRegistry()
	opts := fx.Options()
	registry.Register("test", opts)

	module, ok := registry.GetModule("test")
	if !ok {
		t.Fatalf("expected module to be registered")
	}
	if module.Name != "test" {
		t.Fatalf("unexpected module name: %s", module.Name)
	}
	if len(registry.GetAllModules()) != 1 {
		t.Fatalf("expected 1 module, got %d", len(registry.GetAllModules()))
	}
	if registry.GetModuleOptions() == nil {
		t.Fatalf("expected module options to be returned")
	}

	registry.Deregister("test")
	if _, ok := registry.GetModule("test"); ok {
		t.Fatalf("expected module to be deregistered")
	}
}

func TestRegistry_GetAllModulesReturnsCopy(t *testing.T) {
	registry := NewRegistry()
	registry.Register("test", fx.Options())

	modules := registry.GetAllModules()
	modules["other"] = Module{Name: "other"}

	if _, exists := registry.GetModule("other"); exists {
		t.Fatalf("modifying GetAllModules result should not affect registry")
	}
}
