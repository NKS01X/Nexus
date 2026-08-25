package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOrder_MarshalUnmarshal(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		order    Order
		validate func(*testing.T, Order)
	}{
		{
			name: "complete order",
			order: Order{
				ID:                "order_123",
				BuyerID:           "buyer_456",
				SessionID:         "session_789",
				ProductID:         "prod_001",
				SKU:               "SHOE-RUN-001-RED-42",
				Quantity:          1,
				AmountPaisa:       249900,
				Currency:          "INR",
				Status:            OrderStatusPaid,
				RazorpayOrderID:   "rzp_order_abc",
				RazorpayPaymentID: "rzp_payment_xyz",
				IdempotencyKey:    "idem_999",
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			validate: func(t *testing.T, o Order) {
				if o.ID != "order_123" {
					t.Errorf("expected ID=order_123, got %s", o.ID)
				}
				if o.Status != OrderStatusPaid {
					t.Errorf("expected Status=PAID, got %s", o.Status)
				}
				if o.RazorpayOrderID != "rzp_order_abc" {
					t.Errorf("expected RazorpayOrderID=rzp_order_abc, got %s", o.RazorpayOrderID)
				}
			},
		},
		{
			name: "pending order without razorpay IDs",
			order: Order{
				ID:             "order_124",
				BuyerID:        "buyer_456",
				SessionID:      "session_789",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				Currency:       "INR",
				Status:         OrderStatusPending,
				IdempotencyKey: "idem_999",
			},
			validate: func(t *testing.T, o Order) {
				if o.Status != OrderStatusPending {
					t.Errorf("expected Status=PENDING, got %s", o.Status)
				}
				if o.RazorpayOrderID != "" {
					t.Errorf("expected empty RazorpayOrderID, got %s", o.RazorpayOrderID)
				}
				if o.RazorpayPaymentID != "" {
					t.Errorf("expected empty RazorpayPaymentID, got %s", o.RazorpayPaymentID)
				}
			},
		},
		{
			name:  "zero values",
			order: Order{},
			validate: func(t *testing.T, o Order) {
				if o.Status != "" {
					t.Errorf("expected empty Status")
				}
				if o.Quantity != 0 {
					t.Errorf("expected Quantity=0")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.order)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result Order
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestOrderStatusConstants(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{OrderStatusPending, "PENDING"},
		{OrderStatusPaid, "PAID"},
		{OrderStatusFailed, "FAILED"},
		{OrderStatusRefunded, "REFUNDED"},
		{OrderStatusCancelled, "CANCELLED"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if tt.status != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}
