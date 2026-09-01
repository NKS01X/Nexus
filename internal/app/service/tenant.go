package service

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// TenantServiceInterface defines the interface for tenant operations.
type TenantServiceInterface interface {
	ProvisionTenant(ctx context.Context, req model.CreateTenantRequest) (*model.Tenant, error)
	GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error)
	ListTenants(ctx context.Context) ([]*model.Tenant, error)
}