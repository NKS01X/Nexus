package service

import (
	"context"
	"fmt"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// AuditServiceImpl implements the AuditService interface with hash-chained log.
type AuditServiceImpl struct {
	repo repository.AuditRepository
}

// NewAuditService creates a new AuditServiceImpl.
func NewAuditService(repo repository.AuditRepository) *AuditServiceImpl {
	return &AuditServiceImpl{repo: repo}
}

// Log appends an entry to the hash-chained audit log.
func (s *AuditServiceImpl) Log(ctx context.Context, entry *model.AuditEntry) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = entry.Timestamp.UTC()
	}
	return s.repo.Append(ctx, entry)
}

// GetTrail retrieves audit trail for a specific buyer.
func (s *AuditServiceImpl) GetTrail(ctx context.Context, buyerID string, limit int) ([]*model.AuditEntry, error) {
	return s.repo.GetByBuyer(ctx, buyerID, limit)
}

// GetAll retrieves all audit entries (for dashboard).
func (s *AuditServiceImpl) GetAll(ctx context.Context, limit int) ([]*model.AuditEntry, error) {
	return s.repo.GetAll(ctx, limit)
}

// GetByTraceID retrieves audit entries by trace ID.
func (s *AuditServiceImpl) GetByTraceID(ctx context.Context, traceID string) ([]*model.AuditEntry, error) {
	return s.repo.GetByTraceID(ctx, traceID)
}

// VerifyIntegrity verifies the entire hash chain integrity.
func (s *AuditServiceImpl) VerifyIntegrity(ctx context.Context) (bool, error) {
	latestIndex, err := s.repo.GetLatestIndex(ctx)
	if err != nil {
		return false, fmt.Errorf("get latest index: %w", err)
	}
	if latestIndex == 0 {
		return true, nil
	}

	minIndex, err := s.repo.GetMinIndex(ctx)
	if err != nil {
		return false, fmt.Errorf("get min index: %w", err)
	}

	return s.repo.VerifyChain(ctx, minIndex, latestIndex)
}
