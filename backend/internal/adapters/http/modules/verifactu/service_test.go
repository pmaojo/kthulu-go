package verifactu

import (
	"context"
	"testing"
	"time"
)

type memRepo struct {
	records   map[int][]*Record
	liveModes map[int]bool
}

func newMemRepo() *memRepo {
	return &memRepo{records: make(map[int][]*Record), liveModes: make(map[int]bool)}
}

func (m *memRepo) GetRecordByID(ctx context.Context, id int) (*Record, error) {
	for _, recs := range m.records {
		for _, r := range recs {
			if r.ID == id {
				return r, nil
			}
		}
	}
	return nil, nil
}

func (m *memRepo) CreateRecord(ctx context.Context, record *Record) error {
	recs := m.records[record.OrganizationID]
	record.ID = len(recs) + 1
	m.records[record.OrganizationID] = append(recs, record)
	return nil
}

func (m *memRepo) ListRecordsByOrganization(ctx context.Context, orgID int) ([]*Record, error) {
	return m.records[orgID], nil
}

func (m *memRepo) GetLastHash(ctx context.Context, orgID int) (string, error) {
	recs := m.records[orgID]
	if len(recs) == 0 {
		return "", nil
	}
	return recs[len(recs)-1].Hash, nil
}

func (m *memRepo) GetLiveMode(ctx context.Context, year int) (bool, error) {
	return m.liveModes[year], nil
}

func (m *memRepo) SetLiveMode(ctx context.Context, year int, live bool) error {
	m.liveModes[year] = live
	return nil
}

func TestUpdateConfigLiveMode(t *testing.T) {
	repo := newMemRepo()
	svc := NewService(repo, NewHMACSigner([]byte("key")), "AA", "real-time")
	ctx := context.Background()
	_ = repo.SetLiveMode(ctx, time.Now().Year(), true)
	if _, err := svc.UpdateConfig(ctx, "AA", "queued"); err != ErrModeFrozen {
		t.Fatalf("expected mode frozen error")
	}
}

func TestGenerateRecordIndependentChains(t *testing.T) {
	repo := newMemRepo()
	svc := NewService(repo, NewHMACSigner([]byte("key")), "AA", "queued")
	ctx := context.Background()

	r1, err := svc.GenerateRecord(ctx, 1, 1, "alta")
	if err != nil {
		t.Fatalf("generate record org1: %v", err)
	}
	if r1.Hash != computeRecordHash("", 1, 1, "alta", "AA") {
		t.Fatalf("unexpected hash for first record: %s", r1.Hash)
	}

	r2, err := svc.GenerateRecord(ctx, 2, 1, "alta")
	if err != nil {
		t.Fatalf("generate second record org1: %v", err)
	}
	if r2.Hash != computeRecordHash(r1.Hash, 2, 1, "alta", "AA") {
		t.Fatalf("unexpected hash for second record: %s", r2.Hash)
	}

	r3, err := svc.GenerateRecord(ctx, 3, 2, "alta")
	if err != nil {
		t.Fatalf("generate record org2: %v", err)
	}
	if r3.Hash != computeRecordHash("", 3, 2, "alta", "AA") {
		t.Fatalf("org2 chain should start fresh: %s", r3.Hash)
	}

	r4, err := svc.GenerateRecord(ctx, 4, 2, "alta")
	if err != nil {
		t.Fatalf("generate second record org2: %v", err)
	}
	if r4.Hash != computeRecordHash(r3.Hash, 4, 2, "alta", "AA") {
		t.Fatalf("unexpected hash for second org2 record: %s", r4.Hash)
	}
}

func TestGenerateRecordRequiresSIFCode(t *testing.T) {
	repo := newMemRepo()
	svc := NewService(repo, NewHMACSigner([]byte("key")), "", "queued")
	ctx := context.Background()
	if _, err := svc.GenerateRecord(ctx, 1, 1, "alta"); err == nil {
		t.Fatalf("expected error when SIF code is missing")
	}
}

func TestCancelRecordComputesHash(t *testing.T) {
	repo := newMemRepo()
	svc := NewService(repo, NewHMACSigner([]byte("key")), "AA", "queued")
	ctx := context.Background()

	original, err := svc.GenerateRecord(ctx, 1, 1, "alta")
	if err != nil {
		t.Fatalf("generate original record: %v", err)
	}
	follow, err := svc.GenerateRecord(ctx, 2, 1, "alta")
	if err != nil {
		t.Fatalf("generate follow record: %v", err)
	}

	cancel, err := svc.CancelRecord(ctx, original.ID, 1)
	if err != nil {
		t.Fatalf("cancel record: %v", err)
	}

	expected := computeRecordHash(follow.Hash, original.InvoiceID, original.OrganizationID, "anulacion", original.SIFCode)
	if cancel.Hash != expected {
		t.Fatalf("unexpected cancel hash: %s", cancel.Hash)
	}
	if cancel.OriginalRecordID == nil || *cancel.OriginalRecordID != original.ID {
		t.Fatalf("cancel record should reference original")
	}
}
