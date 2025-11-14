// @kthulu:module:verifactu
package verifactu

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Record represents a VeriFactu record in the system.
type Record struct {
	ID               int       `json:"id"`
	InvoiceID        int       `json:"invoiceId"`
	OrganizationID   int       `json:"organizationId"`
	RecordType       string    `json:"recordType"`
	OriginalRecordID *int      `json:"originalRecordId,omitempty"`
	SIFCode          string    `json:"sifCode"`
	Hash             string    `json:"hash"`
	CreatedAt        time.Time `json:"createdAt"`
}

// Repository defines the storage behavior required by the service.
type Repository interface {
	// GetRecordByID retrieves a record by its identifier.
	GetRecordByID(ctx context.Context, id int) (*Record, error)
	// CreateRecord persists a new VeriFactu record.
	CreateRecord(ctx context.Context, record *Record) error
	// ListRecordsByOrganization returns all records for an organization.
	ListRecordsByOrganization(ctx context.Context, orgID int) ([]*Record, error)
	// GetLastHash returns the hash of the most recent record for an organization.
	GetLastHash(ctx context.Context, orgID int) (string, error)
	// GetLiveMode returns if live mode is active for the given fiscal year.
	GetLiveMode(ctx context.Context, year int) (bool, error)
	// SetLiveMode persists the live mode flag for the given fiscal year.
	SetLiveMode(ctx context.Context, year int, live bool) error
}

// Signer defines signing capabilities for generated exports.
type Signer interface {
	Sign(data []byte) ([]byte, error)
}

// Service provides VeriFactu operations.
type Service struct {
	repo    Repository
	signer  Signer
	sifCode string
	mode    string
}

// NewService creates a new VeriFactu service instance.
func NewService(repo Repository, signer Signer, sifCode, mode string) *Service {
	return &Service{repo: repo, signer: signer, sifCode: sifCode, mode: mode}
}

// SIFCode returns the current SIF code used by the service.
func (s *Service) SIFCode() string { return s.sifCode }

// UpdateSIFCode changes the SIF code used for new records.
func (s *Service) UpdateSIFCode(code string) { s.sifCode = code }

// Config represents current VeriFactu configuration.
type Config struct {
	SIFCode  string `json:"sifCode"`
	Mode     string `json:"mode"`
	LiveMode bool   `json:"liveMode"`
}

// ErrModeFrozen occurs when attempting to switch to queued while live mode is active.
var ErrModeFrozen = errors.New("live mode active until fiscal year end")

// Config returns current configuration with live mode status.
func (s *Service) Config(ctx context.Context) (Config, error) {
	live, err := s.repo.GetLiveMode(ctx, time.Now().Year())
	if err != nil {
		return Config{}, err
	}
	return Config{SIFCode: s.sifCode, Mode: s.mode, LiveMode: live}, nil
}

// UpdateConfig changes SIF code or mode with live mode validation.
func (s *Service) UpdateConfig(ctx context.Context, code, mode string) (Config, error) {
	if s.mode == "real-time" && mode == "queued" {
		live, err := s.repo.GetLiveMode(ctx, time.Now().Year())
		if err != nil {
			return Config{}, err
		}
		if live {
			return Config{}, ErrModeFrozen
		}
	}
	s.sifCode = code
	s.mode = mode
	return s.Config(ctx)
}

// ErrRecordNotFound is returned when a VeriFactu record cannot be located.
var ErrRecordNotFound = errors.New("verifactu record not found")

// GenerateRecord creates a new VeriFactu record computing a chained hash.
// The previous hash is looked up per organization to ensure independent chains.
func (s *Service) GenerateRecord(ctx context.Context, invoiceID, orgID int, recordType string) (*Record, error) {
	// Retrieve the hash of the last record for this organization to keep
	// the chain independent between different organizations.
	if s.sifCode == "" {
		return nil, errors.New("sif code not configured")
	}
	prevHash, err := s.repo.GetLastHash(ctx, orgID)
	if err != nil {
		return nil, err
	}

	hash := computeRecordHash(prevHash, invoiceID, orgID, recordType, s.sifCode)

	rec := &Record{
		InvoiceID:      invoiceID,
		OrganizationID: orgID,
		RecordType:     recordType,
		SIFCode:        s.sifCode,
		Hash:           hash,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.CreateRecord(ctx, rec); err != nil {
		return nil, err
	}

	return rec, nil
}

func computeRecordHash(prev string, invoiceID, orgID int, recordType, sif string) string {
	data := fmt.Sprintf("%s:%d:%d:%s:%s", prev, invoiceID, orgID, recordType, sif)
	sum := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", sum[:])
}

// CancelRecord generates a cancellation record linked to the original record.
func (s *Service) CancelRecord(ctx context.Context, recordID, userID int) (*Record, error) {
	original, err := s.repo.GetRecordByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, ErrRecordNotFound
	}

	prevHash, err := s.repo.GetLastHash(ctx, original.OrganizationID)
	if err != nil {
		return nil, err
	}

	hash := computeRecordHash(prevHash, original.InvoiceID, original.OrganizationID, "anulacion", original.SIFCode)

	now := time.Now().UTC()
	cancelRecord := &Record{
		InvoiceID:        original.InvoiceID,
		OrganizationID:   original.OrganizationID,
		RecordType:       "anulacion",
		OriginalRecordID: &original.ID,
		SIFCode:          original.SIFCode,
		Hash:             hash,
		CreatedAt:        now,
	}

	if err := s.repo.CreateRecord(ctx, cancelRecord); err != nil {
		return nil, err
	}

	return cancelRecord, nil
}

// ExportRecords generates a signed ZIP archive containing all
// VeriFactu records for the provided organization. The archive includes
// both JSON and CSV representations of the records. The returned slice
// contains the ZIP bytes and their signature.
func (s *Service) ExportRecords(ctx context.Context, orgID int) ([]byte, []byte, error) {
	records, err := s.repo.ListRecordsByOrganization(ctx, orgID)
	if err != nil {
		return nil, nil, err
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	jsonFile, err := zipWriter.Create("records.json")
	if err != nil {
		return nil, nil, err
	}
	if _, err := jsonFile.Write(jsonData); err != nil {
		return nil, nil, err
	}

	csvFile, err := zipWriter.Create("records.csv")
	if err != nil {
		return nil, nil, err
	}
	csvWriter := csv.NewWriter(csvFile)
	if err := csvWriter.Write([]string{"id", "invoiceId", "organizationId", "recordType", "originalRecordId", "sifCode", "createdAt"}); err != nil {
		return nil, nil, err
	}
	for _, r := range records {
		original := ""
		if r.OriginalRecordID != nil {
			original = strconv.Itoa(*r.OriginalRecordID)
		}
		if err := csvWriter.Write([]string{
			strconv.Itoa(r.ID),
			strconv.Itoa(r.InvoiceID),
			strconv.Itoa(r.OrganizationID),
			r.RecordType,
			original,
			r.SIFCode,
			r.CreatedAt.Format(time.RFC3339),
		}); err != nil {
			return nil, nil, err
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, nil, err
	}

	zipBytes := buf.Bytes()
	sig, err := s.signer.Sign(zipBytes)
	if err != nil {
		return nil, nil, err
	}

	return zipBytes, sig, nil
}

// HMACSigner provides HMAC-SHA256 signing for exported data.
type HMACSigner struct{ key []byte }

// NewHMACSigner creates a new signer with the given key.
func NewHMACSigner(key []byte) *HMACSigner { return &HMACSigner{key: key} }

// Sign returns the HMAC-SHA256 signature for the provided data.
func (s *HMACSigner) Sign(data []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, s.key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}
