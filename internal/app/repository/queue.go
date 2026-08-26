package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// ApprovalQueueRepository defines the interface for approval queue data access.
type ApprovalQueueRepository interface {
	Enqueue(ctx context.Context, item *model.PendingApproval) error
	GetByID(ctx context.Context, id string) (*model.PendingApproval, error)
	UpdateStatus(ctx context.Context, id string, status model.ApprovalStatus, reviewerID, note string) error
	ListPending(ctx context.Context, limit int) ([]*model.PendingApproval, error)
	GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.PendingApproval, error)
}
