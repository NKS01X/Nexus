package repository

import (
	"context"
	"testing"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

func TestApprovalQueuePG(t *testing.T) {
	dsn := getTestDSN(t)
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewApprovalQueuePG(db)
	ctx := context.Background()

	_, _ = db.ExecCtx(ctx, `DELETE FROM approval_queue`)

	t.Run("Enqueue_and_GetByID", func(t *testing.T) {
		item := &model.PendingApproval{
			ID:        "approval_001",
			BuyerID:   "buyer_1",
			SessionID: "session_1",
			PurchaseRequest: model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-001",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_001",
			},
			PolicyDecision: model.PolicyDecision{
				Allowed: false, Reason: "exceeds spend cap", RuleFired: model.RuleFiredSpendCap,
			},
			BuyerReasoning: "Need shoes for marathon",
			Status:         model.ApprovalStatusPending,
			CreatedAt:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}

		err := repo.Enqueue(ctx, item)
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := repo.GetByID(ctx, "approval_001")
		if err != nil {
			t.Fatal(err)
		}
		if retrieved == nil {
			t.Fatal("expected item, got nil")
		}
		if retrieved.ID != item.ID {
			t.Errorf("expected ID=%s, got %s", item.ID, retrieved.ID)
		}
		if retrieved.BuyerID != item.BuyerID {
			t.Errorf("expected BuyerID=%s, got %s", item.BuyerID, retrieved.BuyerID)
		}
		if retrieved.PurchaseRequest.SKU != item.PurchaseRequest.SKU {
			t.Errorf("expected SKU=%s, got %s", item.PurchaseRequest.SKU, retrieved.PurchaseRequest.SKU)
		}
		if retrieved.PolicyDecision.RuleFired != item.PolicyDecision.RuleFired {
			t.Errorf("expected RuleFired=%s, got %s", item.PolicyDecision.RuleFired, retrieved.PolicyDecision.RuleFired)
		}
		if retrieved.Status != model.ApprovalStatusPending {
			t.Errorf("expected status=PENDING, got %s", retrieved.Status)
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		item := &model.PendingApproval{
			ID:        "approval_002",
			BuyerID:   "buyer_2",
			SessionID: "session_2",
			PurchaseRequest: model.PurchaseRequest{
				BuyerID: "buyer_2", SessionID: "session_2", ProductID: "prod_002",
				SKU: "SHOE-002", Quantity: 1, AmountPaisa: 249900, IdempotencyKey: "idem_002",
			},
			PolicyDecision: model.PolicyDecision{Allowed: false, RuleFired: model.RuleFiredVelocityCap},
			Status:         model.ApprovalStatusPending,
			CreatedAt:      time.Now(),
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}
		_ = repo.Enqueue(ctx, item)

		err := repo.UpdateStatus(ctx, "approval_002", model.ApprovalStatusApproved, "reviewer_1", "Approved for loyal customer")
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := repo.GetByID(ctx, "approval_002")
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Status != model.ApprovalStatusApproved {
			t.Errorf("expected status=APPROVED, got %s", retrieved.Status)
		}
		if retrieved.ReviewedAt == nil {
			t.Error("expected ReviewedAt to be set")
		}
		if retrieved.ReviewerID != "reviewer_1" {
			t.Errorf("expected ReviewerID=reviewer_1, got %s", retrieved.ReviewerID)
		}
		if retrieved.ReviewNote != "Approved for loyal customer" {
			t.Errorf("expected ReviewNote='Approved for loyal customer', got %s", retrieved.ReviewNote)
		}
	})

	t.Run("UpdateStatus_not_found", func(t *testing.T) {
		err := repo.UpdateStatus(ctx, "nonexistent", model.ApprovalStatusApproved, "reviewer_1", "note")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != model.ErrApprovalNotFound {
			t.Errorf("expected ErrApprovalNotFound, got %v", err)
		}
	})

	t.Run("ListPending", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `DELETE FROM approval_queue`)

		expired := &model.PendingApproval{
			ID: "expired_001", BuyerID: "buyer_3", SessionID: "session_3",
			PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_3", SessionID: "session_3", SKU: "SKU-001", IdempotencyKey: "idem_exp"},
			PolicyDecision:  model.PolicyDecision{Allowed: false, RuleFired: model.RuleFiredSpendCap},
			Status:          model.ApprovalStatusPending, CreatedAt: time.Now().Add(-25 * time.Hour),
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		_ = repo.Enqueue(ctx, expired)

		pending := &model.PendingApproval{
			ID: "pending_001", BuyerID: "buyer_4", SessionID: "session_4",
			PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_4", SessionID: "session_4", SKU: "SKU-002", IdempotencyKey: "idem_pen"},
			PolicyDecision:  model.PolicyDecision{Allowed: false, RuleFired: model.RuleFiredVelocityCap},
			Status:          model.ApprovalStatusPending, CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		_ = repo.Enqueue(ctx, pending)

		items, err := repo.ListPending(ctx, 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 pending item, got %d", len(items))
		}
		if items[0].ID != "pending_001" {
			t.Errorf("expected ID=pending_001, got %s", items[0].ID)
		}
	})

	t.Run("GetByBuyer", func(t *testing.T) {
		item := &model.PendingApproval{
			ID: "buyer_item_001", BuyerID: "buyer_5", SessionID: "session_5",
			PurchaseRequest: model.PurchaseRequest{BuyerID: "buyer_5", SessionID: "session_5", SKU: "SKU-003", IdempotencyKey: "idem_b5"},
			PolicyDecision:  model.PolicyDecision{Allowed: false, RuleFired: model.RuleFiredCategoryBlocked},
			Status:          model.ApprovalStatusPending, CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		_ = repo.Enqueue(ctx, item)

		items, err := repo.GetByBuyer(ctx, "buyer_5", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %d", len(items))
		}
		if items[0].BuyerID != "buyer_5" {
			t.Errorf("expected BuyerID=buyer_5, got %s", items[0].BuyerID)
		}
	})
}
