package service

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// PolicyEngine defines the interface for deterministic policy evaluation.
type PolicyEngine interface {
	Evaluate(ctx context.Context, req *model.PurchaseRequest) (*model.PolicyDecision, error)
	RecordSpend(ctx context.Context, buyerID, sessionID, sku string, quantity int, amountPaisa int64) error
	RollbackSpend(ctx context.Context, buyerID, sessionID, sku string, quantity int, amountPaisa int64) error
	IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error
}
