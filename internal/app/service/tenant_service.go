package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// TenantService handles tenant provisioning and management.
type TenantService struct {
	tenantRepo  repository.TenantRepository
	catalogRepo repository.CatalogRepository
}

// NewTenantService creates a new TenantService.
func NewTenantService(tenantRepo repository.TenantRepository, catalogRepo repository.CatalogRepository) *TenantService {
	return &TenantService{tenantRepo: tenantRepo, catalogRepo: catalogRepo}
}

// ProvisionTenant creates a new tenant, seeds their catalog, returns the tenant.
func (s *TenantService) ProvisionTenant(ctx context.Context, req model.CreateTenantRequest) (*model.Tenant, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Platform == "" {
		return nil, fmt.Errorf("platform is required")
	}

	id := "store_" + shortID()
	apiKey := generateAPIKey()
	t := &model.Tenant{
		ID:       id,
		Name:     req.Name,
		Platform: req.Platform,
		APIKey:   apiKey,
		Status:   "active",
	}
	if err := s.tenantRepo.CreateTenant(ctx, t); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}
	// Seed a demo catalog for this tenant so the MCP endpoint works immediately.
	if err := s.seedDemoCatalog(ctx, id, req.Platform); err != nil {
		return nil, fmt.Errorf("seed catalog: %w", err)
	}
	return t, nil
}

// GetTenantByAPIKey resolves a bearer token to a tenant.
func (s *TenantService) GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error) {
	t, err := s.tenantRepo.GetTenantByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, model.ErrTenantNotFound
	}
	return t, nil
}

// ListTenants returns all tenants.
func (s *TenantService) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	return s.tenantRepo.ListTenants(ctx)
}

