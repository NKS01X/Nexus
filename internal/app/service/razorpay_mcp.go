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
type CreateOrderResponse struct {
	OrderID     string
	AmountPaisa int64
	Currency    string
	Status      string
	CheckoutURL string
}

// CapturePaymentResponse holds the response from capturing a payment.
type CapturePaymentResponse struct {
	PaymentID   string
	OrderID     string
	AmountPaisa int64
	Status      string
}

// PaymentResponse holds payment details from Razorpay.
type PaymentResponse struct {
	PaymentID      string
	OrderID        string
	AmountPaisa    int64
	Currency       string
	Status         string
	Method         string
	Captured       bool
	RefundedAmount int64
}

// CreateRefundRequest holds parameters for creating a refund.
type CreateRefundRequest struct {
	PaymentID      string
	AmountPaisa    int64
	Currency       string
	IdempotencyKey string
}

// CreateRefundResponse holds the response from creating a refund.
type CreateRefundResponse struct {
	RefundID    string
	AmountPaisa int64
	Status      string
}
