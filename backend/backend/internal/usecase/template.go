// @kthulu:module:templates
package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	texttemplate "text/template"

	"backend/core"
	"backend/internal/domain"
	"backend/internal/repository"
)

// TemplateUseCase orchestrates template management workflows.
type TemplateUseCase struct {
	templates  repository.TemplateRepository
	registries repository.TemplateRegistryRepository
	logger     core.Logger
}

// NewTemplateUseCase creates a new template use case.
func NewTemplateUseCase(
	templates repository.TemplateRepository,
	registries repository.TemplateRegistryRepository,
	logger core.Logger,
) *TemplateUseCase {
	return &TemplateUseCase{
		templates:  templates,
		registries: registries,
		logger:     logger,
	}
}

// ListTemplates retrieves available templates.
func (t *TemplateUseCase) ListTemplates(ctx context.Context, limit, offset int) ([]*domain.Template, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.ListTemplates")
	defer span.End()

	t.logger.Info("Listing templates", "limit", limit, "offset", offset)

	templates, err := t.templates.List(ctx, limit, offset)
	if err != nil {
		t.logger.Error("Failed to list templates", "error", err)
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	t.logger.Info("Templates listed", "count", len(templates))
	return templates, nil
}

// GetTemplate retrieves a specific template by name.
func (t *TemplateUseCase) GetTemplate(ctx context.Context, name string) (*domain.Template, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.GetTemplate")
	defer span.End()

	t.logger.Info("Getting template", "name", name)

	template, err := t.templates.FindByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to find template", "name", name, "error", err)
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	t.logger.Info("Template retrieved", "name", name)
	return template, nil
}

// CreateTemplate creates a new template.
func (t *TemplateUseCase) CreateTemplate(ctx context.Context, req domain.Template) (*domain.Template, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.CreateTemplate")
	defer span.End()

	t.logger.Info("Creating template", "name", req.Name)

	// Check if template already exists
	exists, err := t.templates.ExistsByName(ctx, req.Name)
	if err != nil {
		t.logger.Error("Failed to check template existence", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to check template existence: %w", err)
	}
	if exists {
		t.logger.Warn("Template already exists", "name", req.Name)
		return nil, fmt.Errorf("template with name %s already exists", req.Name)
	}

	template, err := domain.NewTemplate(req.Name)
	if err != nil {
		t.logger.Error("Failed to create template domain object", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Apply updates
	template.Update(req)

	// Persist template
	if err := t.templates.Create(ctx, template); err != nil {
		t.logger.Error("Failed to persist template", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	t.logger.Info("Template created successfully", "name", req.Name, "id", template.ID)
	return template, nil
}

// UpdateTemplate updates an existing template.
func (t *TemplateUseCase) UpdateTemplate(ctx context.Context, name string, updates domain.Template) (*domain.Template, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.UpdateTemplate")
	defer span.End()

	t.logger.Info("Updating template", "name", name)

	// Find existing template
	template, err := t.templates.FindByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to find template for update", "name", name, "error", err)
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Update template
	template.Update(updates)

	// Persist changes
	if err := t.templates.Update(ctx, template); err != nil {
		t.logger.Error("Failed to update template", "name", name, "error", err)
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	t.logger.Info("Template updated successfully", "name", name)
	return template, nil
}

// DeleteTemplate removes a template.
func (t *TemplateUseCase) DeleteTemplate(ctx context.Context, name string) error {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.DeleteTemplate")
	defer span.End()

	t.logger.Info("Deleting template", "name", name)

	// Find template to get ID
	template, err := t.templates.FindByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to find template for deletion", "name", name, "error", err)
		return fmt.Errorf("failed to find template: %w", err)
	}

	// Delete template
	if err := t.templates.Delete(ctx, template.ID); err != nil {
		t.logger.Error("Failed to delete template", "name", name, "error", err)
		return fmt.Errorf("failed to delete template: %w", err)
	}

	t.logger.Info("Template deleted successfully", "name", name)
	return nil
}

// RenderTemplate renders a template with the given variables.
func (t *TemplateUseCase) RenderTemplate(ctx context.Context, req domain.TemplateRenderRequest) (*domain.TemplateRenderResult, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.RenderTemplate")
	defer span.End()

	t.logger.Info("Rendering template", "name", req.Name)

	// Get template
	template, err := t.templates.FindByName(ctx, req.Name)
	if err != nil {
		t.logger.Error("Failed to find template for rendering", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Render template
	result, err := t.renderTemplate(template, req.Vars)
	if err != nil {
		t.logger.Error("Failed to render template", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	t.logger.Info("Template rendered successfully", "name", req.Name)
	return result, nil
}

// renderTemplate performs the actual template rendering
func (t *TemplateUseCase) renderTemplate(template *domain.Template, vars map[string]interface{}) (*domain.TemplateRenderResult, error) {
	result := &domain.TemplateRenderResult{
		Files: make(map[string]string),
	}

	// For each file in the template content
	for filePath, content := range template.GetAllFiles() {
		// Parse template
		tmpl, err := texttemplate.New(filePath).Parse(content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", filePath, err)
		}

		// Execute template
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, vars); err != nil {
			return nil, fmt.Errorf("failed to execute template %s: %w", filePath, err)
		}

		// Store rendered content (base64 encode as per API spec)
		result.Files[filePath] = buf.String() // Note: should base64 encode
	}

	return result, nil
}

// ValidateTemplate validates a template.
func (t *TemplateUseCase) ValidateTemplate(ctx context.Context, name string) error {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.ValidateTemplate")
	defer span.End()

	t.logger.Info("Validating template", "name", name)

	template, err := t.templates.FindByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to find template for validation", "name", name, "error", err)
		return fmt.Errorf("failed to find template: %w", err)
	}

	// Basic validation - check if content is valid JSON
	if template.Content != nil {
		if _, err := json.Marshal(template.Content); err != nil {
			t.logger.Error("Template content is not valid", "name", name, "error", err)
			return fmt.Errorf("template content is not valid: %w", err)
		}
	}

	t.logger.Info("Template validated successfully", "name", name)
	return nil
}

// AddRegistry adds a new template registry.
func (t *TemplateUseCase) AddRegistry(ctx context.Context, name, url string) (*domain.TemplateRegistry, error) {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.AddRegistry")
	defer span.End()

	t.logger.Info("Adding template registry", "name", name, "url", url)

	// Check if registry already exists
	exists, err := t.registries.ExistsByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to check registry existence", "name", name, "error", err)
		return nil, fmt.Errorf("failed to check registry existence: %w", err)
	}
	if exists {
		t.logger.Warn("Registry already exists", "name", name)
		return nil, fmt.Errorf("registry with name %s already exists", name)
	}

	registry, err := domain.NewTemplateRegistry(name, url)
	if err != nil {
		t.logger.Error("Failed to create registry domain object", "name", name, "error", err)
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	// Persist registry
	if err := t.registries.Create(ctx, registry); err != nil {
		t.logger.Error("Failed to persist registry", "name", name, "error", err)
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	t.logger.Info("Registry added successfully", "name", name, "id", registry.ID)
	return registry, nil
}

// RemoveRegistry removes a template registry.
func (t *TemplateUseCase) RemoveRegistry(ctx context.Context, name string) error {
	ctx, span := startUseCaseSpan(ctx, "TemplateUseCase.RemoveRegistry")
	defer span.End()

	t.logger.Info("Removing template registry", "name", name)

	// Find registry to get ID
	registry, err := t.registries.FindByName(ctx, name)
	if err != nil {
		t.logger.Error("Failed to find registry for removal", "name", name, "error", err)
		return fmt.Errorf("failed to find registry: %w", err)
	}

	// Delete registry
	if err := t.registries.Delete(ctx, registry.ID); err != nil {
		t.logger.Error("Failed to delete registry", "name", name, "error", err)
		return fmt.Errorf("failed to delete registry: %w", err)
	}

	t.logger.Info("Registry removed successfully", "name", name)
	return nil
}
