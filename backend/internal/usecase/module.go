// @kthulu:module:modules
package usecase

import (
	"context"
	"fmt"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/domain"
	"github.com/pmaojo/kthulu-go/backend/internal/domain/repository"
)

// ModuleUseCase orchestrates module management workflows.
type ModuleUseCase struct {
	modules repository.ModuleRepository
	logger  core.Logger
}

// NewModuleUseCase creates a new module use case.
func NewModuleUseCase(
	modules repository.ModuleRepository,
	logger core.Logger,
) *ModuleUseCase {
	return &ModuleUseCase{
		modules: modules,
		logger:  logger,
	}
}

// ListModules retrieves available modules, optionally filtered by category.
func (m *ModuleUseCase) ListModules(ctx context.Context, category *string, limit, offset int) ([]*domain.ModuleInfo, error) {
	ctx, span := startUseCaseSpan(ctx, "ModuleUseCase.ListModules")
	defer span.End()

	m.logger.Info("Listing modules", "category", category, "limit", limit, "offset", offset)

	var modules []*domain.ModuleInfo
	var err error

	if category != nil && *category != "" {
		modules, err = m.modules.ListByCategory(ctx, *category, limit, offset)
	} else {
		modules, err = m.modules.List(ctx, limit, offset)
	}

	if err != nil {
		m.logger.Error("Failed to list modules", "error", err)
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	m.logger.Info("Modules listed", "count", len(modules))
	return modules, nil
}

// GetModule retrieves a specific module by name.
func (m *ModuleUseCase) GetModule(ctx context.Context, name string) (*domain.ModuleInfo, error) {
	ctx, span := startUseCaseSpan(ctx, "ModuleUseCase.GetModule")
	defer span.End()

	m.logger.Info("Getting module", "name", name)

	module, err := m.modules.FindByName(ctx, name)
	if err != nil {
		if err == domain.ErrModuleNotFound {
			m.logger.Warn("Module not found", "name", name)
			return nil, err
		}
		m.logger.Error("Failed to find module", "name", name, "error", err)
		return nil, fmt.Errorf("failed to find module: %w", err)
	}

	m.logger.Info("Module retrieved", "name", name)
	return module, nil
}

// ValidateModules validates a set of modules for compatibility and dependencies.
func (m *ModuleUseCase) ValidateModules(ctx context.Context, moduleNames []string) (*domain.ModuleValidationResult, error) {
	ctx, span := startUseCaseSpan(ctx, "ModuleUseCase.ValidateModules")
	defer span.End()

	m.logger.Info("Validating modules", "modules", moduleNames)

	result := &domain.ModuleValidationResult{
		Valid: true,
	}

	// Get all requested modules
	modules := make(map[string]*domain.ModuleInfo)
	for _, name := range moduleNames {
		module, err := m.modules.FindByName(ctx, name)
		if err != nil {
			if err == domain.ErrModuleNotFound {
				result.Valid = false
				result.Missing = append(result.Missing, name)
			} else {
				m.logger.Error("Failed to find module for validation", "name", name, "error", err)
				return nil, fmt.Errorf("failed to find module %s: %w", name, err)
			}
		} else {
			modules[name] = module
		}
	}

	// Check for circular dependencies (simplified)
	// In a real implementation, this would use graph algorithms

	// Check for conflicts
	for _, module := range modules {
		for _, requestedName := range moduleNames {
			if module.ConflictsWith(requestedName) {
				result.Valid = false
				result.Conflicts = append(result.Conflicts, domain.ModuleConflict{
					Module:    module.Name,
					Conflicts: []string{requestedName},
					Reason:    "Module conflict detected",
				})
			}
		}
	}

	m.logger.Info("Modules validated", "valid", result.Valid)
	return result, nil
}

// PlanModules creates an injection plan for the given modules.
func (m *ModuleUseCase) PlanModules(ctx context.Context, moduleNames []string) (*domain.ModuleInjectionPlan, error) {
	ctx, span := startUseCaseSpan(ctx, "ModuleUseCase.PlanModules")
	defer span.End()

	m.logger.Info("Planning module injection", "modules", moduleNames)

	plan := &domain.ModuleInjectionPlan{
		RequestedModules: moduleNames,
		ResolvedModules:  []string{},
		InjectedModules:  []string{},
		ExecutionOrder:   []string{},
		ModuleDetails:    make(map[string]domain.ModuleInfo),
	}

	// Validate modules first
	validation, err := m.ValidateModules(ctx, moduleNames)
	if err != nil {
		m.logger.Error("Failed to validate modules for planning", "error", err)
		return nil, fmt.Errorf("failed to validate modules: %w", err)
	}

	if !validation.Valid {
		plan.Errors = append(plan.Errors, "Module validation failed")
		for _, missing := range validation.Missing {
			plan.Errors = append(plan.Errors, "Missing module: "+missing)
		}
		for _, conflict := range validation.Conflicts {
			plan.Errors = append(plan.Errors, "Conflict: "+conflict.Module+" conflicts with "+fmt.Sprintf("%v", conflict.Conflicts))
		}
		return plan, nil
	}

	// Resolve dependencies and create execution order
	// Simplified: just use the requested order
	plan.ResolvedModules = moduleNames
	plan.InjectedModules = moduleNames
	plan.ExecutionOrder = moduleNames

	// Add module details
	for _, name := range moduleNames {
		if module, err := m.modules.FindByName(ctx, name); err == nil {
			plan.ModuleDetails[name] = *module
		}
	}

	m.logger.Info("Module injection planned", "modules", len(plan.InjectedModules))
	return plan, nil
}
