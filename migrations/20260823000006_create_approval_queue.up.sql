-- Migration: Create approval_queue table
-- Created: 2026-08-23

CREATE TABLE approval_queue (
    id                  VARCHAR(64) PRIMARY KEY,
    buyer_id            VARCHAR(64) NOT NULL,
    session_id          VARCHAR(64) NOT NULL,
    purchase_request    JSONB NOT NULL,
    policy_decision     JSONB NOT NULL,
    buyer_reasoning     TEXT,
    status              VARCHAR(16) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','REJECTED','EXPIRED')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at         TIMESTAMPTZ,
    reviewer_id         VARCHAR(64),
    review_note         TEXT,
    expires_at          TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_approval_queue_status ON approval_queue(status);
CREATE INDEX idx_approval_queue_buyer ON approval_queue(buyer_id);
CREATE INDEX idx_approval_queue_expires ON approval_queue(expires_at);