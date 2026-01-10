-- +goose Up
-- Create oauth_clients table
CREATE TABLE oauth_clients (
    id TEXT PRIMARY KEY,
    secret TEXT NOT NULL,
    redirect_uris TEXT NOT NULL,
    scopes TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE IF EXISTS oauth_clients;

