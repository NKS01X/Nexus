package service

import (
	"context"
	"errors"
	"testing"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

// mockCatalogRepo implements repository.CatalogRepository for testing.
type mockCatalogRepo struct {
	products map[string]*model.Product
	bySKUErr error
}

func newMockCatalogRepo() *mockCatalogRepo {
	return &mockCatalogRepo{
		products: map[string]*model.Product{
			"SHOE-RUN-001-RED-42": {
				ID:       "prod_001",
				SKU:      "SHOE-RUN-001-RED-42",
				Category: "footwear",
			},
			"SHOE-RUN-002-BLU-42": {
				ID:       "prod_002",
				SKU:      "SHOE-RUN-002-BLU-42",
				Category: "footwear",
			},
			"APP-TEE-001": {
				ID:       "prod_003",
				SKU:      "APP-TEE-001",
				Category: "apparel",
			},
			"ELEC-PHONE-001": {
				ID:       "prod_004",
				SKU:      "ELEC-PHONE-001",
				Category: "electronics",
			},
		},
	}
}

func (m *mockCatalogRepo) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	return nil, nil
}

func (m *mockCatalogRepo) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	if m.bySKUErr != nil {
		return nil, m.bySKUErr
	}
	if p, ok := m.products[sku]; ok {
		return p, nil
	}
	return nil, nil
}

func (m *mockCatalogRepo) SearchProducts(ctx context.Context, filter repository.SearchFilter) ([]*model.Product, error) {
	return nil, nil
}

func (m *mockCatalogRepo) GetAllProducts(ctx context.Context) ([]*model.Product, error) {
	return nil, nil
}

func (m *mockCatalogRepo) CheckAvailability(ctx context.Context, sku string) (*model.InventoryCheck, error) {
	return nil, nil
}

func (m *mockCatalogRepo) ReserveInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

func (m *mockCatalogRepo) ReleaseInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

func (m *mockCatalogRepo) ConfirmInventory(ctx context.Context, sku string, quantity int) error {
	return nil
}

// Tenant-scoped methods (stubs for testing)
func (m *mockCatalogRepo) InsertProduct(ctx context.Context, p *model.Product, tenantID string) error {
	return nil
}

func (m *mockCatalogRepo) SearchProductsByTenant(ctx context.Context, filter repository.SearchFilter, tenantID string) ([]*model.Product, error) {
	return nil, nil
}

func (m *mockCatalogRepo) GetProductByTenant(ctx context.Context, id string, tenantID string) (*model.Product, error) {
	return nil, nil
}

func (m *mockCatalogRepo) CheckAvailabilityByTenant(ctx context.Context, sku string, tenantID string) (*model.InventoryCheck, error) {
	return nil, nil
}

