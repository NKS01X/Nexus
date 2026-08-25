package model

import (
	"encoding/json"
	"errors"
	"time"
)

// Sentinel errors for catalog operations.
var (
	ErrInsufficientInventory = errors.New("insufficient inventory available")
	ErrInventoryNotReserved  = errors.New("inventory not reserved")
)

// Product represents a product in the merchant catalog.
type Product struct {
	ID          string          `json:"id" db:"id"`
	SKU         string          `json:"sku" db:"sku"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Category    string          `json:"category" db:"category"`
	Images      []string        `json:"images" db:"images"`
	Attributes  json.RawMessage `json:"attributes" db:"attributes"`
	Offers      []Offer         `json:"offers" db:"-"`
	Reviews     []Review        `json:"reviews,omitempty" db:"reviews"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// Offer represents a product variant with pricing and inventory.
type Offer struct {
	ID            string     `json:"id" db:"id"`
	ProductID     string     `json:"product_id" db:"product_id"`
	SKU           string     `json:"sku" db:"sku"`
	PricePaisa    int64      `json:"price_paisa" db:"price_paisa"`
	Currency      string     `json:"currency" db:"currency"`
	Inventory     int        `json:"inventory" db:"inventory"`
	ReservedCount int        `json:"reserved_count" db:"reserved_count"`
	Size          string     `json:"size,omitempty" db:"size"`
	Color         string     `json:"color,omitempty" db:"color"`
	ValidFrom     *time.Time `json:"valid_from,omitempty" db:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until,omitempty" db:"valid_until"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Review represents a product review (may contain injection payloads for demo).
type Review struct {
	ID        string    `json:"id"`
	Rating    int       `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductSummary is a lightweight product representation for search results.
type ProductSummary struct {
	ID       string `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Category string `json:"category"`
	MinPrice int64  `json:"min_price_paisa"`
	InStock  bool   `json:"in_stock"`
	Image    string `json:"image,omitempty"`
}

// InventoryCheck represents current inventory status for a SKU.
type InventoryCheck struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
	Reserved  int    `json:"reserved"`
	Total     int    `json:"total"`
}

// SearchFilter holds search parameters for product queries.
type SearchFilter struct {
	Query       string
	Category    string
	MaxPrice    *int64
	MinPrice    *int64
	InStockOnly bool
	Limit       int
	Color       string
	Size        string
	Brand       string
}
