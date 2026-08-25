package model

import (
	"encoding/json"
	"time"
)

// PolicyConfig holds all policy configuration for the Aegis Gateway.
type PolicyConfig struct {
	SpendCapPaisa     int64          `json:"spend_cap_paisa" db:"spend_cap_paisa"`
	PerSKUCap         map[string]int `json:"per_sku_cap" db:"per_sku_cap"`
	VelocityCap       VelocityLimit  `json:"velocity_cap" db:"velocity_cap"`
	AllowedCategories []string       `json:"allowed_categories" db:"allowed_categories"`
	BlockedSKUs       []string       `json:"blocked_skus" db:"blocked_skus"`
	GeoRules          []GeoRule      `json:"geo_rules" db:"geo_rules"`
}

// VelocityLimit defines request rate limiting configuration.
type VelocityLimit struct {
	MaxRequests   int           `json:"max_requests" db:"max_requests"`
	Window        time.Duration `json:"-" db:"-"`
	WindowSeconds int           `json:"window_seconds" db:"window_seconds"`
}

// GetWindow returns the velocity window as a time.Duration.
func (v *VelocityLimit) GetWindow() time.Duration {
	if v.Window == 0 && v.WindowSeconds > 0 {
		v.Window = time.Duration(v.WindowSeconds) * time.Second
	}
	return v.Window
}

// GeoRule defines geographic restrictions for purchases.
type GeoRule struct {
	Country  string   `json:"country" db:"country"`
	Allowed  bool     `json:"allowed" db:"allowed"`
	Pincodes []string `json:"pincodes,omitempty" db:"pincodes"`
}

// PolicyDecision represents the result of a policy evaluation.
type PolicyDecision struct {
	Allowed   bool            `json:"allowed"`
	Reason    string          `json:"reason"`
	RuleFired string          `json:"rule_fired"`
	Details   json.RawMessage `json:"details,omitempty"`
	Remaining PolicyRemaining `json:"remaining"`
}

// PolicyRemaining holds remaining capacity after a policy decision.
type PolicyRemaining struct {
	SpendCapPaisa     int64          `json:"spend_cap_paisa"`
	PerSKUCap         map[string]int `json:"per_sku_cap"`
	VelocityRemaining int            `json:"velocity_remaining"`
}

// PurchaseRequest represents a purchase request from an AI buyer.
type PurchaseRequest struct {
	BuyerID        string          `json:"buyer_id"`
	SessionID      string          `json:"session_id"`
	ProductID      string          `json:"product_id"`
	SKU            string          `json:"sku"`
	Quantity       int             `json:"quantity"`
	AmountPaisa    int64           `json:"amount_paisa"`
	IdempotencyKey string          `json:"idempotency_key"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	BuyerPincode   string          `json:"buyer_pincode,omitempty"`
}

// RuleFired constants for policy decision rule identification.
const (
	RuleFiredSpendCap        = "spend_cap"
	RuleFiredPerSKUCap       = "per_sku_cap"
	RuleFiredVelocityCap     = "velocity_cap"
	RuleFiredCategoryBlocked = "category_blocked"
	RuleFiredSKUBlocked      = "sku_blocked"
	RuleFiredGeoRestricted   = "geo_restricted"
	RuleFiredNone            = "none"
)
