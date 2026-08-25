package model

import (
	"encoding/json"
	"time"
)

// AuditAction represents the type of action being audited.
type AuditAction string

const (
	AuditActionBrowse           AuditAction = "BROWSE"
	AuditActionPurchaseAttempt  AuditAction = "PURCHASE_ATTEMPT"
	AuditActionPurchaseAllowed  AuditAction = "PURCHASE_ALLOWED"
	AuditActionPurchaseBlocked  AuditAction = "PURCHASE_BLOCKED"
	AuditActionPaymentExecuted  AuditAction = "PAYMENT_EXECUTED"
	AuditActionEscalated        AuditAction = "ESCALATED"
	AuditActionPurchaseApproved AuditAction = "PURCHASE_APPROVED"
	AuditActionPurchaseRejected AuditAction = "PURCHASE_REJECTED"
)

// AuditEntry represents a single entry in the hash-chained audit log.
type AuditEntry struct {
	Index          int64           `json:"index"`
	PrevHash       string          `json:"prev_hash"`
	Hash           string          `json:"hash"`
	Timestamp      time.Time       `json:"timestamp"`
	TraceID        string          `json:"trace_id"`
	BuyerID        string          `json:"buyer_id,omitempty"`
	SessionID      string          `json:"session_id,omitempty"`
	Action         AuditAction     `json:"action"`
	PolicyDecision *PolicyDecision `json:"policy_decision,omitempty"`
	Request        json.RawMessage `json:"request"`
	Response       json.RawMessage `json:"response"`
	BuyerReasoning string          `json:"buyer_reasoning,omitempty"`
	Error          string          `json:"error,omitempty"`
}
