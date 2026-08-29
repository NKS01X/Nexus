package razorpay_mcp

import (
	"context"
	"sync"
	"time"

	"github.com/razorpay/aegis/internal/app/service"
)

// MockClient is a mock implementation of RazorpayMCPClient for testing.
type MockClient struct {
	mu sync.Mutex
}

// NewMockClient creates a new mock Razorpay MCP client.
func NewMockClient() *MockClient {
	return &MockClient{}
}

// IncrementRequestCount is a no-op for mock client (velocity tracking not needed in tests).
func (c *MockClient) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	return nil
}

func (c *MockClient) CreateOrder(ctx context.Context, req service.CreateOrderRequest) (*service.CreateOrderResponse, error) {
	return &service.CreateOrderResponse{
		OrderID:     "mock_order_" + time.Now().Format("20060102150405.000000000"),
		AmountPaisa: req.AmountPaisa,
		Currency:    req.Currency,
		Status:      "created",
		CheckoutURL: "https://checkout.razorpay.com/mock",
	}, nil
}

func (c *MockClient) CapturePayment(ctx context.Context, paymentID string) (*service.CapturePaymentResponse, error) {
	return &service.CapturePaymentResponse{
		PaymentID:   "mock_payment_" + time.Now().Format("20060102150405.000000000"),
		OrderID:     "mock_order_123",
		AmountPaisa: 299900,
		Status:      "captured",
	}, nil
}

func (c *MockClient) GetPayment(ctx context.Context, paymentID string) (*service.PaymentResponse, error) {
	return &service.PaymentResponse{
		PaymentID:      paymentID,
		OrderID:        "mock_order_123",
		AmountPaisa:    299900,
		Currency:       "INR",
		Status:         "captured",
		Method:         "card",
		Captured:       true,
		RefundedAmount: 0,
	}, nil
}

func (c *MockClient) CreateRefund(ctx context.Context, req service.CreateRefundRequest) (*service.CreateRefundResponse, error) {
	return &service.CreateRefundResponse{
		RefundID:    "mock_refund_" + time.Now().Format("20060102150405.000000000"),
		AmountPaisa: req.AmountPaisa,
		Status:      "processed",
	}, nil
}

func (c *MockClient) Close() error {
	return nil
}
