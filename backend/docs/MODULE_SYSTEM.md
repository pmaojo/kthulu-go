# Kthulu Module System

The Kthulu framework uses an injectable, dynamic module system that eliminates global coupling and allows for flexible module composition.

## Key Features

- **Injectable**: No global variables or hard-coded module lists
- **Dynamic**: Modules can be enabled/disabled via configuration
- **Deduplication**: Shared repositories are centralized to avoid duplication
- **Testable**: Module sets can be easily mocked and tested
- **CLI-friendly**: Designed for selective module extraction by the CLI generator

## Architecture

### ModuleSet and ModuleSetBuilder

The `ModuleSet` manages a collection of modules and builds Fx options for them:

```go
// Create a custom module set
registry := modules.NewRegistry()
modules.RegisterBuiltinModules(registry)

moduleSet := modules.NewModuleSetBuilder(registry).
    WithModule("auth").
    WithModule("user").
    Build()

// Use in Fx application
app := fx.New(
    moduleSet.Build([]string{}), // Empty slice = all registered modules
)
```

### Targeted Repository Providers

Repositories are grouped into granular provider functions and resolved only for the modules that need them:

```go
var moduleProviderMap = map[string][]string{
    "product":   {providerProductRepo},
    "inventory": {providerInventoryRepo, providerProductRepo},
}

moduleSet := modules.DefaultModuleSet(registry)

// Only the inventory + product repositories are injected.
fx.New(moduleSet.Build([]string{"inventory"}))
```

### Named Dependencies

Use cases receive repositories via named dependencies:

```go
func NewAuthUseCase(p struct {
    UserRepository         repository.UserRepository         `name:"userRepository"`
    RefreshTokenRepository repository.RefreshTokenRepository `name:"refreshTokenRepository"`
    // ...
}) *AuthUseCase
```

## Configuration

Modules can be configured via environment variables:

```bash
# Enable all modules (default)
MODULES=

# Enable only core modules
MODULES=health,auth,user,access,notifier

# Enable multi-tenant setup
MODULES=health,auth,user,access,notifier,organization,contact

# Enable full ERP-lite
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice,realtime

# Enable with compliance
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice,verifactu
```

## Module Types

### Core Modules
- `health` - Health check endpoints
- `auth` - Authentication (login, register, JWT)
- `user` - User profile management
- `access` - Role-based access control
- `notifier` - Email notifications

### ERP-lite Modules
- `organization` - Multi-tenant organizations ✅
- `contact` - Customer/supplier management ✅
- `product` - Product catalog with variants and pricing ✅
- `invoice` - Invoice management with payments and statistics ✅
- `realtime` - WebSocket connections ✅
- `inventory` - Stock management (Planned)
- `calendar` - Appointment scheduling (Planned)

### Compliance Modules
- `verifactu` - Spanish tax compliance (RD 1007/2023 RRSIF) - Specified

## Builder Patterns

### Fluent Interface

```go
registry := modules.NewRegistry()
modules.RegisterBuiltinModules(registry)

moduleSet := modules.NewModuleSetBuilder(registry).
    WithCoreModules().
    WithERPModules().
    Build()
```

### Selective Building

```go
registry := modules.NewRegistry()
modules.RegisterBuiltinModules(registry)

moduleSet := modules.NewModuleSetBuilder(registry).
    WithModule("auth").
    WithModule("user").
    WithModule("organization").
    Build()
```

### Default Configuration

```go
// Equivalent to WithAllModules()
registry := modules.NewRegistry()
modules.RegisterBuiltinModules(registry)
moduleSet := modules.DefaultModuleSet(registry)
```

## Testing

The module system is fully testable:

```go
func TestCustomModuleSet(t *testing.T) {
    registry := modules.NewRegistry()
    modules.RegisterBuiltinModules(registry)

    moduleSet := modules.NewModuleSetBuilder(registry).
        WithModule("auth").
        Build()

    assert.True(t, moduleSet.IsRegistered("auth"))
    assert.False(t, moduleSet.IsRegistered("user"))
}
```

## CLI Integration

The module system is designed for CLI deconstructibility:

1. **Module Detection**: CLI can inspect which modules are registered
2. **Selective Copying**: Only copy files for selected modules
3. **Dependency Resolution**: Automatically include required dependencies
4. **Configuration Generation**: Generate appropriate .env files

## Migration from Old System

The old global `moduleSet` variable has been replaced with injectable `ModuleSet` instances. The `BuildModules()` function maintains backward compatibility while using the new system internally.

### Before
```go
// Hard-coded global registration
var moduleSet = NewModuleSet()
moduleSet.Register("auth", AuthModule) // Global state

// In main.go
modules.BuildModules(cfg.Modules, nil) // Uses global state
```

### After
```go
// Injectable module set
registry := modules.NewRegistry()
modules.RegisterBuiltinModules(registry)
moduleSet := modules.NewModuleSetBuilder(registry).
    WithCoreModules().
    Build()

// In main.go
fx.Supply(moduleSet) // Inject as dependency
moduleSet.Build([]string{}) // Use injected instance
```

## Benefits

1. **No Global Coupling**: Modules are not hard-coded into the system
2. **Flexible Composition**: Easy to create different module combinations
3. **Repository Deduplication**: Shared repositories are provided once
4. **Better Testing**: Module sets can be easily mocked
5. **CLI-Friendly**: Designed for selective module extraction
6. **Configuration-Driven**: Modules can be enabled/disabled via config