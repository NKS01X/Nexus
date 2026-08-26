package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// PolicyRepository defines the interface for policy state data access.
type PolicyRepository interface {
	GetSpend(ctx context.Context, buyerID, sessionID string) (int64, error)
	AddSpend(ctx context.Context, buyerID, sessionID string, amountPaisa int64) error
	GetSKUQuantity(ctx context.Context, buyerID, sessionID, sku string) (int, error)
	AddSKUQuantity(ctx context.Context, buyerID, sessionID, sku string, quantity int) error
	GetRequestCount(ctx context.Context, buyerID, sessionID string, windowSeconds int) (int, error)
	IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error
	GetPolicyConfig(ctx context.Context) (*model.PolicyConfig, error)
	UpdatePolicyConfig(ctx context.Context, config *model.PolicyConfig) error
}
