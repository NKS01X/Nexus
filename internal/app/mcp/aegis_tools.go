package mcp

import (
	"encoding/json"
)

// Aegis MCP Tool Names
const (
	AegisToolPurchase = "purchase"
)

// AegisPurchaseParams holds parameters for a purchase request to the Aegis Gateway.
type AegisPurchaseParams struct {
	BuyerID        string          `json:"buyer_id"`
	SessionID      string          `json:"session_id"`
	ProductID      string          `json:"product_id"`
	SKU            string          `json:"sku"`
	Quantity       int             `json:"quantity"`
	AmountPaisa    int64           `json:"amount_paisa"`
	IdempotencyKey string          `json:"idempotency_key"`
	BuyerPincode   string          `json:"buyer_pincode,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// AegisPurchaseResult holds the result of a purchase request from the Aegis Gateway.
type AegisPurchaseResult struct {
	Allowed         bool            `json:"allowed"`
	Reason          string          `json:"reason"`
	RuleFired       string          `json:"rule_fired"`
	Status          string          `json:"status"`
	OrderID         string          `json:"order_id,omitempty"`
	PaymentID       string          `json:"payment_id,omitempty"`
	ApprovalQueueID string          `json:"approval_queue_id,omitempty"`
	Remaining       PolicyRemaining `json:"remaining"`
}

// PolicyRemaining mirrors the model.PolicyRemaining for MCP serialization.
type PolicyRemaining struct {
	SpendCapPaisa     int64          `json:"spend_cap_paisa"`
	PerSKUCap         map[string]int `json:"per_sku_cap"`
	VelocityRemaining int            `json:"velocity_remaining"`
}
