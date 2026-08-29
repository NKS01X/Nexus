package service

import (
	"context"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
)

// MerchantMCPService defines the interface for the Merchant MCP server tools.
type MerchantMCPService interface {
	SearchProducts(ctx context.Context, params mcp.SearchProductsParams) (*mcp.SearchProductsResult, error)
	GetProduct(ctx context.Context, params mcp.GetProductParams) (*model.Product, error)
	CheckAvailability(ctx context.Context, params mcp.CheckAvailabilityParams) (*mcp.CheckAvailabilityResult, error)
	Purchase(ctx context.Context, params mcp.PurchaseParams) (*mcp.PurchaseResult, error)
	GetOrderStatus(ctx context.Context, params mcp.GetOrderStatusParams) (*mcp.GetOrderStatusResult, error)
}
