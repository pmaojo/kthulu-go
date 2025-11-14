-- +goose Up
-- Create oauth_tokens table
CREATE TABLE oauth_tokens (
    signature TEXT PRIMARY KEY,
    request TEXT,
    kind TEXT NOT NULL DEFAULT 'token',
    expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_oauth_tokens_kind ON oauth_tokens(kind);
CREATE INDEX idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);

-- +goose Down
DROP TABLE IF EXISTS oauth_tokens;

