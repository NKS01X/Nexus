-- Migration: Create orders table
-- Created: 2026-08-23

CREATE TABLE orders (
    id                  VARCHAR(64) PRIMARY KEY,
    buyer_id            VARCHAR(64) NOT NULL,
    session_id          VARCHAR(64) NOT NULL,
    product_id          VARCHAR(64) NOT NULL,
    sku                 VARCHAR(64) NOT NULL,
    quantity            INT NOT NULL CHECK (quantity > 0),
    amount_paisa        BIGINT NOT NULL CHECK (amount_paisa >= 0),
    currency            VARCHAR(3) NOT NULL DEFAULT 'INR',
    status              VARCHAR(16) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','PAID','FAILED','REFUNDED','CANCELLED')),
    razorpay_order_id   VARCHAR(64),
    razorpay_payment_id VARCHAR(64),
    idempotency_key     VARCHAR(64) UNIQUE NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_buyer ON orders(buyer_id);
CREATE INDEX idx_orders_idempotency ON orders(idempotency_key);
CREATE INDEX idx_orders_status ON orders(status);