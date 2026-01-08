-- +goose Up
ALTER TABLE users ADD COLUMN confirmation_code TEXT;
CREATE INDEX idx_users_confirmation_code ON users(confirmation_code);

-- +goose Down
DROP INDEX IF EXISTS idx_users_confirmation_code;
ALTER TABLE users DROP COLUMN confirmation_code;
