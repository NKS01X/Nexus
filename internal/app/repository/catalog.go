package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// CatalogRepository defines the interface for catalog data access.
type CatalogRepository interface {
	GetProduct(ctx context.Context, id string) (*model.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*model.Product, error)
	SearchProducts(ctx context.Context, filter SearchFilter) ([]*model.Product, error)
	GetAllProducts(ctx context.Context) ([]*model.Product, error)
	CheckAvailability(ctx context.Context, sku string) (*model.InventoryCheck, error)
	ReserveInventory(ctx context.Context, sku string, quantity int) error
	ReleaseInventory(ctx context.Context, sku string, quantity int) error
	ConfirmInventory(ctx context.Context, sku string, quantity int) error
}

// SearchFilter holds search parameters for product queries.
type SearchFilter struct {
	Query       string
	Category    string
	MaxPrice    *int64
	MinPrice    *int64
	InStockOnly bool
	Limit       int
	Color       string
	Size        string
	Brand       string
}
