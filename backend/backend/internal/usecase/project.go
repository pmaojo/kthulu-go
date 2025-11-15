// @kthulu:module:projects
package usecase

import (
	"context"
	"fmt"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/domain"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// ProjectUseCase orchestrates project management workflows.
type ProjectUseCase struct {
	projects repository.ProjectRepository
	logger   core.Logger
}

// NewProjectUseCase creates a new project use case.
func NewProjectUseCase(
	projects repository.ProjectRepository,
	logger core.Logger,
) *ProjectUseCase {
	return &ProjectUseCase{
		projects: projects,
		logger:   logger,
	}
}

// PlanProject creates a project plan based on the request.
func (p *ProjectUseCase) PlanProject(ctx context.Context, req domain.ProjectRequest) (*domain.ProjectPlan, error) {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.PlanProject")
	defer span.End()

	p.logger.Info("Planning project", "name", req.Name)

	// Create project structure
	structure := &domain.ProjectStructure{
		Name: req.Name,
		Path: req.Path,
	}

	if req.Template != "" {
		structure.Backend = &domain.BackendConfig{
			Template: req.Template,
		}
	}

	if req.Database != "" {
		structure.Database = &domain.DatabaseConfig{
			Type: req.Database,
		}
	}

	if req.Frontend != "" {
		structure.Frontend = &domain.FrontendConfig{
			Template: req.Frontend,
		}
	}

	// Create project plan
	plan := &domain.ProjectPlan{
		Options:            req,
		Structure:          *structure,
		Modules:            req.Modules,
		ProjectDirectories: []string{"src", "tests", "docs"},
	}

	if req.Template != "" {
		plan.BackendTemplate = req.Template
	}

	if req.Frontend != "" {
		plan.FrontendTemplate = req.Frontend
	}

	p.logger.Info("Project plan created", "name", req.Name)
	return plan, nil
}

// GenerateProject creates and persists a new project.
func (p *ProjectUseCase) GenerateProject(ctx context.Context, req domain.ProjectRequest) (*domain.Project, error) {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.GenerateProject")
	defer span.End()

	p.logger.Info("Generating project", "name", req.Name)

	// Check if project already exists
	exists, err := p.projects.ExistsByName(ctx, req.Name)
	if err != nil {
		p.logger.Error("Failed to check project existence", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}
	if exists {
		p.logger.Warn("Project already exists", "name", req.Name)
		return nil, fmt.Errorf("project with name %s already exists", req.Name)
	}

	// Create project domain object
	project, err := domain.NewProject(req)
	if err != nil {
		p.logger.Error("Failed to create project domain object", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Persist project
	if err := p.projects.Create(ctx, project); err != nil {
		p.logger.Error("Failed to persist project", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	p.logger.Info("Project generated successfully", "name", req.Name, "id", project.ID)
	return project, nil
}

// GetProject retrieves a project by ID.
func (p *ProjectUseCase) GetProject(ctx context.Context, id uint) (*domain.Project, error) {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.GetProject")
	defer span.End()

	p.logger.Info("Getting project", "id", id)

	project, err := p.projects.FindByID(ctx, id)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			p.logger.Warn("Project not found", "id", id)
			return nil, err
		}
		p.logger.Error("Failed to find project", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	p.logger.Info("Project retrieved", "id", id)
	return project, nil
}

// UpdateProject updates an existing project.
func (p *ProjectUseCase) UpdateProject(ctx context.Context, id uint, req domain.ProjectRequest) (*domain.Project, error) {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.UpdateProject")
	defer span.End()

	p.logger.Info("Updating project", "id", id)

	// Find existing project
	project, err := p.projects.FindByID(ctx, id)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			p.logger.Warn("Project not found for update", "id", id)
			return nil, err
		}
		p.logger.Error("Failed to find project for update", "id", id, "error", err)
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// Update project
	project.Update(req)

	// Persist changes
	if err := p.projects.Update(ctx, project); err != nil {
		p.logger.Error("Failed to update project", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	p.logger.Info("Project updated successfully", "id", id)
	return project, nil
}

// DeleteProject removes a project.
func (p *ProjectUseCase) DeleteProject(ctx context.Context, id uint) error {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.DeleteProject")
	defer span.End()

	p.logger.Info("Deleting project", "id", id)

	// Check if project exists
	exists, err := p.projects.ExistsByID(ctx, id)
	if err != nil {
		p.logger.Error("Failed to check project existence", "id", id, "error", err)
		return fmt.Errorf("failed to check project existence: %w", err)
	}
	if !exists {
		p.logger.Warn("Project not found for deletion", "id", id)
		return domain.ErrProjectNotFound
	}

	// Delete project
	if err := p.projects.Delete(ctx, id); err != nil {
		p.logger.Error("Failed to delete project", "id", id, "error", err)
		return fmt.Errorf("failed to delete project: %w", err)
	}

	p.logger.Info("Project deleted successfully", "id", id)
	return nil
}

// ListProjects retrieves projects with pagination.
func (p *ProjectUseCase) ListProjects(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	ctx, span := startUseCaseSpan(ctx, "ProjectUseCase.ListProjects")
	defer span.End()

	p.logger.Info("Listing projects", "limit", limit, "offset", offset)

	projects, err := p.projects.List(ctx, limit, offset)
	if err != nil {
		p.logger.Error("Failed to list projects", "error", err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	p.logger.Info("Projects listed", "count", len(projects))
	return projects, nil
}
