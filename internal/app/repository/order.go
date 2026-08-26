package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// OrderRepository defines the interface for order data access.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrder(ctx context.Context, id string) (*model.Order, error)
	GetOrdersByBuyer(ctx context.Context, buyerID string, limit int) ([]*model.Order, error)
	UpdateOrder(ctx context.Context, order *model.Order) error
	GetByIdempotencyKey(ctx context.Context, key string) (*model.Order, error)
}
