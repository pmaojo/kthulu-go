-- +goose Up
-- Create table for VeriFactu records with cancellation support
CREATE TABLE verifactu_records (
    id INTEGER PRIMARY KEY,
    invoice_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    record_type TEXT NOT NULL,
    original_record_id INTEGER,
    hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (original_record_id) REFERENCES verifactu_records(id)
);

-- +goose Down
-- Drop VeriFactu records table
DROP TABLE IF EXISTS verifactu_records;

