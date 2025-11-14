-- +goose Up
CREATE TABLE IF NOT EXISTS verifactu_settings (
    fiscal_year INT PRIMARY KEY,
    live_mode BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS verifactu_settings;