// TestPolicyEngine_Evaluate tests all policy evaluation scenarios.
func TestPolicyEngine_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func() *model.PolicyConfig
		setupState     func(repo *mockPolicyRepo)
		request        *model.PurchaseRequest
		expectedAllow  bool
		expectedRule   string
		expectedReason string
	}{
		{
			name: "allowed - all checks pass",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{"SHOE-RUN-001-RED-42": 2},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{"footwear", "apparel"},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{{Country: "IN", Allowed: true, Pincodes: []string{"560001", "560002"}}},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
				BuyerPincode:   "560001",
			},
			expectedAllow: true,
			expectedRule:  model.RuleFiredNone,
		},
		{
			name: "blocked - spend cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 295000
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    10000,
				IdempotencyKey: "idem_1",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredSpendCap,
			expectedReason: "exceeds session spend cap",
		},
		{
			name: "blocked - per sku cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{"SHOE-RUN-001-RED-42": 2},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 2
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredPerSKUCap,
			expectedReason: "exceeds per-SKU cap",
		},
		{
			name: "blocked - velocity cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 3, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 3
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredVelocityCap,
			expectedReason: "exceeds velocity cap",
		},
		{
			name: "blocked - category not allowed",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{"footwear", "apparel"},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|ELEC-PHONE-001"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_004",
				SKU:            "ELEC-PHONE-001",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredCategoryBlocked,
			expectedReason: "not in allowed categories",
		},
		{
			name: "blocked - sku blocked",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{"SHOE-RUN-001-RED-42"},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredSKUBlocked,
			expectedReason: "is blocked",
		},
		{
			name: "blocked - geo restriction",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{{Country: "IN", Allowed: true, Pincodes: []string{"560001", "560002"}}},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
				BuyerPincode:   "400001",
			},
			expectedAllow:  false,
			expectedRule:   model.RuleFiredGeoRestricted,
			expectedReason: "not allowed by geo rules",
		},
		{
			name: "boundary - spend exactly at cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 250000
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    50000,
				IdempotencyKey: "idem_1",
			},
			expectedAllow: true,
			expectedRule:  model.RuleFiredNone,
		},
		{
			name: "boundary - sku exactly at cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{"SHOE-RUN-001-RED-42": 2},
					VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 1
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow: true,
			expectedRule:  model.RuleFiredNone,
		},
		{
			name: "boundary - velocity exactly at cap",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     300000,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 1, WindowSeconds: 60},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       1,
				AmountPaisa:    249900,
				IdempotencyKey: "idem_1",
			},
			expectedAllow: true,
			expectedRule:  model.RuleFiredNone,
		},
		{
			name: "empty config - all allowed",
			setupConfig: func() *model.PolicyConfig {
				return &model.PolicyConfig{
					SpendCapPaisa:     0,
					PerSKUCap:         map[string]int{},
					VelocityCap:       model.VelocityLimit{MaxRequests: 0, WindowSeconds: 0},
					AllowedCategories: []string{},
					BlockedSKUs:       []string{},
					GeoRules:          []model.GeoRule{},
				}
			},
			setupState: func(repo *mockPolicyRepo) {
				repo.spend["buyer_1|session_1"] = 0
				repo.skuQty["buyer_1|session_1|SHOE-RUN-001-RED-42"] = 0
				repo.requestCount["buyer_1|session_1"] = 0
			},
			request: &model.PurchaseRequest{
				BuyerID:        "buyer_1",
				SessionID:      "session_1",
				ProductID:      "prod_001",
				SKU:            "SHOE-RUN-001-RED-42",
				Quantity:       100,
				AmountPaisa:    999999999,
				IdempotencyKey: "idem_1",
			},
			expectedAllow: true,
			expectedRule:  model.RuleFiredNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockPolicyRepo()
			if tt.setupConfig != nil {
				repo.config = tt.setupConfig()
			}
			if tt.setupState != nil {
				tt.setupState(repo)
			}

			engine := NewPolicyEngine(repo, newMockCatalogRepo())
			decision, err := engine.Evaluate(context.Background(), tt.request)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if decision == nil {
				t.Fatal("expected decision, got nil")
			}
			if decision.Allowed != tt.expectedAllow {
				t.Errorf("expected allow=%v, got %v (reason: %s)", tt.expectedAllow, decision.Allowed, decision.Reason)
			}
			if decision.RuleFired != tt.expectedRule {
				t.Errorf("expected rule=%s, got %s", tt.expectedRule, decision.RuleFired)
			}
			if tt.expectedReason != "" && !containsString(decision.Reason, tt.expectedReason) {
				t.Errorf("expected reason to contain '%s', got: %s", tt.expectedReason, decision.Reason)
			}
		})
	}
}

// TestPolicyEngine_Evaluate_InvalidRequests tests validation errors.
func TestPolicyEngine_Evaluate_InvalidRequests(t *testing.T) {
	repo := newMockPolicyRepo()
	engine := NewPolicyEngine(repo, newMockCatalogRepo())

	tests := []struct {
		name    string
		request *model.PurchaseRequest
	}{
		{
			name:    "nil request",
			request: nil,
		},
		{
			name: "missing buyer_id",
			request: &model.PurchaseRequest{
				SessionID:   "session_1",
				SKU:         "SHOE-RUN-001",
				Quantity:    1,
				AmountPaisa: 100,
			},
		},
		{
			name: "missing sku",
			request: &model.PurchaseRequest{
				BuyerID:     "buyer_1",
				SessionID:   "session_1",
				Quantity:    1,
				AmountPaisa: 100,
			},
		},
		{
			name: "zero quantity",
			request: &model.PurchaseRequest{
				BuyerID:     "buyer_1",
				SessionID:   "session_1",
				SKU:         "SHOE-RUN-001",
				Quantity:    0,
				AmountPaisa: 100,
			},
		},
		{
			name: "negative amount",
			request: &model.PurchaseRequest{
				BuyerID:     "buyer_1",
				SessionID:   "session_1",
				SKU:         "SHOE-RUN-001",
				Quantity:    1,
				AmountPaisa: -100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Evaluate(context.Background(), tt.request)
			if err == nil {
				t.Error("expected error for invalid request")
			}
			if !errors.Is(err, ErrInvalidRequest) {
				t.Errorf("expected ErrInvalidRequest, got %v", err)
			}
		})
	}
}

