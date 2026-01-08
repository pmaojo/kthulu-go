package modules

import "testing"

func TestRegisterBuiltinModules(t *testing.T) {
	registry := NewRegistry()
	RegisterBuiltinModules(registry)
	for name := range BuiltinModules {
		if _, ok := registry.GetModule(name); !ok {
			t.Fatalf("module %s not registered", name)
		}
	}
}
