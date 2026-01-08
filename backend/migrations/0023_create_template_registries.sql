-- +goose Up
-- Create template_registries table for template sources

CREATE TABLE template_registries (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for template_registries table
CREATE UNIQUE INDEX idx_template_registries_name ON template_registries(name);
CREATE UNIQUE INDEX idx_template_registries_url ON template_registries(url);

-- +goose Down
DROP TABLE IF EXISTS template_registries;