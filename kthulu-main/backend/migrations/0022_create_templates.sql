-- +goose Up
-- Create templates table for code templates

CREATE TABLE templates (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT,
    description TEXT,
    author TEXT,
    category TEXT,
    tags TEXT, -- JSON array
    content TEXT, -- JSON object (file path -> content)
    remote INTEGER DEFAULT 0, -- boolean as integer
    url TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for templates table
CREATE UNIQUE INDEX idx_templates_name ON templates(name);
CREATE INDEX idx_templates_category ON templates(category);
CREATE INDEX idx_templates_author ON templates(author);
CREATE INDEX idx_templates_remote ON templates(remote);

-- +goose Down
DROP TABLE IF EXISTS templates;