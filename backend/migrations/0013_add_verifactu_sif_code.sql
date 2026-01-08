-- +goose Up
ALTER TABLE IF EXISTS verifactu_records
    ADD COLUMN IF NOT EXISTS sif_code CHAR(2) NOT NULL;

-- +goose Down
ALTER TABLE IF EXISTS verifactu_records
    DROP COLUMN IF EXISTS sif_code;

