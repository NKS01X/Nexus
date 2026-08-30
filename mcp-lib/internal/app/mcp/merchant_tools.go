package mcp

import (
	"encoding/json"

	"github.com/razorpay/aegis/internal/app/model"
)

// Merchant MCP Tool Names
const (
	MerchantToolSearchProducts    = "search_products"
	MerchantToolGetProduct        = "get_product"
	MerchantToolCheckAvailability = "check_availability"
	MerchantToolPurchase          = "purchase"
	MerchantToolGetOrderStatus    = "get_order_status"
)

// SearchProductsParams holds parameters for product search.
type SearchProductsParams struct {
	Query       string `json:"query,omitempty"`
	Category    string `json:"category,omitempty"`
	MaxPrice    *int64 `json:"max_price_paisa,omitempty"`
	MinPrice    *int64 `json:"min_price_paisa,omitempty"`
	InStockOnly bool   `json:"in_stock_only,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Color       string `json:"color,omitempty"`
	Size        string `json:"size,omitempty"`
	Brand       string `json:"brand,omitempty"`
}

// SearchProductsResult holds the search results.
type SearchProductsResult struct {
	Products []model.ProductSummary `json:"products"`
}

// GetProductParams holds parameters for getting a single product.
type GetProductParams struct {
	ProductID string `json:"product_id"`
}

// CheckAvailabilityParams holds parameters for checking SKU availability.
type CheckAvailabilityParams struct {
	SKU string `json:"sku"`
}

// CheckAvailabilityResult holds availability information.
type CheckAvailabilityResult struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
	Reserved  int    `json:"reserved"`
}

// PurchaseParams holds parameters for a purchase request.
type PurchaseParams struct {
	BuyerID        string `json:"buyer_id"`
	SessionID      string `json:"session_id"`
	ProductID      string `json:"product_id"`
	SKU            string `json:"sku"`
	Quantity       int    `json:"quantity"`
	IdempotencyKey string          `json:"idempotency_key"`
	BuyerPincode   string          `json:"buyer_pincode,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

// PurchaseResult holds the result of a purchase request.
type PurchaseResult struct {
	Allowed         bool   `json:"allowed"`
	Reason          string `json:"reason"`
	Status          string `json:"status"`
	OrderID         string `json:"order_id,omitempty"`
	PaymentID       string `json:"payment_id,omitempty"`
	ApprovalQueueID string `json:"approval_queue_id,omitempty"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
}

// GetOrderStatusParams holds parameters for getting order status.
type GetOrderStatusParams struct {
	OrderID string `json:"order_id"`
}

// GetOrderStatusResult holds order status information.
type GetOrderStatusResult struct {
	OrderID     string      `json:"order_id"`
	Status      string      `json:"status"`
	AmountPaisa int64       `json:"amount_paisa"`
	Currency    string      `json:"currency"`
	Items       []OrderItem `json:"items"`
}

// OrderItem represents an item in an order.
type OrderItem struct {
	SKU        string `json:"sku"`
	Quantity   int    `json:"quantity"`
	PricePaisa int64  `json:"price_paisa"`
}
