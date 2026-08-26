package repository

import (
	"context"
	"testing"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

func TestAuditPG(t *testing.T) {
	dsn := getTestDSN(t)
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAuditPG(db)
	ctx := context.Background()

	_, _ = db.ExecCtx(ctx, `DELETE FROM audit_log`)
	_, _ = db.ExecCtx(ctx, `ALTER SEQUENCE audit_log_index_seq RESTART WITH 1`)

	t.Run("Append_first_entry", func(t *testing.T) {
		entry := &model.AuditEntry{
			Timestamp:      time.Now(),
			TraceID:        "trace_001",
			BuyerID:        "buyer_1",
			SessionID:      "session_1",
			Action:         model.AuditActionBrowse,
			Request:        []byte(`{"query": "shoes"}`),
			Response:       []byte(`{"products": []}`),
			BuyerReasoning: "User wants to browse",
		}

		err := repo.Append(ctx, entry)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Index != 1 {
			t.Errorf("expected index=1, got %d", entry.Index)
		}
		if entry.PrevHash != "" {
			t.Errorf("expected empty prev hash, got %s", entry.PrevHash)
		}
		if len(entry.Hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(entry.Hash))
		}
	})

	t.Run("Append_second_entry_chains", func(t *testing.T) {
		entry := &model.AuditEntry{
			Timestamp:      time.Now(),
			TraceID:        "trace_002",
			BuyerID:        "buyer_1",
			SessionID:      "session_1",
			Action:         model.AuditActionPurchaseAttempt,
			Request:        []byte(`{"sku": "SHOE-001", "qty": 1}`),
			Response:       []byte(`{"allowed": true}`),
			BuyerReasoning: "User wants to buy",
		}

		err := repo.Append(ctx, entry)
		if err != nil {
			t.Fatal(err)
		}
		if entry.Index != 2 {
			t.Errorf("expected index=2, got %d", entry.Index)
		}
		if entry.PrevHash == "" {
			t.Error("expected non-empty prev hash")
		}
		if len(entry.Hash) != 64 {
			t.Errorf("expected hash length 64, got %d", len(entry.Hash))
		}
	})

	t.Run("GetByIndex", func(t *testing.T) {
		entry, err := repo.GetByIndex(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if entry == nil {
			t.Fatal("expected entry, got nil")
		}
		if entry.Index != 1 {
			t.Errorf("expected index=1, got %d", entry.Index)
		}
		if entry.Action != model.AuditActionBrowse {
			t.Errorf("expected action=BROWSE, got %s", entry.Action)
		}
	})

	t.Run("GetByIndex_not_found", func(t *testing.T) {
		entry, err := repo.GetByIndex(ctx, 999)
		if err != nil {
			t.Fatal(err)
		}
		if entry != nil {
			t.Errorf("expected nil entry, got %v", entry)
		}
	})

	t.Run("GetRange", func(t *testing.T) {
		entries, err := repo.GetRange(ctx, 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Index != 1 {
			t.Errorf("expected first index=1, got %d", entries[0].Index)
		}
		if entries[1].Index != 2 {
			t.Errorf("expected second index=2, got %d", entries[1].Index)
		}
	})

	t.Run("GetByBuyer", func(t *testing.T) {
		entries, err := repo.GetByBuyer(ctx, "buyer_1", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})

	t.Run("GetByTraceID", func(t *testing.T) {
		entries, err := repo.GetByTraceID(ctx, "trace_001")
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].TraceID != "trace_001" {
			t.Errorf("expected trace_001, got %s", entries[0].TraceID)
		}
	})

	t.Run("VerifyChain_valid", func(t *testing.T) {
		valid, err := repo.VerifyChain(ctx, 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !valid {
			t.Error("expected valid chain")
		}
	})

	t.Run("VerifyChain_invalid_tampered", func(t *testing.T) {

		_, err := db.ExecCtx(ctx, `UPDATE audit_log SET hash = 'tampered' WHERE index = 1`)
		if err != nil {
			t.Fatal(err)
		}

		valid, err := repo.VerifyChain(ctx, 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if valid {
			t.Error("expected invalid chain after tampering")
		}

		_, _ = db.ExecCtx(ctx, `DELETE FROM audit_log WHERE index IN (1, 2)`)
	})

	t.Run("GetLatestIndex", func(t *testing.T) {

		e1 := &model.AuditEntry{
			Timestamp: time.Now(), TraceID: "t1", BuyerID: "b1", SessionID: "s1",
			Action: model.AuditActionBrowse, Request: []byte(`{}`), Response: []byte(`{}`),
		}
		_ = repo.Append(ctx, e1)

		e2 := &model.AuditEntry{
			Timestamp: time.Now(), TraceID: "t2", BuyerID: "b1", SessionID: "s1",
			Action: model.AuditActionPurchaseAttempt, Request: []byte(`{}`), Response: []byte(`{}`),
		}
		_ = repo.Append(ctx, e2)

		index, err := repo.GetLatestIndex(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if index != 4 {
			t.Errorf("expected index=4, got %d", index)
		}
	})

	t.Run("PolicyDecision_in_audit", func(t *testing.T) {
		entry := &model.AuditEntry{
			Timestamp: time.Now(),
			TraceID:   "trace_policy",
			BuyerID:   "buyer_1",
			SessionID: "session_1",
			Action:    model.AuditActionPurchaseAllowed,
			Request:   []byte(`{"sku": "SHOE-001"}`),
			Response:  []byte(`{"allowed": true}`),
			PolicyDecision: &model.PolicyDecision{
				Allowed: true, Reason: "within limits", RuleFired: model.RuleFiredNone,
			},
		}

		err := repo.Append(ctx, entry)
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := repo.GetByIndex(ctx, entry.Index)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.PolicyDecision == nil {
			t.Fatal("expected PolicyDecision, got nil")
		}
		if !retrieved.PolicyDecision.Allowed {
			t.Error("expected Allowed=true")
		}
		if retrieved.PolicyDecision.RuleFired != model.RuleFiredNone {
			t.Errorf("expected RuleFired=none, got %s", retrieved.PolicyDecision.RuleFired)
		}
	})
}
