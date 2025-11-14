-- +goose Up
-- Create oauth_sessions table
CREATE TABLE oauth_sessions (
    signature TEXT PRIMARY KEY,
    request TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS oauth_sessions;

