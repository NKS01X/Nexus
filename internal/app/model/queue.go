package model

import (
	"errors"
	"time"
)

// Sentinel errors for approval queue operations.
var (
	ErrApprovalNotFound = errors.New("approval not found")
)

// ApprovalStatus represents the status of a pending approval.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
	ApprovalStatusExpired  ApprovalStatus = "EXPIRED"
)

// PendingApproval represents a purchase request awaiting human review.
type PendingApproval struct {
	ID              string          `json:"id" db:"id"`
	BuyerID         string          `json:"buyer_id" db:"buyer_id"`
	SessionID       string          `json:"session_id" db:"session_id"`
	PurchaseRequest PurchaseRequest `json:"purchase_request" db:"purchase_request"`
	PolicyDecision  PolicyDecision  `json:"policy_decision" db:"policy_decision"`
	BuyerReasoning  string          `json:"buyer_reasoning" db:"buyer_reasoning"`
	Status          ApprovalStatus  `json:"status" db:"status"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewerID      string          `json:"reviewer_id,omitempty" db:"reviewer_id"`
	ReviewNote      string          `json:"review_note,omitempty" db:"review_note"`
	ExpiresAt       time.Time       `json:"expires_at" db:"expires_at"`
}
