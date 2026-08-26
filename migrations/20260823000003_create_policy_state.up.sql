-- Migration: Create policy_state table
-- Created: 2026-08-23

CREATE TABLE policy_state (
    buyer_id        VARCHAR(64) NOT NULL,
    session_id      VARCHAR(64) NOT NULL,
    sku             VARCHAR(64) NOT NULL,  -- '*' for spend/velocity rows
    spend_paisa     BIGINT NOT NULL DEFAULT 0 CHECK (spend_paisa >= 0),
    sku_quantity    INT NOT NULL DEFAULT 0 CHECK (sku_quantity >= 0),
    request_count   INT NOT NULL DEFAULT 0 CHECK (request_count >= 0),
    window_start    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (buyer_id, session_id, sku)
);

CREATE INDEX idx_policy_state_buyer_session ON policy_state(buyer_id, session_id);