// TestPolicyEngine_Evaluate_ProductNotFound tests product lookup failure.
func TestPolicyEngine_Evaluate_ProductNotFound(t *testing.T) {
	repo := newMockPolicyRepo()
	catalog := newMockCatalogRepo()
	catalog.products = map[string]*model.Product{}

	engine := NewPolicyEngine(repo, catalog)
	req := &model.PurchaseRequest{
		BuyerID:     "buyer_1",
		SessionID:   "session_1",
		SKU:         "NONEXISTENT-SKU",
		Quantity:    1,
		AmountPaisa: 100,
	}

	_, err := engine.Evaluate(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for product not found")
	}
}

// TestPolicyEngine_RecordSpend tests spend recording.
// Note: Request count is incremented in gateway service, not here.
func TestPolicyEngine_RecordSpend(t *testing.T) {
	repo := newMockPolicyRepo()
	engine := NewPolicyEngine(repo, newMockCatalogRepo())

	err := engine.RecordSpend(context.Background(), "buyer_1", "session_1", "SKU-001", 2, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.spend["buyer_1|session_1"] != 5000 {
		t.Errorf("expected spend=5000, got %d", repo.spend["buyer_1|session_1"])
	}
	if repo.skuQty["buyer_1|session_1|SKU-001"] != 2 {
		t.Errorf("expected sku qty=2, got %d", repo.skuQty["buyer_1|session_1|SKU-001"])
	}

	if repo.requestCount["buyer_1|session_1"] != 0 {
		t.Errorf("expected request count=0 (incremented in gateway), got %d", repo.requestCount["buyer_1|session_1"])
	}
}

// TestPolicyEngine_RollbackSpend tests spend rollback.
func TestPolicyEngine_RollbackSpend(t *testing.T) {
	repo := newMockPolicyRepo()
	engine := NewPolicyEngine(repo, newMockCatalogRepo())

	repo.requestCount["buyer_1|session_1"] = 1

	err := engine.RecordSpend(context.Background(), "buyer_1", "session_1", "SKU-001", 2, 5000)
	if err != nil {
		t.Fatalf("record error: %v", err)
	}

	err = engine.RollbackSpend(context.Background(), "buyer_1", "session_1", "SKU-001", 2, 5000)
	if err != nil {
		t.Fatalf("rollback error: %v", err)
	}

	if repo.spend["buyer_1|session_1"] != 0 {
		t.Errorf("expected spend=0 after rollback, got %d", repo.spend["buyer_1|session_1"])
	}
	if repo.skuQty["buyer_1|session_1|SKU-001"] != 0 {
		t.Errorf("expected sku qty=0 after rollback, got %d", repo.skuQty["buyer_1|session_1|SKU-001"])
	}

	if repo.requestCount["buyer_1|session_1"] != 1 {
		t.Errorf("expected request count=1 (not rolled back), got %d", repo.requestCount["buyer_1|session_1"])
	}
}

// mockPolicyRepo implements repository.PolicyRepository for testing.
type mockPolicyRepo struct {
	config       *model.PolicyConfig
	spend        map[string]int64
	skuQty       map[string]int
	requestCount map[string]int
}

func newMockPolicyRepo() *mockPolicyRepo {
	return &mockPolicyRepo{
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

func (m *mockPolicyRepo) GetSpend(ctx context.Context, buyerID, sessionID string) (int64, error) {
	key := buyerID + "|" + sessionID
	if v, ok := m.spend[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockPolicyRepo) AddSpend(ctx context.Context, buyerID, sessionID string, amountPaisa int64) error {
	key := buyerID + "|" + sessionID
	m.spend[key] += amountPaisa
	return nil
}

func (m *mockPolicyRepo) GetSKUQuantity(ctx context.Context, buyerID, sessionID, sku string) (int, error) {
	key := buyerID + "|" + sessionID + "|" + sku
	if v, ok := m.skuQty[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockPolicyRepo) AddSKUQuantity(ctx context.Context, buyerID, sessionID, sku string, quantity int) error {
	key := buyerID + "|" + sessionID + "|" + sku
	m.skuQty[key] += quantity
	return nil
}

func (m *mockPolicyRepo) GetRequestCount(ctx context.Context, buyerID, sessionID string, windowSeconds int) (int, error) {
	key := buyerID + "|" + sessionID
	if v, ok := m.requestCount[key]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *mockPolicyRepo) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	key := buyerID + "|" + sessionID
	m.requestCount[key]++
	return nil
}

func (m *mockPolicyRepo) GetPolicyConfig(ctx context.Context) (*model.PolicyConfig, error) {
	return m.config, nil
}

func (m *mockPolicyRepo) UpdatePolicyConfig(ctx context.Context, config *model.PolicyConfig) error {
	m.config = config
	return nil
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
