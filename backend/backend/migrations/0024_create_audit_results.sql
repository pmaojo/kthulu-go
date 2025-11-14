-- +goose Up
-- Create audit_results table for code audit results

CREATE TABLE audit_results (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL,
    duration TEXT NOT NULL,
    counts TEXT, -- JSON object
    findings TEXT, -- JSON array
    strict INTEGER DEFAULT 0, -- boolean as integer
    warnings TEXT, -- JSON array
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Create indexes for audit_results table
CREATE INDEX idx_audit_results_path ON audit_results(path);
CREATE INDEX idx_audit_results_created_at ON audit_results(created_at);

-- +goose Down
DROP TABLE IF EXISTS audit_results;