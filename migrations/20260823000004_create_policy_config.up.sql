-- Migration: Create policy_config table
-- Created: 2026-08-23

CREATE TABLE policy_config (
    id                      INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    spend_cap_paisa         BIGINT NOT NULL DEFAULT 300000,
    per_sku_cap             JSONB NOT NULL DEFAULT '{}',
    velocity_max_requests   INT NOT NULL DEFAULT 10,
    velocity_window_seconds INT NOT NULL DEFAULT 60,
    allowed_categories      JSONB NOT NULL DEFAULT '[]',
    blocked_skus            JSONB NOT NULL DEFAULT '[]',
    geo_rules               JSONB NOT NULL DEFAULT '[]',
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure single row
INSERT INTO policy_config (id) VALUES (1) ON CONFLICT DO NOTHING;