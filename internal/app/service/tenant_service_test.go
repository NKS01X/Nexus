package service

import (
	"context"
	"testing"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/testutil"
)

// TestTenantService_ProvisionTenant tests the ProvisionTenant method.
func TestTenantService_ProvisionTenant(t *testing.T) {
	tests := []struct {
		name       string
		req        model.CreateTenantRequest
		wantErr    bool
		checkFn    func(t *testing.T, tenant *model.Tenant, err error)
	}{
		{
			name: "valid_shopify",
			req: model.CreateTenantRequest{
				Name:     "Test Store",
				Platform: "shopify",
			},
			wantErr: false,
			checkFn: func(t *testing.T, tenant *model.Tenant, err error) {
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tenant.Name != "Test Store" {
					t.Errorf("expected name=Test Store, got %s", tenant.Name)
				}
				if tenant.Platform != "shopify" {
					t.Errorf("expected platform=shopify, got %s", tenant.Platform)
				}
				if tenant.Status != "active" {
					t.Errorf("expected status=active, got %s", tenant.Status)
				}
				if tenant.ID == "" {
					t.Error("expected non-empty ID")
				}
				if tenant.APIKey == "" {
					t.Error("expected non-empty APIKey")
				}
				if len(tenant.APIKey) < 10 {
					t.Errorf("APIKey too short: %s", tenant.APIKey)
				}
			},
		},
		{
			name: "valid_woocommerce",
			req: model.CreateTenantRequest{
				Name:     "Woo Store",
				Platform: "woocommerce",
			},
			wantErr: false,
			checkFn: func(t *testing.T, tenant *model.Tenant, err error) {
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tenant.Platform != "woocommerce" {
					t.Errorf("expected platform=woocommerce, got %s", tenant.Platform)
				}
			},
		},
		{
			name: "valid_custom",
			req: model.CreateTenantRequest{
				Name:     "Custom Store",
				Platform: "custom",
			},
			wantErr: false,
			checkFn: func(t *testing.T, tenant *model.Tenant, err error) {
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tenant.Platform != "custom" {
					t.Errorf("expected platform=custom, got %s", tenant.Platform)
				}
			},
		},
		{
			name: "empty_name",
			req: model.CreateTenantRequest{
				Name:     "",
				Platform: "shopify",
			},
			wantErr: true,
			checkFn: func(t *testing.T, tenant *model.Tenant, err error) {
				if err == nil {
					t.Error("expected error for empty name")
				}
			},
		},
		{
			name: "empty_platform",
			req: model.CreateTenantRequest{
				Name:     "Test Store",
				Platform: "",
			},
			wantErr: true,
			checkFn: func(t *testing.T, tenant *model.Tenant, err error) {
				if err == nil {
					t.Error("expected error for empty platform")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := testutil.NewMockTenantRepo()
			catalogRepo := testutil.NewMockCatalogRepo()

			svc := NewTenantService(tenantRepo, catalogRepo)

			ctx := context.Background()
			tenant, err := svc.ProvisionTenant(ctx, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if tt.checkFn != nil {
				tt.checkFn(t, tenant, err)
			}
		})
	}
}

// TestTenantService_GetTenantByAPIKey tests the GetTenantByAPIKey method.
func TestTenantService_GetTenantByAPIKey(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		tenantRepo := testutil.NewMockTenantRepo()
		catalogRepo := testutil.NewMockCatalogRepo()

		svc := NewTenantService(tenantRepo, catalogRepo)

		// First provision a tenant
		ctx := context.Background()
		created, err := svc.ProvisionTenant(ctx, model.CreateTenantRequest{
			Name:     "Test Store",
			Platform: "shopify",
		})
		if err != nil {
			t.Fatal(err)
		}

		// Now lookup by API key
		found, err := svc.GetTenantByAPIKey(ctx, created.APIKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found == nil {
			t.Fatal("expected tenant, got nil")
		}
		if found.ID != created.ID {
			t.Errorf("expected ID=%s, got %s", created.ID, found.ID)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		tenantRepo := testutil.NewMockTenantRepo()
		catalogRepo := testutil.NewMockCatalogRepo()

		svc := NewTenantService(tenantRepo, catalogRepo)

		found, err := svc.GetTenantByAPIKey(context.Background(), "nonexistent_key")
		if err == nil {
			t.Error("expected error for nonexistent key")
		}
		if found != nil {
			t.Errorf("expected nil, got %v", found)
		}
	})
}

// TestTenantService_ListTenants tests the ListTenants method.
func TestTenantService_ListTenants(t *testing.T) {
	tenantRepo := testutil.NewMockTenantRepo()
	catalogRepo := testutil.NewMockCatalogRepo()

	svc := NewTenantService(tenantRepo, catalogRepo)

	ctx := context.Background()

	// Provision multiple tenants
	for i := 0; i < 3; i++ {
		_, err := svc.ProvisionTenant(ctx, model.CreateTenantRequest{
			Name:     "Store " + string(rune('A'+i)),
			Platform: "shopify",
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	tenants, err := svc.ListTenants(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tenants) != 3 {
		t.Errorf("expected 3 tenants, got %d", len(tenants))
	}
}