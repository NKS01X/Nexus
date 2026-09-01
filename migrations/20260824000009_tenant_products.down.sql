-- Migration: Remove tenant_id from products and offers
-- Created: 2026-08-24

ALTER TABLE products DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE offers   DROP COLUMN IF EXISTS tenant_id;