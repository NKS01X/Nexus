package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

// ApprovalQueuePG implements ApprovalQueueRepository using PostgreSQL.
type ApprovalQueuePG struct {
	db *DB
}

// NewApprovalQueuePG creates a new PostgreSQL approval queue repository.
func NewApprovalQueuePG(db *DB) *ApprovalQueuePG {
	return &ApprovalQueuePG{db: db}
}

// Enqueue adds a new pending approval to the queue.
func (r *ApprovalQueuePG) Enqueue(ctx context.Context, item *model.PendingApproval) error {
	purchaseReqJSON, _ := json.Marshal(item.PurchaseRequest)
	policyDecisionJSON, _ := json.Marshal(item.PolicyDecision)

	const query = `
		INSERT INTO approval_queue (id, buyer_id, session_id, purchase_request, policy_decision,
		                            buyer_reasoning, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecCtx(ctx, query,
		item.ID, item.BuyerID, item.SessionID,
		purchaseReqJSON, policyDecisionJSON, item.BuyerReasoning,
		item.Status, item.CreatedAt, item.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("enqueue approval: %w", err)
	}
	return nil
}

// GetByID retrieves a pending approval by ID.
func (r *ApprovalQueuePG) GetByID(ctx context.Context, id string) (*model.PendingApproval, error) {
	const query = `
		SELECT id, buyer_id, session_id, purchase_request, policy_decision, buyer_reasoning,
		       status, created_at, reviewed_at, reviewer_id, review_note, expires_at
		FROM approval_queue WHERE id = $1
	`
	return r.scanApprovalItem(r.db.QueryRowCtx(ctx, query, id))
}

// UpdateStatus updates the status of a pending approval.
func (r *ApprovalQueuePG) UpdateStatus(ctx context.Context, id string, status model.ApprovalStatus, reviewerID, note string) error {
	now := time.Now()
	const query = `
		UPDATE approval_queue
		SET status = $1, reviewed_at = $2, reviewer_id = $3, review_note = $4
		WHERE id = $5
	`
	result, err := r.db.ExecCtx(ctx, query, status, now, reviewerID, note, id)
	if err != nil {
		return fmt.Errorf("update approval status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrApprovalNotFound
	}
	return nil
}

// ListPending lists pending approvals that haven't expired.
func (r *ApprovalQueuePG) ListPending(ctx context.Context, limit int) ([]*model.PendingApproval, error) {
	const query = `
		SELECT id, buyer_id, session_id, purchase_request, policy_decision, buyer_reasoning,
		       status, created_at, reviewed_at, reviewer_id, review_note, expires_at
		FROM approval_queue
		WHERE status = 'PENDING' AND expires_at > NOW()
		ORDER BY created_at ASC LIMIT $1
	`
	rows, err := r.db.QueryCtx(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending: %w", err)
	}
	defer rows.Close()

	var items []*model.PendingApproval
	for rows.Next() {
		item, err := r.scanApprovalItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return items, nil
}

// GetByBuyer retrieves pending approvals for a specific buyer.
func (r *ApprovalQueuePG) GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.PendingApproval, error) {
	const query = `
		SELECT id, buyer_id, session_id, purchase_request, policy_decision, buyer_reasoning,
		       status, created_at, reviewed_at, reviewer_id, review_note, expires_at
		FROM approval_queue
		WHERE buyer_id = $1
		ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.QueryCtx(ctx, query, buyerID, limit)
	if err != nil {
		return nil, fmt.Errorf("get by buyer: %w", err)
	}
	defer rows.Close()

	var items []*model.PendingApproval
	for rows.Next() {
		item, err := r.scanApprovalItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return items, nil
}

// scanApprovalItem scans an approval item from a Row or Rows.
func (r *ApprovalQueuePG) scanApprovalItem(scanner interface{ Scan(dest ...any) error }) (*model.PendingApproval, error) {
	var item model.PendingApproval
	var purchaseReqJSON, policyDecisionJSON []byte
	var reviewedAt sql.NullTime
	var reviewerID, reviewNote sql.NullString

	err := scanner.Scan(
		&item.ID, &item.BuyerID, &item.SessionID,
		&purchaseReqJSON, &policyDecisionJSON, &item.BuyerReasoning,
		&item.Status, &item.CreatedAt, &reviewedAt, &reviewerID, &reviewNote, &item.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan approval item: %w", err)
	}

	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	if reviewerID.Valid {
		item.ReviewerID = reviewerID.String
	}
	if reviewNote.Valid {
		item.ReviewNote = reviewNote.String
	}

	_ = json.Unmarshal(purchaseReqJSON, &item.PurchaseRequest)
	_ = json.Unmarshal(policyDecisionJSON, &item.PolicyDecision)

	return &item, nil
}
