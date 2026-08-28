package service

import (
	"context"
	"testing"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

// mockApprovalQueueRepo implements repository.ApprovalQueueRepository for testing.
type mockApprovalQueueRepo struct {
	items map[string]*model.PendingApproval
}

func newMockApprovalQueueRepo() *mockApprovalQueueRepo {
	return &mockApprovalQueueRepo{
		items: make(map[string]*model.PendingApproval),
	}
}

func (m *mockApprovalQueueRepo) Enqueue(ctx context.Context, item *model.PendingApproval) error {
	m.items[item.ID] = item
	return nil
}

func (m *mockApprovalQueueRepo) GetByID(ctx context.Context, id string) (*model.PendingApproval, error) {
	if item, ok := m.items[id]; ok {
		return item, nil
	}
	return nil, nil
}

func (m *mockApprovalQueueRepo) UpdateStatus(ctx context.Context, id string, status model.ApprovalStatus, reviewerID, note string) error {
	item, ok := m.items[id]
	if !ok {
		return model.ErrApprovalNotFound
	}
	now := time.Now()
	item.Status = status
	item.ReviewedAt = &now
	item.ReviewerID = reviewerID
	item.ReviewNote = note
	return nil
}

func (m *mockApprovalQueueRepo) ListPending(ctx context.Context, limit int) ([]*model.PendingApproval, error) {
	var result []*model.PendingApproval
	for _, item := range m.items {
		if item.Status == model.ApprovalStatusPending && item.ExpiresAt.After(time.Now()) {
			result = append(result, item)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *mockApprovalQueueRepo) GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.PendingApproval, error) {
	var result []*model.PendingApproval
	for _, item := range m.items {
		if item.BuyerID == buyerID {
			result = append(result, item)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

// TestApprovalQueueService_Enqueue tests enqueueing approvals.
func TestApprovalQueueService_Enqueue(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	item := &model.PendingApproval{
		ID:              "appr_1",
		BuyerID:         "buyer_1",
		SessionID:       "session_1",
		BuyerReasoning:  "need shoes",
		Status:          model.ApprovalStatusPending,
		PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_1", SKU: "SHOE-001"},
	}

	err := service.Enqueue(context.Background(), item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, err := service.GetByID(context.Background(), "appr_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected item, got nil")
	}
	if retrieved.ID != "appr_1" {
		t.Errorf("expected ID=appr_1, got %s", retrieved.ID)
	}
	if retrieved.Status != model.ApprovalStatusPending {
		t.Errorf("expected status=PENDING, got %s", retrieved.Status)
	}
	if retrieved.ExpiresAt.IsZero() {
		t.Error("expected ExpiresAt to be set")
	}
}

// TestApprovalQueueService_Enqueue_Validation tests validation.
func TestApprovalQueueService_Enqueue_Validation(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	tests := []struct {
		name    string
		item    *model.PendingApproval
		wantErr bool
	}{
		{
			name:    "nil item",
			item:    nil,
			wantErr: true,
		},
		{
			name:    "missing ID",
			item:    &model.PendingApproval{BuyerID: "b1", SessionID: "s1"},
			wantErr: true,
		},
		{
			name:    "missing buyer_id",
			item:    &model.PendingApproval{ID: "appr_1", SessionID: "s1"},
			wantErr: true,
		},
		{
			name:    "missing session_id",
			item:    &model.PendingApproval{ID: "appr_1", BuyerID: "b1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Enqueue(context.Background(), tt.item)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestApprovalQueueService_Approve tests approving an approval.
func TestApprovalQueueService_Approve(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	item := &model.PendingApproval{
		ID:              "appr_1",
		BuyerID:         "buyer_1",
		SessionID:       "session_1",
		Status:          model.ApprovalStatusPending,
		PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_1", SKU: "SHOE-001"},
	}
	_ = service.Enqueue(context.Background(), item)

	err := service.Approve(context.Background(), "appr_1", "reviewer_1", "approved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := service.GetByID(context.Background(), "appr_1")
	if retrieved.Status != model.ApprovalStatusApproved {
		t.Errorf("expected status=APPROVED, got %s", retrieved.Status)
	}
	if retrieved.ReviewerID != "reviewer_1" {
		t.Errorf("expected reviewer_1, got %s", retrieved.ReviewerID)
	}
	if retrieved.ReviewNote != "approved" {
		t.Errorf("expected note=approved, got %s", retrieved.ReviewNote)
	}
	if retrieved.ReviewedAt == nil {
		t.Error("expected ReviewedAt to be set")
	}
}

// TestApprovalQueueService_Reject tests rejecting an approval.
func TestApprovalQueueService_Reject(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	item := &model.PendingApproval{
		ID:              "appr_1",
		BuyerID:         "buyer_1",
		SessionID:       "session_1",
		Status:          model.ApprovalStatusPending,
		PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_1", SKU: "SHOE-001"},
	}
	_ = service.Enqueue(context.Background(), item)

	err := service.Reject(context.Background(), "appr_1", "reviewer_1", "policy violation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, _ := service.GetByID(context.Background(), "appr_1")
	if retrieved.Status != model.ApprovalStatusRejected {
		t.Errorf("expected status=REJECTED, got %s", retrieved.Status)
	}
	if retrieved.ReviewNote != "policy violation" {
		t.Errorf("expected note=policy violation, got %s", retrieved.ReviewNote)
	}
}

// TestApprovalQueueService_Approve_NotFound tests approving non-existent.
func TestApprovalQueueService_Approve_NotFound(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	err := service.Approve(context.Background(), "nonexistent", "reviewer_1", "note")
	if err == nil {
		t.Error("expected error for non-existent approval")
	}
}

// TestApprovalQueueService_Approve_Validation tests validation.
func TestApprovalQueueService_Approve_Validation(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	tests := []struct {
		name     string
		id       string
		reviewer string
		note     string
		wantErr  bool
	}{
		{
			name:     "empty ID",
			id:       "",
			reviewer: "reviewer_1",
			note:     "note",
			wantErr:  true,
		},
		{
			name:     "empty reviewer",
			id:       "appr_1",
			reviewer: "",
			note:     "note",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Approve(context.Background(), tt.id, tt.reviewer, tt.note)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

// TestApprovalQueueService_ListPending tests listing pending approvals.
func TestApprovalQueueService_ListPending(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	item1 := &model.PendingApproval{
		ID:        "appr_1",
		BuyerID:   "buyer_1",
		SessionID: "session_1",
		Status:    model.ApprovalStatusPending,
		CreatedAt: time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	_ = service.Enqueue(context.Background(), item1)

	item2 := &model.PendingApproval{
		ID:        "appr_2",
		BuyerID:   "buyer_1",
		SessionID: "session_2",
		Status:    model.ApprovalStatusApproved,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	_ = service.Enqueue(context.Background(), item2)

	item3 := &model.PendingApproval{
		ID:        "appr_3",
		BuyerID:   "buyer_2",
		SessionID: "session_3",
		Status:    model.ApprovalStatusPending,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	_ = service.Enqueue(context.Background(), item3)

	pending, err := service.ListPending(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].ID != "appr_1" {
		t.Errorf("expected appr_1, got %s", pending[0].ID)
	}
}

// TestApprovalQueueService_GetByBuyer tests getting by buyer.
func TestApprovalQueueService_GetByBuyer(t *testing.T) {
	repo := newMockApprovalQueueRepo()
	service := NewApprovalQueueService(repo)

	item1 := &model.PendingApproval{
		ID:        "appr_1",
		BuyerID:   "buyer_1",
		SessionID: "session_1",
		Status:    model.ApprovalStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	_ = service.Enqueue(context.Background(), item1)

	item2 := &model.PendingApproval{
		ID:        "appr_2",
		BuyerID:   "buyer_2",
		SessionID: "session_2",
		Status:    model.ApprovalStatusPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	_ = service.Enqueue(context.Background(), item2)

	buyer1, err := service.GetByBuyer(context.Background(), "buyer_1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(buyer1) != 1 {
		t.Errorf("expected 1 for buyer_1, got %d", len(buyer1))
	}
	if buyer1[0].ID != "appr_1" {
		t.Errorf("expected appr_1, got %s", buyer1[0].ID)
	}
}
