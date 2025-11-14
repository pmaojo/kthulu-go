// @kthulu:core
package modules

import (
	"os"

	"go.uber.org/fx"
)

// ModuleSet manages available modules and builds Fx options.
// It's designed to be injectable and configurable rather than global.
type ModuleSet struct {
	registry *Registry
}

// NewModuleSet creates a module set backed by the provided registry.
// If registry is nil, a new empty registry is used.
func NewModuleSet(registry *Registry) *ModuleSet {
	if registry == nil {
		registry = NewRegistry()
	}
	return &ModuleSet{registry: registry}
}

// Register adds a module to the set.
func (ms *ModuleSet) Register(name string, option fx.Option) {
	ms.registry.Register(name, option)
}

// Deregister removes a module from the set.
func (ms *ModuleSet) Deregister(name string) {
	ms.registry.Deregister(name)
}

// IsRegistered checks if a module is registered.
func (ms *ModuleSet) IsRegistered(name string) bool {
	_, exists := ms.registry.GetModule(name)
	return exists
}

// GetRegisteredModules returns the names of all registered modules.
func (ms *ModuleSet) GetRegisteredModules() []string {
	modules := ms.registry.GetAllModules()
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	return names
}

// Build constructs fx.Options for the selected modules.
// If active is empty, all registered modules are included.
func (ms *ModuleSet) Build(active []string) fx.Option {
	// Always include shared infrastructure
	opts := []fx.Option{
		fx.Provide(NewRouteRegistry),
		// Centralized repository providers to avoid duplication
		SharedRepositoryProviders(),
	}

	if len(active) == 0 {
		// Include all registered modules
		opts = append(opts, ms.registry.GetModuleOptions())
		return fx.Options(opts...)
	}

	// Include only specified modules
	for _, name := range active {
		if m, ok := ms.registry.GetModule(name); ok {
			opts = append(opts, m.Options)
		}
	}

	return fx.Options(opts...)
}

// ModuleSetBuilder provides a fluent interface for building module sets.
type ModuleSetBuilder struct {
	source    *Registry
	moduleSet *ModuleSet
}

// NewModuleSetBuilder creates a new builder using the provided source registry.
// Modules registered in the source can be selected and added to the resulting set.
func NewModuleSetBuilder(source *Registry) *ModuleSetBuilder {
	if source == nil {
		source = NewRegistry()
	}
	return &ModuleSetBuilder{
		source:    source,
		moduleSet: NewModuleSet(NewRegistry()),
	}
}

// WithModule selects a module from the source registry to be included in the set.
func (b *ModuleSetBuilder) WithModule(name string) *ModuleSetBuilder {
	if m, ok := b.source.GetModule(name); ok {
		b.moduleSet.Register(name, m.Options)
	}
	return b
}

var coreModuleNames = []string{"health", "oauth-sso", "user", "access", "notifier", "static"}
var erpModuleNames = []string{"organization", "contact", "product", "invoice", "inventory", "calendar", "realtime", "verifactu"}

func init() {
	if os.Getenv("LEGACY_AUTH") == "true" {
		coreModuleNames = append(coreModuleNames, "auth")
	}
}

// WithCoreModules adds all core modules (health, auth, user, access, notifier, static).
func (b *ModuleSetBuilder) WithCoreModules() *ModuleSetBuilder {
	for _, name := range coreModuleNames {
		b.WithModule(name)
	}
	return b
}

// WithERPModules adds all ERP-lite modules (org, contacts, products, invoices, inventory, calendar).
func (b *ModuleSetBuilder) WithERPModules() *ModuleSetBuilder {
	for _, name := range erpModuleNames {
		b.WithModule(name)
	}
	return b
}

// WithAllModules adds both core and ERP modules.
func (b *ModuleSetBuilder) WithAllModules() *ModuleSetBuilder {
	for name := range b.source.GetAllModules() {
		b.WithModule(name)
	}
	return b
}

// Build returns the configured module set.
func (b *ModuleSetBuilder) Build() *ModuleSet {
	return b.moduleSet
}

// DefaultModuleSet creates a module set with all modules from the provided registry.
// If registry is nil, an empty set is returned.
func DefaultModuleSet(registry *Registry) *ModuleSet {
	return NewModuleSetBuilder(registry).WithAllModules().Build()
}

// BuildModules creates fx.Options for the specified modules using a default module set.
func BuildModules(active []string, registry *Registry) fx.Option {
	moduleSet := DefaultModuleSet(registry)
	return moduleSet.Build(active)
}
