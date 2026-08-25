package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPolicyConfig_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		config   PolicyConfig
		validate func(*testing.T, PolicyConfig)
	}{
		{
			name: "valid config with all fields",
			config: PolicyConfig{
				SpendCapPaisa:     300000,
				PerSKUCap:         map[string]int{"SHOE-RUN-001": 2},
				VelocityCap:       VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
				AllowedCategories: []string{"footwear", "apparel"},
				BlockedSKUs:       []string{},
				GeoRules: []GeoRule{
					{Country: "IN", Allowed: true, Pincodes: []string{"560001", "560002"}},
				},
			},
			validate: func(t *testing.T, c PolicyConfig) {
				if c.SpendCapPaisa != 300000 {
					t.Errorf("expected SpendCapPaisa=300000, got %d", c.SpendCapPaisa)
				}
				if c.PerSKUCap["SHOE-RUN-001"] != 2 {
					t.Errorf("expected PerSKUCap[SHOE-RUN-001]=2, got %d", c.PerSKUCap["SHOE-RUN-001"])
				}
				if c.VelocityCap.MaxRequests != 10 {
					t.Errorf("expected MaxRequests=10, got %d", c.VelocityCap.MaxRequests)
				}
				if len(c.AllowedCategories) != 2 {
					t.Errorf("expected 2 allowed categories, got %d", len(c.AllowedCategories))
				}
				if len(c.GeoRules) != 1 {
					t.Errorf("expected 1 geo rule, got %d", len(c.GeoRules))
				}
			},
		},
		{
			name: "empty config",
			config: PolicyConfig{
				PerSKUCap:         map[string]int{},
				AllowedCategories: []string{},
				BlockedSKUs:       []string{},
				GeoRules:          []GeoRule{},
			},
			validate: func(t *testing.T, c PolicyConfig) {
				if c.SpendCapPaisa != 0 {
					t.Errorf("expected SpendCapPaisa=0, got %d", c.SpendCapPaisa)
				}
				if c.VelocityCap.MaxRequests != 0 {
					t.Errorf("expected MaxRequests=0, got %d", c.VelocityCap.MaxRequests)
				}
			},
		},
		{
			name: "config with nil maps/slices",
			config: PolicyConfig{
				PerSKUCap:         nil,
				AllowedCategories: nil,
				BlockedSKUs:       nil,
				GeoRules:          nil,
			},
			validate: func(t *testing.T, c PolicyConfig) {
				if c.PerSKUCap != nil {
					t.Errorf("expected nil PerSKUCap")
				}
				if c.AllowedCategories != nil {
					t.Errorf("expected nil AllowedCategories")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result PolicyConfig
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestVelocityLimit_GetWindow(t *testing.T) {
	tests := []struct {
		name           string
		limit          VelocityLimit
		expectedWindow time.Duration
	}{
		{
			name:           "window seconds set",
			limit:          VelocityLimit{WindowSeconds: 60},
			expectedWindow: 60 * time.Second,
		},
		{
			name:           "window already set",
			limit:          VelocityLimit{Window: 30 * time.Second, WindowSeconds: 60},
			expectedWindow: 30 * time.Second,
		},
		{
			name:           "zero values",
			limit:          VelocityLimit{},
			expectedWindow: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limit.GetWindow()
			if result != tt.expectedWindow {
				t.Errorf("expected %v, got %v", tt.expectedWindow, result)
			}
		})
	}
}

func TestPolicyDecision_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		decision PolicyDecision
		validate func(*testing.T, PolicyDecision)
	}{
		{
			name: "allowed decision",
			decision: PolicyDecision{
				Allowed:   true,
				Reason:    "within limits",
				RuleFired: RuleFiredNone,
				Remaining: PolicyRemaining{
					SpendCapPaisa:     297501,
					PerSKUCap:         map[string]int{"SHOE-RUN-001": 1},
					VelocityRemaining: 9,
				},
			},
			validate: func(t *testing.T, d PolicyDecision) {
				if !d.Allowed {
					t.Error("expected Allowed=true")
				}
				if d.RuleFired != RuleFiredNone {
					t.Errorf("expected RuleFired=none, got %s", d.RuleFired)
				}
				if d.Remaining.SpendCapPaisa != 297501 {
					t.Errorf("expected SpendCapPaisa=297501, got %d", d.Remaining.SpendCapPaisa)
				}
			},
		},
		{
			name: "blocked decision with spend cap",
			decision: PolicyDecision{
				Allowed:   false,
				Reason:    "exceeds session budget",
				RuleFired: RuleFiredSpendCap,
				Details:   json.RawMessage(`{"current_spend": 295000, "requested": 10000}`),
				Remaining: PolicyRemaining{
					SpendCapPaisa:     0,
					PerSKUCap:         map[string]int{},
					VelocityRemaining: 0,
				},
			},
			validate: func(t *testing.T, d PolicyDecision) {
				if d.Allowed {
					t.Error("expected Allowed=false")
				}
				if d.RuleFired != RuleFiredSpendCap {
					t.Errorf("expected RuleFired=spend_cap, got %s", d.RuleFired)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.decision)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result PolicyDecision
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestPurchaseRequest_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		request  PurchaseRequest
		validate func(*testing.T, PurchaseRequest)
	}{
		{
			name: "complete request",
			request: PurchaseRequest{
				BuyerID:        "buyer_123",
				SessionID:      "session_456",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_789",
				BuyerPincode:   "560001",
				Metadata:       json.RawMessage(`{"reasoning": "user wants red shoes"}`),
			},
			validate: func(t *testing.T, r PurchaseRequest) {
				if r.BuyerID != "buyer_123" {
					t.Errorf("expected BuyerID=buyer_123, got %s", r.BuyerID)
				}
				if r.Quantity != 1 {
					t.Errorf("expected Quantity=1, got %d", r.Quantity)
				}
				if r.Metadata == nil {
					t.Error("expected Metadata to be set")
				}
			},
		},
		{
			name: "minimal request",
			request: PurchaseRequest{
				BuyerID:        "buyer_123",
				SessionID:      "session_456",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_789",
			},
			validate: func(t *testing.T, r PurchaseRequest) {
				if r.BuyerPincode != "" {
					t.Errorf("expected empty BuyerPincode, got %s", r.BuyerPincode)
				}
				if r.Metadata != nil {
					t.Error("expected nil Metadata")
				}
			},
		},
		{
			name:    "zero values",
			request: PurchaseRequest{},
			validate: func(t *testing.T, r PurchaseRequest) {
				if r.BuyerID != "" {
					t.Errorf("expected empty BuyerID")
				}
				if r.Quantity != 0 {
					t.Errorf("expected Quantity=0")
				}
				if r.AmountPaisa != 0 {
					t.Errorf("expected AmountPaisa=0")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result PurchaseRequest
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}
