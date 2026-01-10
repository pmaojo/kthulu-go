// @kthulu:core
package modules

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"sync"
)

// RouteRegistrar defines the interface for components that can register HTTP routes
type RouteRegistrar interface {
	RegisterRoutes(r chi.Router)
}

// Module represents a functional module in the application
type Module struct {
	Name    string
	Options fx.Option
}

// Registry holds all available modules
type Registry struct {
	mu      sync.RWMutex
	modules map[string]Module
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(name string, options fx.Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.modules[name] = Module{
		Name:    name,
		Options: options,
	}
}

// Deregister removes a module from the registry
func (r *Registry) Deregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.modules, name)
}

// GetModule returns a module by name
func (r *Registry) GetModule(name string) (Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.modules[name]
	return module, exists
}

// GetAllModules returns all registered modules
func (r *Registry) GetAllModules() map[string]Module {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modulesCopy := make(map[string]Module, len(r.modules))
	for name, module := range r.modules {
		modulesCopy[name] = module
	}

	return modulesCopy
}

// GetModuleOptions returns fx.Options for all registered modules
func (r *Registry) GetModuleOptions() fx.Option {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var options []fx.Option
	for _, module := range r.modules {
		options = append(options, module.Options)
	}
	return fx.Options(options...)
}
