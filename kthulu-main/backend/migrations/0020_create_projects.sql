-- +goose Up
-- Create projects table for code generation projects

CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    modules TEXT, -- JSON array
    template TEXT,
    database TEXT,
    frontend TEXT,
    skip_git INTEGER DEFAULT 0, -- boolean as integer
    skip_docker INTEGER DEFAULT 0, -- boolean as integer
    author TEXT,
    license TEXT,
    description TEXT,
    path TEXT,
    dry_run INTEGER DEFAULT 0, -- boolean as integer
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for projects table
CREATE UNIQUE INDEX idx_projects_name ON projects(name);
CREATE INDEX idx_projects_created_at ON projects(created_at);

-- +goose Down
DROP TABLE IF EXISTS projects;