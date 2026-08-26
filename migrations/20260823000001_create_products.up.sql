-- Migration: Create products table
-- Created: 2026-08-23

CREATE TABLE products (
    id              VARCHAR(64) PRIMARY KEY,
    sku             VARCHAR(64) UNIQUE NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(64) NOT NULL,
    images          JSONB DEFAULT '[]',
    attributes      JSONB DEFAULT '{}',
    reviews         JSONB DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_sku ON products(sku);