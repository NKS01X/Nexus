package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// AuditRepository defines the interface for hash-chained audit log access.
type AuditRepository interface {
	Append(ctx context.Context, entry *model.AuditEntry) error
	GetByIndex(ctx context.Context, index int64) (*model.AuditEntry, error)
	GetRange(ctx context.Context, from, to int64) ([]*model.AuditEntry, error)
	GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.AuditEntry, error)
	GetAll(ctx context.Context, limit int) ([]*model.AuditEntry, error)
	GetByTraceID(ctx context.Context, traceID string) ([]*model.AuditEntry, error)
	VerifyChain(ctx context.Context, from, to int64) (bool, error)
	GetLatestIndex(ctx context.Context) (int64, error)
	GetMinIndex(ctx context.Context) (int64, error)
}
