-- Migration: Create audit_log table
-- Created: 2026-08-23

CREATE TABLE audit_log (
    index           BIGSERIAL PRIMARY KEY,
    prev_hash       VARCHAR(64) NOT NULL DEFAULT '',
    hash            VARCHAR(64) NOT NULL DEFAULT '',
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trace_id        VARCHAR(64) NOT NULL,
    buyer_id        VARCHAR(64),
    session_id      VARCHAR(64),
    action          VARCHAR(32) NOT NULL,
    policy_decision JSONB,
    request         JSONB,
    response        JSONB,
    buyer_reasoning TEXT,
    error           TEXT
);

CREATE INDEX idx_audit_log_buyer ON audit_log(buyer_id);
CREATE INDEX idx_audit_log_trace ON audit_log(trace_id);
CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);