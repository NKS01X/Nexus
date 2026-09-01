-- Migration: Create tenants table
-- Created: 2026-08-24

CREATE TABLE tenants (
    id           VARCHAR(64)  PRIMARY KEY,          -- e.g. "store_abc123"
    name         VARCHAR(255) NOT NULL,
    platform     VARCHAR(32)  NOT NULL,             -- "shopify" | "woocommerce" | "custom"
    api_key      VARCHAR(128) UNIQUE NOT NULL,      -- bearer token for MCP auth
    status       VARCHAR(32)  NOT NULL DEFAULT 'active',
    config       JSONB        NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tenants_api_key ON tenants(api_key);
CREATE INDEX idx_tenants_status  ON tenants(status);