-- +goose Up
-- Create modules table for code module catalog

CREATE TABLE modules (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    version TEXT,
    dependencies TEXT, -- JSON array
    optional INTEGER DEFAULT 0, -- boolean as integer
    category TEXT,
    tags TEXT, -- JSON array
    entities TEXT, -- JSON array
    routes TEXT, -- JSON array
    migrations TEXT, -- JSON array
    frontend INTEGER DEFAULT 0, -- boolean as integer
    backend INTEGER DEFAULT 1, -- boolean as integer
    config TEXT, -- JSON object
    conflicts TEXT, -- JSON array
    min_version TEXT,
    max_version TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for modules table
CREATE UNIQUE INDEX idx_modules_name ON modules(name);
CREATE INDEX idx_modules_category ON modules(category);
CREATE INDEX idx_modules_frontend ON modules(frontend);
CREATE INDEX idx_modules_backend ON modules(backend);

-- +goose Down
DROP TABLE IF EXISTS modules;