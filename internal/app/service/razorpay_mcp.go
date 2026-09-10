package service

import (
	"context"
)

// RazorpayMCPClient defines the interface for the Razorpay MCP client.
type RazorpayMCPClient interface {
	CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResponse, error)
	CapturePayment(ctx context.Context, paymentID string) (*CapturePaymentResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*PaymentResponse, error)
	CreateRefund(ctx context.Context, req CreateRefundRequest) (*CreateRefundResponse, error)
	Close() error
	IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error
}

// CreateOrderRequest holds parameters for creating a Razorpay order.
type CreateOrderRequest struct {
	AmountPaisa int64
	Currency    string
	Receipt     string
	Notes       map[string]string
}

// CreateOrderResponse holds the response from creating a Razorpay order.
// JSON tags match the actual Razorpay API response field names.
type CreateOrderResponse struct {
	OrderID     string `json:"id"`
	AmountPaisa int64  `json:"amount"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
	CheckoutURL string `json:"checkout_url,omitempty"`
}

// CapturePaymentResponse holds the response from capturing a payment.
// JSON tags match the actual Razorpay API response field names.
type CapturePaymentResponse struct {
	PaymentID   string `json:"id"`
	OrderID     string `json:"order_id"`
	AmountPaisa int64  `json:"amount"`
	Status      string `json:"status"`
}

// PaymentResponse holds payment details from Razorpay.
// JSON tags match the actual Razorpay API response field names.
type PaymentResponse struct {
	PaymentID      string `json:"id"`
	OrderID        string `json:"order_id"`
	AmountPaisa    int64  `json:"amount"`
	Currency       string `json:"currency"`
	Status         string `json:"status"`
	Method         string `json:"method"`
	Captured       bool   `json:"captured"`
	RefundedAmount int64  `json:"amount_refunded"`
}

// CreateRefundRequest holds parameters for creating a refund.
type CreateRefundRequest struct {
	PaymentID      string
	AmountPaisa    int64
	Currency       string
	IdempotencyKey string
}

// CreateRefundResponse holds the response from creating a refund.
// JSON tags match the actual Razorpay API response field names.
type CreateRefundResponse struct {
	RefundID    string `json:"id"`
	AmountPaisa int64  `json:"amount"`
	Status      string `json:"status"`
}
