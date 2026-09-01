package repository

import (
	"context"

	"github.com/razorpay/aegis/internal/app/model"
)

// TenantRepository defines the interface for tenant data access.
type TenantRepository interface {
	CreateTenant(ctx context.Context, t *model.Tenant) error
	GetTenantByID(ctx context.Context, id string) (*model.Tenant, error)
	GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error)
	ListTenants(ctx context.Context) ([]*model.Tenant, error)
}