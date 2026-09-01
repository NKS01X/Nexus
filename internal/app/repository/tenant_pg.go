package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/razorpay/aegis/internal/app/model"
)

// TenantPG implements TenantRepository using PostgreSQL.
type TenantPG struct {
	db *DB
}

// NewTenantPG creates a new PostgreSQL tenant repository.
func NewTenantPG(db *DB) *TenantPG {
	return &TenantPG{db: db}
}

// CreateTenant creates a new tenant.
func (r *TenantPG) CreateTenant(ctx context.Context, t *model.Tenant) error {
	configJSON, err := json.Marshal(t.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	const query = `
		INSERT INTO tenants (id, name, platform, api_key, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = r.db.ExecCtx(ctx, query, t.ID, t.Name, t.Platform, t.APIKey, t.Status, configJSON)
	if err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

// GetTenantByID retrieves a tenant by ID.
func (r *TenantPG) GetTenantByID(ctx context.Context, id string) (*model.Tenant, error) {
	const query = `
		SELECT id, name, platform, api_key, status, config, created_at, updated_at
		FROM tenants WHERE id = $1
	`
	var t model.Tenant
	var configJSON []byte
	err := r.db.QueryRowCtx(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Platform, &t.APIKey, &t.Status, &configJSON, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant by id: %w", err)
	}
	t.Config = configJSON
	return &t, nil
}

// GetTenantByAPIKey retrieves a tenant by API key.
func (r *TenantPG) GetTenantByAPIKey(ctx context.Context, apiKey string) (*model.Tenant, error) {
	const query = `
		SELECT id, name, platform, api_key, status, config, created_at, updated_at
		FROM tenants WHERE api_key = $1
	`
	var t model.Tenant
	var configJSON []byte
	err := r.db.QueryRowCtx(ctx, query, apiKey).Scan(
		&t.ID, &t.Name, &t.Platform, &t.APIKey, &t.Status, &configJSON, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tenant by api key: %w", err)
	}
	t.Config = configJSON
	return &t, nil
}

// ListTenants returns all tenants.
func (r *TenantPG) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	const query = `
		SELECT id, name, platform, api_key, status, config, created_at, updated_at
		FROM tenants ORDER BY created_at DESC
	`
	rows, err := r.db.QueryCtx(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*model.Tenant
	for rows.Next() {
		var t model.Tenant
		var configJSON []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.Platform, &t.APIKey, &t.Status, &configJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}
		t.Config = configJSON
		tenants = append(tenants, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}
	return tenants, nil
}