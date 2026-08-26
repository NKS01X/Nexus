package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

// OrderPG implements OrderRepository using PostgreSQL.
type OrderPG struct {
	db *DB
}

// NewOrderPG creates a new PostgreSQL order repository.
func NewOrderPG(db *DB) *OrderPG {
	return &OrderPG{db: db}
}

// CreateOrder creates a new order with idempotency key protection.
func (r *OrderPG) CreateOrder(ctx context.Context, order *model.Order) error {
	const query = `
		INSERT INTO orders (id, buyer_id, session_id, product_id, sku, quantity, amount_paisa,
		                    currency, status, razorpay_order_id, razorpay_payment_id, idempotency_key,
		                    created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	_, err := r.db.ExecCtx(ctx, query,
		order.ID, order.BuyerID, order.SessionID, order.ProductID, order.SKU,
		order.Quantity, order.AmountPaisa, order.Currency, order.Status,
		order.RazorpayOrderID, order.RazorpayPaymentID, order.IdempotencyKey,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	return nil
}

// GetOrder retrieves an order by ID.
func (r *OrderPG) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	const query = `
		SELECT id, buyer_id, session_id, product_id, sku, quantity, amount_paisa,
		       currency, status, razorpay_order_id, razorpay_payment_id, idempotency_key,
		       created_at, updated_at
		FROM orders WHERE id = $1
	`
	return r.scanOrder(r.db.QueryRowCtx(ctx, query, id))
}

// GetOrdersByBuyer retrieves orders for a specific buyer.
func (r *OrderPG) GetOrdersByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.Order, error) {
	const query = `
		SELECT id, buyer_id, session_id, product_id, sku, quantity, amount_paisa,
		       currency, status, razorpay_order_id, razorpay_payment_id, idempotency_key,
		       created_at, updated_at
		FROM orders WHERE buyer_id = $1 ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.QueryCtx(ctx, query, buyerID, limit)
	if err != nil {
		return nil, fmt.Errorf("get orders by buyer: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order, err := r.scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return orders, nil
}

// UpdateOrder updates an existing order.
func (r *OrderPG) UpdateOrder(ctx context.Context, order *model.Order) error {
	const query = `
		UPDATE orders
		SET status = $1, razorpay_order_id = $2, razorpay_payment_id = $3, updated_at = $4
		WHERE id = $5
	`
	result, err := r.db.ExecCtx(ctx, query,
		order.Status, order.RazorpayOrderID, order.RazorpayPaymentID,
		time.Now(), order.ID,
	)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrOrderNotFound
	}
	return nil
}

// GetByIdempotencyKey retrieves an order by its idempotency key.
func (r *OrderPG) GetByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	const query = `
		SELECT id, buyer_id, session_id, product_id, sku, quantity, amount_paisa,
		       currency, status, razorpay_order_id, razorpay_payment_id, idempotency_key,
		       created_at, updated_at
		FROM orders WHERE idempotency_key = $1
	`
	return r.scanOrder(r.db.QueryRowCtx(ctx, query, key))
}

// scanOrder scans an order from a Row or Rows.
func (r *OrderPG) scanOrder(scanner interface{ Scan(dest ...any) error }) (*model.Order, error) {
	var o model.Order
	var razorpayOrderID, razorpayPaymentID sql.NullString

	err := scanner.Scan(
		&o.ID, &o.BuyerID, &o.SessionID, &o.ProductID, &o.SKU,
		&o.Quantity, &o.AmountPaisa, &o.Currency, &o.Status,
		&razorpayOrderID, &razorpayPaymentID, &o.IdempotencyKey,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}

	if razorpayOrderID.Valid {
		o.RazorpayOrderID = razorpayOrderID.String
	}
	if razorpayPaymentID.Valid {
		o.RazorpayPaymentID = razorpayPaymentID.String
	}

	return &o, nil
}