// seedDemoCatalog inserts 3-5 demo products scoped to tenantID.
// Products are inserted directly via catalogRepo; use a new method InsertProduct.
// For the hackathon, hardcode 3 products per platform type (Shopify -> sneakers,
// WooCommerce -> electronics, Custom -> books). Each product gets a single Offer.
func (s *TenantService) seedDemoCatalog(ctx context.Context, tenantID string, platform string) error {
	// Use a short suffix derived from tenantID to make all IDs and SKUs globally unique.
	// This prevents unique-constraint violations when multiple merchants of the same
	// platform type are provisioned (e.g. two Shopify stores).
	suffix := tenantID // e.g. "store_a1b2c3d4"

	type catalogEntry struct {
		baseProdID  string
		baseSKU     string
		name        string
		description string
		category    string
		images      []string
		attributes  json.RawMessage
		reviews     []model.Review
		offers      []struct {
			baseOfferID string
			baseSKU     string
			pricePaisa  int64
			inventory   int
			size        string
			color       string
		}
	}

	var catalog []catalogEntry

	switch platform {
	case "shopify":
		catalog = []catalogEntry{
			{
				baseProdID: "prod_shopify_001", baseSKU: "SHOE-RUN-001",
				name: "Air Runner Pro", description: "Lightweight running shoes with responsive cushioning",
				category: "footwear", images: []string{"https://example.com/shoe1.jpg"},
				attributes: json.RawMessage(`{"brand": "Nike", "type": "running"}`),
				reviews: []model.Review{{ID: "rev_001", Rating: 5, Title: "Great!", Body: "Amazing quality!", Author: "RunnerJoe", CreatedAt: time.Now()}},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_shopify_001", "SHOE-RUN-001-RED-42", 89900, 50, "42", "Red"}},
			},
			{
				baseProdID: "prod_shopify_002", baseSKU: "SHOE-TRAIL-001",
				name: "Trail Blazer", description: "Rugged trail shoes with superior grip",
				category: "footwear", images: []string{"https://example.com/shoe2.jpg"},
				attributes: json.RawMessage(`{"brand": "Salomon", "type": "trail"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_shopify_002", "SHOE-TRAIL-001-BLK-43", 129900, 30, "43", "Black"}},
			},
			{
				baseProdID: "prod_shopify_003", baseSKU: "APPAREL-TEE-001",
				name: "Performance Tee", description: "Moisture-wicking athletic t-shirt",
				category: "apparel", images: []string{"https://example.com/tee.jpg"},
				attributes: json.RawMessage(`{"brand": "Under Armour", "type": "tee"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_shopify_003", "APPAREL-TEE-001-BLU-M", 24900, 100, "M", "Blue"}},
			},
		}
	case "woocommerce":
		catalog = []catalogEntry{
			{
				baseProdID: "prod_woo_001", baseSKU: "ELEC-PHONE-001",
				name: "SmartPhone X", description: "Latest flagship smartphone with AI camera",
				category: "electronics", images: []string{"https://example.com/phone.jpg"},
				attributes: json.RawMessage(`{"brand": "TechCorp", "type": "smartphone"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_woo_001", "ELEC-PHONE-001-128-BLK", 699900, 20, "", "Black"}},
			},
			{
				baseProdID: "prod_woo_002", baseSKU: "ELEC-LAPTOP-001",
				name: "UltraBook Pro 14", description: "Lightweight laptop for professionals",
				category: "electronics", images: []string{"https://example.com/laptop.jpg"},
				attributes: json.RawMessage(`{"brand": "ComputeCo", "type": "laptop"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_woo_002", "ELEC-LAPTOP-001-512-SLV", 12990000, 10, "", "Silver"}},
			},
			{
				baseProdID: "prod_woo_003", baseSKU: "ELEC-EARBUD-001",
				name: "Wireless Earbuds Pro", description: "Noise-cancelling true wireless earbuds",
				category: "electronics", images: []string{"https://example.com/earbuds.jpg"},
				attributes: json.RawMessage(`{"brand": "SoundMax", "type": "earbuds"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_woo_003", "ELEC-EARBUD-001-WHT", 199000, 40, "", "White"}},
			},
		}
	case "custom":
		catalog = []catalogEntry{
			{
				baseProdID: "prod_custom_001", baseSKU: "BOOK-FICTION-001",
				name: "The AI Revolution", description: "A thrilling sci-fi novel about artificial intelligence",
				category: "books", images: []string{"https://example.com/book1.jpg"},
				attributes: json.RawMessage(`{"author": "Jane Doe", "genre": "sci-fi"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_custom_001", "BOOK-FICTION-001-HB", 49900, 200, "", ""}},
			},
			{
				baseProdID: "prod_custom_002", baseSKU: "BOOK-NONFICTION-001",
				name: "Building AI Products", description: "Practical guide to building AI-powered applications",
				category: "books", images: []string{"https://example.com/book2.jpg"},
				attributes: json.RawMessage(`{"author": "John Smith", "genre": "technology"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_custom_002", "BOOK-NONFICTION-001-PB", 34900, 150, "", ""}},
			},
			{
				baseProdID: "prod_custom_003", baseSKU: "BOOK-BIOGRAPHY-001",
				name: "Code & Dreams", description: "Biography of a tech visionary",
				category: "books", images: []string{"https://example.com/book3.jpg"},
				attributes: json.RawMessage(`{"author": "Alice Brown", "genre": "biography"}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_custom_003", "BOOK-BIOGRAPHY-001-HB", 59900, 100, "", ""}},
			},
		}
	default:
		catalog = []catalogEntry{
			{
				baseProdID: "prod_default_001", baseSKU: "DEFAULT-ITEM-001",
				name: "Default Product", description: "A default product for unknown platforms",
				category: "general", images: []string{},
				attributes: json.RawMessage(`{}`),
				reviews: []model.Review{},
				offers: []struct {
					baseOfferID string; baseSKU string; pricePaisa int64; inventory int; size string; color string
				}{{"offer_default_001", "DEFAULT-ITEM-001", 10000, 10, "", ""}},
			},
		}
	}

	for _, entry := range catalog {
		// Scope all IDs and SKUs to this tenant to avoid unique-constraint conflicts.
		prodID := entry.baseProdID + "_" + suffix
		prodSKU := entry.baseSKU + "_" + suffix

		offers := make([]model.Offer, 0, len(entry.offers))
		for _, o := range entry.offers {
			offers = append(offers, model.Offer{
				ID:            o.baseOfferID + "_" + suffix,
				ProductID:     prodID,
				SKU:           o.baseSKU + "_" + suffix,
				PricePaisa:    o.pricePaisa,
				Currency:      "INR",
				Inventory:     o.inventory,
				ReservedCount: 0,
				Size:          o.size,
				Color:         o.color,
			})
		}

		p := &model.Product{
			ID:          prodID,
			SKU:         prodSKU,
			Name:        entry.name,
			Description: entry.description,
			Category:    entry.category,
			Images:      entry.images,
			Attributes:  entry.attributes,
			Reviews:     entry.reviews,
			Offers:      offers,
		}
		if err := s.catalogRepo.InsertProduct(ctx, p, tenantID); err != nil {
			return fmt.Errorf("insert product %s: %w", p.ID, err)
		}
	}
	return nil
}

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "nexus_" + hex.EncodeToString(b)
}