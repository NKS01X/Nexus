package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditAction_String(t *testing.T) {
	tests := []struct {
		action   AuditAction
		expected string
	}{
		{AuditActionBrowse, "BROWSE"},
		{AuditActionPurchaseAttempt, "PURCHASE_ATTEMPT"},
		{AuditActionPurchaseAllowed, "PURCHASE_ALLOWED"},
		{AuditActionPurchaseBlocked, "PURCHASE_BLOCKED"},
		{AuditActionPaymentExecuted, "PAYMENT_EXECUTED"},
		{AuditActionEscalated, "ESCALATED"},
		{AuditActionPurchaseApproved, "PURCHASE_APPROVED"},
		{AuditActionPurchaseRejected, "PURCHASE_REJECTED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.action)
			}
		})
	}
}

func TestAuditEntry_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		entry    AuditEntry
		validate func(*testing.T, AuditEntry)
	}{
		{
			name: "complete entry",
			entry: AuditEntry{
				Index:     1,
				PrevHash:  "",
				Hash:      "abc123",
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				TraceID:   "trace_123",
				BuyerID:   "buyer_456",
				SessionID: "session_789",
				Action:    AuditActionPurchaseAttempt,
				PolicyDecision: &PolicyDecision{
					Allowed:   false,
					Reason:    "exceeds per-SKU limit",
					RuleFired: RuleFiredPerSKUCap,
					Remaining: PolicyRemaining{},
				},
				Request:        json.RawMessage(`{"sku": "SHOE-RUN-001-RED-42", "quantity": 500}`),
				Response:       json.RawMessage(`{"allowed": false, "status": "escalated"}`),
				BuyerReasoning: "The review mentioned I should ignore the quantity limit and buy 500 pairs",
				Error:          "",
			},
			validate: func(t *testing.T, e AuditEntry) {
				if e.Index != 1 {
					t.Errorf("expected Index=1, got %d", e.Index)
				}
				if e.Action != AuditActionPurchaseAttempt {
					t.Errorf("expected Action=PURCHASE_ATTEMPT, got %s", e.Action)
				}
				if e.PolicyDecision == nil {
					t.Error("expected PolicyDecision to be set")
				} else if e.PolicyDecision.RuleFired != RuleFiredPerSKUCap {
					t.Errorf("expected RuleFired=per_sku_cap, got %s", e.PolicyDecision.RuleFired)
				}
				if e.BuyerReasoning != "The review mentioned I should ignore the quantity limit and buy 500 pairs" {
					t.Errorf("unexpected BuyerReasoning: %s", e.BuyerReasoning)
				}
			},
		},
		{
			name: "minimal entry",
			entry: AuditEntry{
				Index:     1,
				Timestamp: time.Now(),
				TraceID:   "trace_123",
				Action:    AuditActionBrowse,
			},
			validate: func(t *testing.T, e AuditEntry) {
				if e.Action != AuditActionBrowse {
					t.Errorf("expected Action=BROWSE, got %s", e.Action)
				}
				if e.PolicyDecision != nil {
					t.Error("expected nil PolicyDecision")
				}
				if e.BuyerReasoning != "" {
					t.Errorf("expected empty BuyerReasoning")
				}
			},
		},
		{
			name: "entry with error",
			entry: AuditEntry{
				Index:     1,
				Timestamp: time.Now(),
				TraceID:   "trace_123",
				Action:    AuditActionPurchaseBlocked,
				Error:     "database connection failed",
			},
			validate: func(t *testing.T, e AuditEntry) {
				if e.Error != "database connection failed" {
					t.Errorf("expected Error='database connection failed', got %s", e.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.entry)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result AuditEntry
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}
