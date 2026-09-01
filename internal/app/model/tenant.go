package model

import (
	"errors"
	"time"
)

var ErrTenantNotFound = errors.New("tenant not found")

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"` // "shopify" | "woocommerce" | "custom"
	APIKey    string    `json:"api_key"`
	Status    string    `json:"status"`
	Config    []byte    `json:"config"`   // raw JSONB
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	Name     string `json:"name"`
	Platform string `json:"platform"`
}