package service

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// AuditService defines the interface for audit log operations.
type AuditService interface {
	Log(ctx context.Context, entry *model.AuditEntry) error
	GetTrail(ctx context.Context, buyerID string, limit int) ([]*model.AuditEntry, error)
	GetAll(ctx context.Context, limit int) ([]*model.AuditEntry, error)
	GetByTraceID(ctx context.Context, traceID string) ([]*model.AuditEntry, error)
	VerifyIntegrity(ctx context.Context) (bool, error)
}
