package service

import (
	"context"
	"fmt"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// MerchantMCPServiceImpl implements the MerchantMCPService interface.
type MerchantMCPServiceImpl struct {
	catalogRepo repository.CatalogRepository
	orderRepo   repository.OrderRepository
	aegisClient AegisMCPClient
}

// AegisMCPClient defines the interface for calling Aegis Gateway MCP.
type AegisMCPClient interface {
	Purchase(ctx context.Context, params mcp.AegisPurchaseParams) (*mcp.AegisPurchaseResult, error)
}

// NewMerchantMCPService creates a new MerchantMCPServiceImpl.
func NewMerchantMCPService(catalogRepo repository.CatalogRepository, orderRepo repository.OrderRepository, aegisClient AegisMCPClient) *MerchantMCPServiceImpl {
	return &MerchantMCPServiceImpl{
		catalogRepo: catalogRepo,
		orderRepo:   orderRepo,
		aegisClient: aegisClient,
	}
}

// GetOrderRepo returns the order repository.
func (s *MerchantMCPServiceImpl) GetOrderRepo() repository.OrderRepository {
	return s.orderRepo
}

// GetAegisClient returns the Aegis MCP client.
func (s *MerchantMCPServiceImpl) GetAegisClient() AegisMCPClient {
	return s.aegisClient
}

// SearchProducts searches for products in the catalog.
func (s *MerchantMCPServiceImpl) SearchProducts(ctx context.Context, params mcp.SearchProductsParams) (*mcp.SearchProductsResult, error) {
	filter := repository.SearchFilter{
		Query:       params.Query,
		Category:    params.Category,
		MaxPrice:    params.MaxPrice,
		MinPrice:    params.MinPrice,
		InStockOnly: params.InStockOnly,
		Limit:       params.Limit,
		Color:       params.Color,
		Size:        params.Size,
		Brand:       params.Brand,
	}

	products, err := s.catalogRepo.SearchProducts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}

	summaries := make([]model.ProductSummary, len(products))
	for i, p := range products {
		var minPrice int64
		inStock := false
		var firstOfferSKU string
		if len(p.Offers) > 0 {
			minPrice = p.Offers[0].PricePaisa
			for _, o := range p.Offers {
				if o.PricePaisa < minPrice {
					minPrice = o.PricePaisa
				}
				if o.Inventory > o.ReservedCount {
					inStock = true
					if firstOfferSKU == "" {
						firstOfferSKU = o.SKU
					}
				}
			}
		}
		var image string
		if len(p.Images) > 0 {
			image = p.Images[0]
		}

		sku := p.SKU
		if firstOfferSKU != "" {
			sku = firstOfferSKU
		}
		summaries[i] = model.ProductSummary{
			ID:       p.ID,
			SKU:      sku,
			Name:     p.Name,
			Category: p.Category,
			MinPrice: minPrice,
			InStock:  inStock,
			Image:    image,
		}
	}

	return &mcp.SearchProductsResult{Products: summaries}, nil
}

// GetProduct retrieves a product by ID.
func (s *MerchantMCPServiceImpl) GetProduct(ctx context.Context, params mcp.GetProductParams) (*model.Product, error) {
	product, err := s.catalogRepo.GetProduct(ctx, params.ProductID)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

// CheckAvailability checks inventory for a SKU.
func (s *MerchantMCPServiceImpl) CheckAvailability(ctx context.Context, params mcp.CheckAvailabilityParams) (*mcp.CheckAvailabilityResult, error) {
	check, err := s.catalogRepo.CheckAvailability(ctx, params.SKU)
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}

	if check == nil {
		return &mcp.CheckAvailabilityResult{SKU: params.SKU, Available: 0, Reserved: 0}, nil
	}

	return &mcp.CheckAvailabilityResult{
		SKU:       check.SKU,
		Available: check.Available,
		Reserved:  check.Reserved,
	}, nil
}

// Purchase initiates a purchase through Aegis Gateway.
func (s *MerchantMCPServiceImpl) Purchase(ctx context.Context, params mcp.PurchaseParams) (*mcp.PurchaseResult, error) {

	product, err := s.catalogRepo.GetProduct(ctx, params.ProductID)
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	if product == nil {
		return &mcp.PurchaseResult{
			Allowed: false,
			Reason:  "product not found",
			Status:  "NOT_FOUND",
		}, nil
	}

	// Find the offer for the SKU
	var offer *model.Offer
	for i := range product.Offers {
		if product.Offers[i].SKU == params.SKU {
			offer = &product.Offers[i]
			break
		}
	}
	if offer == nil {
		return &mcp.PurchaseResult{
			Allowed: false,
			Reason:  "SKU not found in product offers",
			Status:  "NOT_FOUND",
		}, nil
	}

	amountPaisa := offer.PricePaisa * int64(params.Quantity)

	aegisParams := mcp.AegisPurchaseParams{
		BuyerID:        params.BuyerID,
		SessionID:      params.SessionID,
		ProductID:      params.ProductID,
		SKU:            params.SKU,
		Quantity:       params.Quantity,
		AmountPaisa:    amountPaisa,
		IdempotencyKey: params.IdempotencyKey,
		BuyerPincode:   params.BuyerPincode,
		Metadata:       params.Metadata,
	}

	result, err := s.aegisClient.Purchase(ctx, aegisParams)
	if err != nil {
		return &mcp.PurchaseResult{
			Allowed: false,
			Reason:  fmt.Sprintf("gateway error: %v", err),
			Status:  "ERROR",
		}, nil
	}

	return &mcp.PurchaseResult{
		Allowed: result.Allowed,
		Reason:  result.Reason,
		Status:  result.Status,
		OrderID: result.OrderID,
	}, nil
}

// GetOrderStatus retrieves order status.
func (s *MerchantMCPServiceImpl) GetOrderStatus(ctx context.Context, params mcp.GetOrderStatusParams) (*mcp.GetOrderStatusResult, error) {
	order, err := s.orderRepo.GetOrder(ctx, params.OrderID)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	if order == nil {
		return &mcp.GetOrderStatusResult{
			OrderID: params.OrderID,
			Status:  "NOT_FOUND",
		}, nil
	}

	return &mcp.GetOrderStatusResult{
		OrderID:     order.ID,
		Status:      order.Status,
		AmountPaisa: order.AmountPaisa,
		Currency:    order.Currency,
		Items: []mcp.OrderItem{
			{
				SKU:        order.SKU,
				Quantity:   order.Quantity,
				PricePaisa: order.AmountPaisa,
			},
		},
	}, nil
}
