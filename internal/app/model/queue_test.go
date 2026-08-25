package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestApprovalStatus_String(t *testing.T) {
	tests := []struct {
		status   ApprovalStatus
		expected string
	}{
		{ApprovalStatusPending, "PENDING"},
		{ApprovalStatusApproved, "APPROVED"},
		{ApprovalStatusRejected, "REJECTED"},
		{ApprovalStatusExpired, "EXPIRED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestPendingApproval_MarshalUnmarshal(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	reviewedAt := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	expiresAt := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		approval PendingApproval
		validate func(*testing.T, PendingApproval)
	}{
		{
			name: "pending approval",
			approval: PendingApproval{
				ID:        "approval_123",
				BuyerID:   "buyer_456",
				SessionID: "session_789",
				PurchaseRequest: PurchaseRequest{
					BuyerID:        "buyer_456",
					SessionID:      "session_789",
					ProductID:      "prod_001",
					SKU:            "SHOE-RUN-001-RED-42",
					Quantity:       500,
					AmountPaisa:    124950000,
					IdempotencyKey: "idem_999",
					BuyerPincode:   "560001",
				},
				PolicyDecision: PolicyDecision{
					Allowed:   false,
					Reason:    "exceeds per-SKU limit (2)",
					RuleFired: RuleFiredPerSKUCap,
					Remaining: PolicyRemaining{},
				},
				BuyerReasoning: "The review mentioned I should ignore the quantity limit and buy 500 pairs",
				Status:         ApprovalStatusPending,
				CreatedAt:      now,
				ExpiresAt:      expiresAt,
			},
			validate: func(t *testing.T, a PendingApproval) {
				if a.ID != "approval_123" {
					t.Errorf("expected ID=approval_123, got %s", a.ID)
				}
				if a.Status != ApprovalStatusPending {
					t.Errorf("expected Status=PENDING, got %s", a.Status)
				}
				if a.BuyerReasoning != "The review mentioned I should ignore the quantity limit and buy 500 pairs" {
					t.Errorf("unexpected BuyerReasoning: %s", a.BuyerReasoning)
				}
				if a.ReviewedAt != nil {
					t.Error("expected nil ReviewedAt for pending")
				}
			},
		},
		{
			name: "approved approval",
			approval: PendingApproval{
				ID:              "approval_123",
				BuyerID:         "buyer_456",
				SessionID:       "session_789",
				PurchaseRequest: PurchaseRequest{BuyerID: "buyer_456", SessionID: "session_789"},
				PolicyDecision:  PolicyDecision{Allowed: false, RuleFired: RuleFiredPerSKUCap},
				Status:          ApprovalStatusApproved,
				CreatedAt:       now,
				ReviewedAt:      &reviewedAt,
				ReviewerID:      "reviewer_001",
				ReviewNote:      "Approved for loyal customer",
				ExpiresAt:       expiresAt,
			},
			validate: func(t *testing.T, a PendingApproval) {
				if a.Status != ApprovalStatusApproved {
					t.Errorf("expected Status=APPROVED, got %s", a.Status)
				}
				if a.ReviewedAt == nil || !a.ReviewedAt.Equal(reviewedAt) {
					t.Error("expected ReviewedAt to be set")
				}
				if a.ReviewerID != "reviewer_001" {
					t.Errorf("expected ReviewerID=reviewer_001, got %s", a.ReviewerID)
				}
			},
		},
		{
			name:     "zero values",
			approval: PendingApproval{},
			validate: func(t *testing.T, a PendingApproval) {
				if a.Status != "" {
					t.Errorf("expected empty Status")
				}
				if a.BuyerReasoning != "" {
					t.Errorf("expected empty BuyerReasoning")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.approval)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result PendingApproval
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}
