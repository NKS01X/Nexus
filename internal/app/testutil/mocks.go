package testutil

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// MockTenantRepo implements repository.TenantRepository for testing.
type MockTenantRepo struct {
	tenants map[string]*model.Tenant
	byKey   map[string]*model.Tenant
}

func NewMockTenantRepo() *MockTenantRepo {
	return &MockTenantRepo{
		tenants: make(map[string]*model.Tenant),
		byKey:   make(map[string]*model.Tenant),
	}
}

func (m *MockTenantRepo) CreateTenant(ctx context.Context, t *model.Tenant) error {
	m.tenants[t.ID] = t
	m.byKey[t.APIKey] = t
	return nil
}

func (m *MockTenantRepo) GetTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *MockTenantRepo) GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error) {
	if t, ok := m.byKey[apiKey]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *MockTenantRepo) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	result := make([]*model.Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		result = append(result, t)
	}
	return result, nil
}

// MockCatalogRepo implements repository.CatalogRepository for testing.
type MockCatalogRepo struct {
	products map[string]*model.Product
}

func NewMockCatalogRepo() *MockCatalogRepo {
	return &MockCatalogRepo{
		products: make(map[string]*model.Product),
	}
}

func (m *MockCatalogRepo) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) SearchProducts(ctx context.Context, filter repository.SearchFilter) ([]*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) GetAllProducts(ctx context.Context) ([]*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) CheckAvailability(ctx context.Context, sku string) (*model.InventoryCheck, error) {
	return nil, nil
}

func (m *MockCatalogRepo) ReserveInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

func (m *MockCatalogRepo) ReleaseInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

func (m *MockCatalogRepo) ConfirmInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

func (m *MockCatalogRepo) InsertProduct(ctx context.Context, p *model.Product, tenantID string) error {
	if m.products == nil {
		m.products = make(map[string]*model.Product)
	}
	key := tenantID + ":" + p.ID
	m.products[key] = p
	return nil
}

func (m *MockCatalogRepo) SearchProductsByTenant(ctx context.Context, filter repository.SearchFilter, tenantID string) ([]*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) GetProductByTenant(ctx context.Context, id string, tenantID string) (*model.Product, error) {
	return nil, nil
}

func (m *MockCatalogRepo) CheckAvailabilityByTenant(ctx context.Context, sku string, tenantID string) (*model.InventoryCheck, error) {
	return nil, nil
}

// MockPolicyRepo implements repository.PolicyRepository for testing.
type MockPolicyRepo struct {
	config       *model.PolicyConfig
	spend        map[string]int64
	skuQty       map[string]int
	requestCount map[string]int
}

func NewMockPolicyRepo() *MockPolicyRepo {
	return &MockPolicyRepo{
		config: &model.PolicyConfig{
			SpendCapPaisa:     300000,
			PerSKUCap:         map[string]int{},
			VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
			AllowedCategories: []string{},
			BlockedSKUs:       []string{},
			GeoRules:          []model.GeoRule{},
		},
		spend:        map[string]int64{},
		skuQty:       map[string]int{},
		requestCount: map[string]int{},
	}
}

func (m *MockPolicyRepo) GetSpend(ctx context.Context, buyerID, sessionID string) (int64, error) {
	key := buyerID + "|" + sessionID
	if v, ok := m.spend[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *MockPolicyRepo) AddSpend(ctx context.Context, buyerID, sessionID string, amount int64) error {
	key := buyerID + "|" + sessionID
	m.spend[key] += amount
	return nil
}

func (m *MockPolicyRepo) RollbackSpend(ctx context.Context, buyerID, sessionID string, amount int64) error {
	key := buyerID + "|" + sessionID
	m.spend[key] -= amount
	if m.spend[key] < 0 {
		m.spend[key] = 0
	}
	return nil
}

func (m *MockPolicyRepo) GetSKUQuantity(ctx context.Context, buyerID, sessionID, sku string) (int, error) {
	key := buyerID + "|" + sessionID + "|" + sku
	if v, ok := m.skuQty[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *MockPolicyRepo) AddSKUQuantity(ctx context.Context, buyerID, sessionID, sku string, qty int) error {
	key := buyerID + "|" + sessionID + "|" + sku
	m.skuQty[key] += qty
	return nil
}

func (m *MockPolicyRepo) RollbackSKUQuantity(ctx context.Context, buyerID, sessionID, sku string, qty int) error {
	key := buyerID + "|" + sessionID + "|" + sku
	m.skuQty[key] -= qty
	if m.skuQty[key] < 0 {
		m.skuQty[key] = 0
	}
	return nil
}

func (m *MockPolicyRepo) GetRequestCount(ctx context.Context, buyerID, sessionID string) (int, error) {
	key := buyerID + "|" + sessionID
	if v, ok := m.requestCount[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *MockPolicyRepo) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	key := buyerID + "|" + sessionID
	m.requestCount[key]++
	return nil
}

func (m *MockPolicyRepo) GetPolicyConfig(ctx context.Context) (*model.PolicyConfig, error) {
	return m.config, nil
}

func (m *MockPolicyRepo) UpdatePolicyConfig(ctx context.Context, cfg *model.PolicyConfig) error {
	m.config = cfg
	return nil
}