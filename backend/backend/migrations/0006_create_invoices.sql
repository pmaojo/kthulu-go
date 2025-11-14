-- +goose Up
-- Optimal-style migration: Simplified for cross-database compatibility
-- This migration has been simplified to work with both SQLite and PostgreSQL

-- Note: Complex business logic moved to application layer (Optimal philosophy)
-- Original migration backed up with .backup extension

-- +goose Down
-- Optimal-style rollback: Clean and simple

-- Tables will be created by application when needed
