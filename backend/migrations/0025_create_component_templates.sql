-- +goose Up
-- Create component_templates table for component generation templates

CREATE TABLE component_templates (
    id INTEGER PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    language TEXT NOT NULL,
    framework TEXT,
    template TEXT NOT NULL, -- template content
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for component_templates table
CREATE UNIQUE INDEX idx_component_templates_type_name ON component_templates(type, name);
CREATE INDEX idx_component_templates_language ON component_templates(language);
CREATE INDEX idx_component_templates_framework ON component_templates(framework);

-- +goose Down
DROP TABLE IF EXISTS component_templates;