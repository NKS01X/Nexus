package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

// mockAuditRepo implements repository.AuditRepository for testing.
type mockAuditRepo struct {
	entries     []*model.AuditEntry
	latestIndex int64
	verifyOK    bool
}

func newMockAuditRepo() *mockAuditRepo {
	return &mockAuditRepo{
		entries:     []*model.AuditEntry{},
		latestIndex: 0,
		verifyOK:    true,
	}
}

func (m *mockAuditRepo) Append(ctx context.Context, entry *model.AuditEntry) error {
	m.latestIndex++
	entry.Index = m.latestIndex

	if m.latestIndex > 1 {
		entry.PrevHash = m.entries[m.latestIndex-2].Hash
	}
	entry.Hash = "hash_" + string(rune(m.latestIndex))
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditRepo) GetByIndex(ctx context.Context, index int64) (*model.AuditEntry, error) {
	if index <= 0 || index > int64(len(m.entries)) {
		return nil, nil
	}
	return m.entries[index-1], nil
}

func (m *mockAuditRepo) GetRange(ctx context.Context, from, to int64) ([]*model.AuditEntry, error) {
	if from < 1 {
		from = 1
	}
	if to > int64(len(m.entries)) {
		to = int64(len(m.entries))
	}
	if from > to {
		return []*model.AuditEntry{}, nil
	}
	return m.entries[from-1 : to], nil
}

func (m *mockAuditRepo) GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.AuditEntry, error) {
	var result []*model.AuditEntry
	for i := len(m.entries) - 1; i >= 0 && len(result) < limit; i-- {
		if m.entries[i].BuyerID == buyerID {
			result = append(result, m.entries[i])
		}
	}
	return result, nil
}

func (m *mockAuditRepo) GetByTraceID(ctx context.Context, traceID string) ([]*model.AuditEntry, error) {
	var result []*model.AuditEntry
	for _, e := range m.entries {
		if e.TraceID == traceID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockAuditRepo) VerifyChain(ctx context.Context, from, to int64) (bool, error) {
	return m.verifyOK, nil
}

func (m *mockAuditRepo) GetLatestIndex(ctx context.Context) (int64, error) {
	return m.latestIndex, nil
}

func (m *mockAuditRepo) GetMinIndex(ctx context.Context) (int64, error) {
	if len(m.entries) == 0 {
		return 0, nil
	}
	return 1, nil
}

func (m *mockAuditRepo) GetAll(ctx context.Context, limit int) ([]*model.AuditEntry, error) {
	if limit <= 0 || limit > len(m.entries) {
		limit = len(m.entries)
	}

	result := make([]*model.AuditEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = m.entries[len(m.entries)-1-i]
	}
	return result, nil
}

// TestAuditService_Log tests logging audit entries.
func TestAuditService_Log(t *testing.T) {
	repo := newMockAuditRepo()
	service := NewAuditService(repo)

	entry := &model.AuditEntry{
		Timestamp: time.Now(),
		TraceID:   "trace_123",
		BuyerID:   "buyer_1",
		Action:    model.AuditActionPurchaseAttempt,
		Request:   json.RawMessage(`{"sku": "SHOE-001"}`),
		Response:  json.RawMessage(`{"allowed": true}`),
	}

	err := service.Log(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.Index != 1 {
		t.Errorf("expected index=1, got %d", entry.Index)
	}
	if entry.PrevHash != "" {
		t.Errorf("expected empty prev_hash for first entry, got %s", entry.PrevHash)
	}
	if entry.Hash == "" {
		t.Error("expected hash to be set")
	}

	entry2 := &model.AuditEntry{
		Timestamp: time.Now(),
		TraceID:   "trace_456",
		BuyerID:   "buyer_1",
		Action:    model.AuditActionPaymentExecuted,
		Request:   json.RawMessage(`{"order_id": "order_1"}`),
		Response:  json.RawMessage(`{"payment_id": "pay_1"}`),
	}

	err = service.Log(context.Background(), entry2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry2.Index != 2 {
		t.Errorf("expected index=2, got %d", entry2.Index)
	}
	if entry2.PrevHash != entry.Hash {
		t.Errorf("expected prev_hash=%s, got %s", entry.Hash, entry2.PrevHash)
	}
}

// TestAuditService_GetTrail tests retrieving buyer audit trail.
func TestAuditService_GetTrail(t *testing.T) {
	repo := newMockAuditRepo()
	service := NewAuditService(repo)

	for i := 0; i < 3; i++ {
		entry := &model.AuditEntry{
			Timestamp: time.Now(),
			TraceID:   "trace_" + string(rune(i)),
			BuyerID:   "buyer_1",
			Action:    model.AuditActionBrowse,
		}
		_ = service.Log(context.Background(), entry)
	}

	entry2 := &model.AuditEntry{
		Timestamp: time.Now(),
		TraceID:   "trace_other",
		BuyerID:   "buyer_2",
		Action:    model.AuditActionBrowse,
	}
	_ = service.Log(context.Background(), entry2)

	trail, err := service.GetTrail(context.Background(), "buyer_1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(trail) != 3 {
		t.Errorf("expected 3 entries for buyer_1, got %d", len(trail))
	}
	for _, e := range trail {
		if e.BuyerID != "buyer_1" {
			t.Errorf("expected all entries for buyer_1, got %s", e.BuyerID)
		}
	}
}

// TestAuditService_GetByTraceID tests retrieving by trace ID.
func TestAuditService_GetByTraceID(t *testing.T) {
	repo := newMockAuditRepo()
	service := NewAuditService(repo)

	entry := &model.AuditEntry{
		Timestamp: time.Now(),
		TraceID:   "trace_abc",
		BuyerID:   "buyer_1",
		Action:    model.AuditActionPurchaseAttempt,
	}
	_ = service.Log(context.Background(), entry)

	entries, err := service.GetByTraceID(context.Background(), "trace_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].TraceID != "trace_abc" {
		t.Errorf("expected trace_abc, got %s", entries[0].TraceID)
	}
}

// TestAuditService_VerifyIntegrity tests chain verification.
func TestAuditService_VerifyIntegrity(t *testing.T) {
	repo := newMockAuditRepo()
	service := NewAuditService(repo)

	valid, err := service.VerifyIntegrity(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected empty log to be valid")
	}

	for i := 0; i < 3; i++ {
		entry := &model.AuditEntry{
			Timestamp: time.Now(),
			TraceID:   "trace_" + string(rune(i)),
			BuyerID:   "buyer_1",
			Action:    model.AuditActionBrowse,
		}
		_ = service.Log(context.Background(), entry)
	}

	valid, err = service.VerifyIntegrity(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid chain")
	}

	repo.verifyOK = false
	valid, err = service.VerifyIntegrity(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Error("expected invalid chain")
	}
}

// TestAuditService_Log_NilEntry tests nil entry handling.
func TestAuditService_Log_NilEntry(t *testing.T) {
	repo := newMockAuditRepo()
	service := NewAuditService(repo)

	err := service.Log(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil entry")
	}
}
