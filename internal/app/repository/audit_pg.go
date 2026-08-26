package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/razorpay/aegis/internal/app/model"
)

// AuditPG implements AuditRepository using PostgreSQL.
type AuditPG struct {
	db *DB
}

// NewAuditPG creates a new PostgreSQL audit repository.
func NewAuditPG(db *DB) *AuditPG {
	return &AuditPG{db: db}
}

// Append adds a new entry to the hash-chained audit log.
func (r *AuditPG) Append(ctx context.Context, entry *model.AuditEntry) error {

	prevHash, err := r.getLatestHash(ctx)
	if err != nil {
		return fmt.Errorf("get latest hash: %w", err)
	}
	entry.PrevHash = prevHash

	hash, err := r.computeHash(entry)
	if err != nil {
		return fmt.Errorf("compute hash: %w", err)
	}
	entry.Hash = hash
	entry.Index = 0

	const query = `
		INSERT INTO audit_log (prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		                       action, policy_decision, request, response, buyer_reasoning, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING index
	`

	var policyDecisionJSON, requestJSON, responseJSON []byte
	if entry.PolicyDecision != nil {
		pdJSON, _ := json.Marshal(entry.PolicyDecision)
		policyDecisionJSON, _ = canonicalizeJSON(pdJSON)
	}
	if entry.Request != nil {
		requestJSON, _ = canonicalizeJSON(entry.Request)
	}
	if entry.Response != nil {
		responseJSON, _ = canonicalizeJSON(entry.Response)
	}

	err = r.db.QueryRowCtx(ctx, query,
		entry.PrevHash, entry.Hash, entry.Timestamp, entry.TraceID,
		entry.BuyerID, entry.SessionID, entry.Action,
		policyDecisionJSON, requestJSON, responseJSON,
		entry.BuyerReasoning, entry.Error,
	).Scan(&entry.Index)
	if err != nil {
		return fmt.Errorf("append audit: %w", err)
	}

	return nil
}

// GetByIndex retrieves an audit entry by its index.
func (r *AuditPG) GetByIndex(ctx context.Context, index int64) (*model.AuditEntry, error) {
	const query = `
		SELECT index, prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		       action, policy_decision, request, response, buyer_reasoning, error
		FROM audit_log WHERE index = $1
	`
	return r.scanAuditEntry(r.db.QueryRowCtx(ctx, query, index))
}

