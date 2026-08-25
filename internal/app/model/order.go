package model

import (
	"errors"
	"time"
)

// Sentinel errors for order operations.
var (
	ErrOrderNotFound = errors.New("order not found")
)

// Order represents a purchase order in the system.
type Order struct {
	ID                string    `json:"id" db:"id"`
	BuyerID           string    `json:"buyer_id" db:"buyer_id"`
	SessionID         string    `json:"session_id" db:"session_id"`
	ProductID         string    `json:"product_id" db:"product_id"`
	SKU               string    `json:"sku" db:"sku"`
	Quantity          int       `json:"quantity" db:"quantity"`
	AmountPaisa       int64     `json:"amount_paisa" db:"amount_paisa"`
	Currency          string    `json:"currency" db:"currency"`
	Status            string    `json:"status" db:"status"`
	RazorpayOrderID   string    `json:"razorpay_order_id,omitempty" db:"razorpay_order_id"`
	RazorpayPaymentID string    `json:"razorpay_payment_id,omitempty" db:"razorpay_payment_id"`
	IdempotencyKey    string    `json:"idempotency_key" db:"idempotency_key"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// OrderStatus constants for order lifecycle.
const (
	OrderStatusPending   = "PENDING"
	OrderStatusPaid      = "PAID"
	OrderStatusFailed    = "FAILED"
	OrderStatusRefunded  = "REFUNDED"
	OrderStatusCancelled = "CANCELLED"
)
