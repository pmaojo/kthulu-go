// @kthulu:module:verifactu
package db

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/modules/verifactu"
)

// VerifactuRepository implements verifactu.Repository using database/sql.
type VerifactuRepository struct {
	db *sql.DB
}

// NewVerifactuRepository creates a new repository instance.
func NewVerifactuRepository(db *sql.DB) verifactu.Repository {
	return &VerifactuRepository{db: db}
}

// GetRecordByID retrieves a record by its ID.
func (r *VerifactuRepository) GetRecordByID(ctx context.Context, id int) (*verifactu.Record, error) {
	const query = `SELECT id, invoice_id, organization_id, record_type, original_record_id, sif_code, hash, created_at FROM verifactu_records WHERE id = $1`
	rec := &verifactu.Record{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&rec.ID, &rec.InvoiceID, &rec.OrganizationID, &rec.RecordType, &rec.OriginalRecordID, &rec.SIFCode, &rec.Hash, &rec.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get verifactu record: %w", err)
	}
	return rec, nil
}

// CreateRecord inserts a new VeriFactu record.
func (r *VerifactuRepository) CreateRecord(ctx context.Context, record *verifactu.Record) error {
	const query = `INSERT INTO verifactu_records (invoice_id, organization_id, record_type, original_record_id, sif_code, hash, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	return r.db.QueryRowContext(ctx, query, record.InvoiceID, record.OrganizationID, record.RecordType, record.OriginalRecordID, record.SIFCode, record.Hash, record.CreatedAt).Scan(&record.ID)
}

// ListRecordsByOrganization returns all records for the given organization.
func (r *VerifactuRepository) ListRecordsByOrganization(ctx context.Context, orgID int) ([]*verifactu.Record, error) {
	const query = `SELECT id, invoice_id, organization_id, record_type, original_record_id, sif_code, hash, created_at FROM verifactu_records WHERE organization_id = $1 ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list verifactu records: %w", err)
	}
	defer rows.Close()

	var records []*verifactu.Record
	for rows.Next() {
		rec := &verifactu.Record{}
		if err := rows.Scan(&rec.ID, &rec.InvoiceID, &rec.OrganizationID, &rec.RecordType, &rec.OriginalRecordID, &rec.SIFCode, &rec.Hash, &rec.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan verifactu record: %w", err)
		}
		records = append(records, rec)
	}
	return records, nil
}

// GetLastHash returns the hash of the most recent record for the organization.
func (r *VerifactuRepository) GetLastHash(ctx context.Context, orgID int) (string, error) {
	// Select the most recent hash for the given organization. Ordering by
	// created_at allows using the composite index on (organization_id,
	// created_at) for efficient lookups.
	const query = `SELECT hash FROM verifactu_records WHERE organization_id = $1 ORDER BY created_at DESC LIMIT 1`
	var h sql.NullString
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&h)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get last verifactu hash: %w", err)
	}
	return h.String, nil
}

// GetLiveMode returns whether live mode is active for the given fiscal year.
func (r *VerifactuRepository) GetLiveMode(ctx context.Context, year int) (bool, error) {
	const query = `SELECT live_mode FROM verifactu_settings WHERE fiscal_year = $1`
	var live sql.NullBool
	err := r.db.QueryRowContext(ctx, query, year).Scan(&live)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get verifactu live mode: %w", err)
	}
	return live.Bool, nil
}

// SetLiveMode persists the live mode flag for the given fiscal year.
func (r *VerifactuRepository) SetLiveMode(ctx context.Context, year int, live bool) error {
	const query = `INSERT INTO verifactu_settings (fiscal_year, live_mode) VALUES ($1, $2)
ON CONFLICT (fiscal_year) DO UPDATE SET live_mode = EXCLUDED.live_mode`
	if _, err := r.db.ExecContext(ctx, query, year, live); err != nil {
		return fmt.Errorf("set verifactu live mode: %w", err)
	}
	return nil
}
