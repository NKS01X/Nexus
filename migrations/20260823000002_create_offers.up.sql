-- Migration: Create offers table
-- Created: 2026-08-23

CREATE TABLE offers (
    id              VARCHAR(64) PRIMARY KEY,
    product_id      VARCHAR(64) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku             VARCHAR(64) UNIQUE NOT NULL,
    price_paisa     BIGINT NOT NULL CHECK (price_paisa >= 0),
    currency        VARCHAR(3) NOT NULL DEFAULT 'INR',
    inventory       INT NOT NULL DEFAULT 0 CHECK (inventory >= 0),
    reserved_count  INT NOT NULL DEFAULT 0 CHECK (reserved_count >= 0),
    size            VARCHAR(32),
    color           VARCHAR(32),
    valid_from      TIMESTAMPTZ,
    valid_until     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offers_product_id ON offers(product_id);
CREATE INDEX idx_offers_sku ON offers(sku);
CREATE INDEX idx_offers_valid_window ON offers(valid_from, valid_until);