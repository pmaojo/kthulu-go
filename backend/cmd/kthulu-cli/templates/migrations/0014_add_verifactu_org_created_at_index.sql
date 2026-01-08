-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS verifactu_records_org_created_at_idx
    ON verifactu_records (organization_id, created_at);

-- +goose Down
DROP INDEX IF EXISTS verifactu_records_org_created_at_idx;

