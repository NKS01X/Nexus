package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProduct_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		validate func(*testing.T, Product)
	}{
		{
			name: "complete product",
			product: Product{
				ID:          "prod_001",
				SKU:         "SHOE-RUN-001",
				Name:        "Nike Running Shoes",
				Description: "High performance running shoes",
				Category:    "footwear",
				Images:      []string{"img1.jpg", "img2.jpg"},
				Attributes:  json.RawMessage(`{"brand": "Nike", "type": "running"}`),
				Offers: []Offer{
					{ID: "offer_001", SKU: "SHOE-RUN-001-RED-42", PricePaisa: 249900, Currency: "INR", Inventory: 10, Size: "42", Color: "Red"},
				},
				Reviews: []Review{
					{ID: "rev_001", Rating: 5, Title: "Great!", Body: "Amazing quality!", Author: "user1"},
				},
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			validate: func(t *testing.T, p Product) {
				if p.ID != "prod_001" {
					t.Errorf("expected ID=prod_001, got %s", p.ID)
				}
				if len(p.Offers) != 1 {
					t.Errorf("expected 1 offer, got %d", len(p.Offers))
				}
				if len(p.Reviews) != 1 {
					t.Errorf("expected 1 review, got %d", len(p.Reviews))
				}
				if p.Attributes == nil {
					t.Error("expected Attributes to be set")
				}
			},
		},
		{
			name: "product with no offers or reviews",
			product: Product{
				ID:          "prod_002",
				SKU:         "SHIRT-001",
				Name:        "Cotton T-Shirt",
				Description: "Comfortable cotton t-shirt",
				Category:    "apparel",
				Images:      []string{},
				Attributes:  json.RawMessage(`{}`),
				Offers:      []Offer{},
				Reviews:     []Review{},
			},
			validate: func(t *testing.T, p Product) {
				if len(p.Offers) != 0 {
					t.Errorf("expected 0 offers, got %d", len(p.Offers))
				}
				if len(p.Reviews) != 0 {
					t.Errorf("expected 0 reviews, got %d", len(p.Reviews))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.product)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result Product
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestOffer_MarshalUnmarshal(t *testing.T) {
	validFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validUntil := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name     string
		offer    Offer
		validate func(*testing.T, Offer)
	}{
		{
			name: "complete offer",
			offer: Offer{
				ID:            "offer_001",
				ProductID:     "prod_001",
				SKU:           "SHOE-RUN-001-RED-42",
				PricePaisa:    249900,
				Currency:      "INR",
				Inventory:     10,
				ReservedCount: 2,
				Size:          "42",
				Color:         "Red",
				ValidFrom:     &validFrom,
				ValidUntil:    &validUntil,
			},
			validate: func(t *testing.T, o Offer) {
				if o.PricePaisa != 249900 {
					t.Errorf("expected PricePaisa=249900, got %d", o.PricePaisa)
				}
				if o.Inventory != 10 {
					t.Errorf("expected Inventory=10, got %d", o.Inventory)
				}
				if o.ReservedCount != 2 {
					t.Errorf("expected ReservedCount=2, got %d", o.ReservedCount)
				}
				if o.ValidFrom == nil || !o.ValidFrom.Equal(validFrom) {
					t.Error("expected ValidFrom to be set")
				}
				if o.ValidUntil == nil || !o.ValidUntil.Equal(validUntil) {
					t.Error("expected ValidUntil to be set")
				}
			},
		},
		{
			name: "offer without validity dates",
			offer: Offer{
				ID:         "offer_002",
				ProductID:  "prod_001",
				SKU:        "SHOE-RUN-001-BLUE-41",
				PricePaisa: 249900,
				Currency:   "INR",
				Inventory:  5,
			},
			validate: func(t *testing.T, o Offer) {
				if o.ValidFrom != nil {
					t.Error("expected nil ValidFrom")
				}
				if o.ValidUntil != nil {
					t.Error("expected nil ValidUntil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.offer)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result Offer
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestProductSummary_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		summary  ProductSummary
		validate func(*testing.T, ProductSummary)
	}{
		{
			name: "complete summary",
			summary: ProductSummary{
				ID:       "prod_001",
				SKU:      "SHOE-RUN-001-RED-42",
				Name:     "Nike Running Shoes",
				Category: "footwear",
				MinPrice: 249900,
				InStock:  true,
				Image:    "img1.jpg",
			},
			validate: func(t *testing.T, s ProductSummary) {
				if s.InStock != true {
					t.Error("expected InStock=true")
				}
			},
		},
		{
			name: "out of stock summary",
			summary: ProductSummary{
				ID:       "prod_002",
				SKU:      "SHIRT-001-M-L",
				Name:     "Cotton T-Shirt",
				Category: "apparel",
				MinPrice: 99900,
				InStock:  false,
			},
			validate: func(t *testing.T, s ProductSummary) {
				if s.InStock != false {
					t.Error("expected InStock=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.summary)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result ProductSummary
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}

func TestInventoryCheck_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		check    InventoryCheck
		validate func(*testing.T, InventoryCheck)
	}{
		{
			name: "available stock",
			check: InventoryCheck{
				SKU:       "SHOE-RUN-001-RED-42",
				Available: 8,
				Reserved:  2,
				Total:     10,
			},
			validate: func(t *testing.T, c InventoryCheck) {
				if c.Available != 8 {
					t.Errorf("expected Available=8, got %d", c.Available)
				}
				if c.Total != 10 {
					t.Errorf("expected Total=10, got %d", c.Total)
				}
			},
		},
		{
			name: "zero stock",
			check: InventoryCheck{
				SKU:       "SHOE-RUN-001-RED-42",
				Available: 0,
				Reserved:  0,
				Total:     0,
			},
			validate: func(t *testing.T, c InventoryCheck) {
				if c.Available != 0 {
					t.Errorf("expected Available=0, got %d", c.Available)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.check)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var result InventoryCheck
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			tt.validate(t, result)
		})
	}
}