// GetRange retrieves audit entries in a range [from, to].
func (r *AuditPG) GetRange(ctx context.Context, from, to int64) ([]*model.AuditEntry, error) {
	const query = `
		SELECT index, prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		       action, policy_decision, request, response, buyer_reasoning, error
		FROM audit_log WHERE index >= $1 AND index <= $2 ORDER BY index
	`
	rows, err := r.db.QueryCtx(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("get range: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditEntry
	for rows.Next() {
		entry, err := r.scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return entries, nil
}

// GetByBuyer retrieves audit entries for a specific buyer.
func (r *AuditPG) GetByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.AuditEntry, error) {
	const query = `
		SELECT index, prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		       action, policy_decision, request, response, buyer_reasoning, error
		FROM audit_log WHERE buyer_id = $1 ORDER BY index DESC LIMIT $2
	`
	rows, err := r.db.QueryCtx(ctx, query, buyerID, limit)
	if err != nil {
		return nil, fmt.Errorf("get by buyer: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditEntry
	for rows.Next() {
		entry, err := r.scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return entries, nil
}

// GetAll retrieves all audit entries (for dashboard).
func (r *AuditPG) GetAll(ctx context.Context, limit int) ([]*model.AuditEntry, error) {
	const query = `
		SELECT index, prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		       action, policy_decision, request, response, buyer_reasoning, error
		FROM audit_log ORDER BY index DESC LIMIT $1
	`
	rows, err := r.db.QueryCtx(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get all: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditEntry
	for rows.Next() {
		entry, err := r.scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return entries, nil
}

// GetByTraceID retrieves audit entries for a specific trace ID.
func (r *AuditPG) GetByTraceID(ctx context.Context, traceID string) ([]*model.AuditEntry, error) {
	const query = `
		SELECT index, prev_hash, hash, timestamp, trace_id, buyer_id, session_id,
		       action, policy_decision, request, response, buyer_reasoning, error
		FROM audit_log WHERE trace_id = $1 ORDER BY index
	`
	rows, err := r.db.QueryCtx(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("get by trace: %w", err)
	}
	defer rows.Close()

	var entries []*model.AuditEntry
	for rows.Next() {
		entry, err := r.scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return entries, nil
}

// VerifyChain verifies the hash chain integrity from 'from' to 'to' (inclusive).
// 'from' should be the minimum index in the log (or the start of the chain to verify).
func (r *AuditPG) VerifyChain(ctx context.Context, from, to int64) (bool, error) {
	entries, err := r.GetRange(ctx, from, to)
	if err != nil {
		return false, err
	}

	var expectedPrevHash string

	prevEntry, err := r.GetByIndex(ctx, from-1)
	if err != nil {
		return false, fmt.Errorf("get prev entry: %w", err)
	}
	if prevEntry != nil {
		expectedPrevHash = prevEntry.Hash
	}

	for i, entry := range entries {

		if entry.PrevHash != expectedPrevHash {
			return false, nil
		}

		computedHash, err := r.computeHash(entry)
		if err != nil {
			return false, fmt.Errorf("compute hash for verification: %w", err)
		}
		if entry.Hash != computedHash {
			return false, nil
		}

		expectedPrevHash = entry.Hash

		expectedIndex := from + int64(i)
		if entry.Index != expectedIndex {
			return false, nil
		}
	}

	return true, nil
}

// GetLatestIndex returns the latest index in the audit log.
func (r *AuditPG) GetLatestIndex(ctx context.Context) (int64, error) {
	const query = `SELECT COALESCE(MAX(index), 0) FROM audit_log`
	var index int64
	err := r.db.QueryRowCtx(ctx, query).Scan(&index)
	if err != nil {
		return 0, fmt.Errorf("get latest index: %w", err)
	}
	return index, nil
}

// GetMinIndex returns the minimum index in the audit log.
func (r *AuditPG) GetMinIndex(ctx context.Context) (int64, error) {
	const query = `SELECT COALESCE(MIN(index), 0) FROM audit_log`
	var index int64
	err := r.db.QueryRowCtx(ctx, query).Scan(&index)
	if err != nil {
		return 0, fmt.Errorf("get min index: %w", err)
	}
	return index, nil
}

// getLatestHash returns the hash of the most recent entry.
func (r *AuditPG) getLatestHash(ctx context.Context) (string, error) {
	const query = `SELECT hash FROM audit_log ORDER BY index DESC LIMIT 1`
	var hash string
	err := r.db.QueryRowCtx(ctx, query).Scan(&hash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get latest hash: %w", err)
	}
	return hash, nil
}

// computeHash computes SHA256 hash of an audit entry.
// Index is excluded from hash since it's auto-generated by the database.
// Timestamp uses microsecond precision to match PostgreSQL TIMESTAMPTZ storage.
func (r *AuditPG) computeHash(entry *model.AuditEntry) (string, error) {

	canonRequest, _ := canonicalizeJSON(entry.Request)
	canonResponse, _ := canonicalizeJSON(entry.Response)

	var canonPolicyDecision json.RawMessage
	if entry.PolicyDecision != nil {
		pdJSON, _ := json.Marshal(entry.PolicyDecision)
		canonPolicyDecision, _ = canonicalizeJSON(pdJSON)
	}

	data := struct {
		PrevHash       string          `json:"prev_hash"`
		Timestamp      int64           `json:"timestamp"`
		TraceID        string          `json:"trace_id"`
		BuyerID        string          `json:"buyer_id,omitempty"`
		SessionID      string          `json:"session_id,omitempty"`
		Action         string          `json:"action"`
		PolicyDecision json.RawMessage `json:"policy_decision,omitempty"`
		Request        json.RawMessage `json:"request"`
		Response       json.RawMessage `json:"response"`
		BuyerReasoning string          `json:"buyer_reasoning,omitempty"`
		Error          string          `json:"error,omitempty"`
	}{
		PrevHash:       entry.PrevHash,
		Timestamp:      entry.Timestamp.UnixMicro(),
		TraceID:        entry.TraceID,
		BuyerID:        entry.BuyerID,
		SessionID:      entry.SessionID,
		Action:         string(entry.Action),
		PolicyDecision: canonPolicyDecision,
		Request:        canonRequest,
		Response:       canonResponse,
		BuyerReasoning: entry.BuyerReasoning,
		Error:          entry.Error,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal for hash: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// canonicalizeJSON ensures consistent JSON formatting by unmarshaling and re-marshaling.
// This handles cases where the database stores jsonb with different formatting.
func canonicalizeJSON(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// scanAuditEntry scans an audit entry from a Row or Rows.
func (r *AuditPG) scanAuditEntry(scanner interface{ Scan(dest ...any) error }) (*model.AuditEntry, error) {
	var e model.AuditEntry
	var policyDecisionJSON, requestJSON, responseJSON []byte
	var buyerID, sessionID, buyerReasoning, errorStr sql.NullString

	err := scanner.Scan(
		&e.Index, &e.PrevHash, &e.Hash, &e.Timestamp, &e.TraceID,
		&buyerID, &sessionID, &e.Action,
		&policyDecisionJSON, &requestJSON, &responseJSON,
		&buyerReasoning, &errorStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan audit entry: %w", err)
	}

	if buyerID.Valid {
		e.BuyerID = buyerID.String
	}
	if sessionID.Valid {
		e.SessionID = sessionID.String
	}
	if buyerReasoning.Valid {
		e.BuyerReasoning = buyerReasoning.String
	}
	if errorStr.Valid {
		e.Error = errorStr.String
	}

	if len(policyDecisionJSON) > 0 {
		var pd model.PolicyDecision
		_ = json.Unmarshal(policyDecisionJSON, &pd)
		e.PolicyDecision = &pd
	}
	e.Request = requestJSON
	e.Response = responseJSON

	return &e, nil
}
