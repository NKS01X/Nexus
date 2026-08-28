package service

import (
	"context"
	"fmt"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// ApprovalQueueService defines the interface for approval queue operations.
type ApprovalQueueService interface {
	Enqueue(ctx context.Context, item *model.PendingApproval) error
	GetByID(ctx context.Context, id string) (*model.PendingApproval, error)
	Approve(ctx context.Context, id, reviewerID, note string) error
	Reject(ctx context.Context, id, reviewerID, note string) error
	ListPending(ctx context.Context, limit int) ([]*model.PendingApproval, error)
	GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.PendingApproval, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

// ApprovalQueueServiceImpl implements ApprovalQueueService.
type ApprovalQueueServiceImpl struct {
	repo repository.ApprovalQueueRepository
}

// NewApprovalQueueService creates a new ApprovalQueueServiceImpl.
func NewApprovalQueueService(repo repository.ApprovalQueueRepository) *ApprovalQueueServiceImpl {
	return &ApprovalQueueServiceImpl{repo: repo}
}

// Enqueue adds a new pending approval to the queue.
func (s *ApprovalQueueServiceImpl) Enqueue(ctx context.Context, item *model.PendingApproval) error {
	if item == nil {
		return fmt.Errorf("item is nil")
	}
	if item.ID == "" {
		return fmt.Errorf("approval ID is required")
	}
	if item.BuyerID == "" || item.SessionID == "" {
		return fmt.Errorf("buyer_id and session_id are required")
	}
	if item.ExpiresAt.IsZero() {
		item.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.Status == "" {
		item.Status = model.ApprovalStatusPending
	}
	return s.repo.Enqueue(ctx, item)
}

// GetByID retrieves a pending approval by ID.
func (s *ApprovalQueueServiceImpl) GetByID(ctx context.Context, id string) (*model.PendingApproval, error) {
	return s.repo.GetByID(ctx, id)
}

// Approve approves a pending approval.
func (s *ApprovalQueueServiceImpl) Approve(ctx context.Context, id, reviewerID, note string) error {
	if id == "" {
		return fmt.Errorf("approval ID is required")
	}
	if reviewerID == "" {
		return fmt.Errorf("reviewer ID is required")
	}
	return s.repo.UpdateStatus(ctx, id, model.ApprovalStatusApproved, reviewerID, note)
}

// Reject rejects a pending approval.
func (s *ApprovalQueueServiceImpl) Reject(ctx context.Context, id, reviewerID, note string) error {
	if id == "" {
		return fmt.Errorf("approval ID is required")
	}
	if reviewerID == "" {
		return fmt.Errorf("reviewer ID is required")
	}
	return s.repo.UpdateStatus(ctx, id, model.ApprovalStatusRejected, reviewerID, note)
}

// ListPending lists all pending (non-expired) approvals.
func (s *ApprovalQueueServiceImpl) ListPending(ctx context.Context, limit int) ([]*model.PendingApproval, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListPending(ctx, limit)
}

// GetByBuyer retrieves pending approvals for a specific buyer.
func (s *ApprovalQueueServiceImpl) GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.PendingApproval, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetByBuyer(ctx, buyerID, limit)
}

// CleanupExpired marks expired pending approvals as EXPIRED.
// Returns the count of approvals that were expired.
func (s *ApprovalQueueServiceImpl) CleanupExpired(ctx context.Context) (int64, error) {

	return 0, fmt.Errorf("not implemented: requires repository method")
}
