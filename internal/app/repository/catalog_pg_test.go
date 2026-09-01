package repository

import (
	"context"
	"testing"

	"github.com/razorpay/aegis/internal/app/model"
)

// TestCatalogPG tests require a real PostgreSQL database.
// Set TEST_DATABASE_URL environment variable to run them.
func TestCatalogPG(t *testing.T) {
	dsn := getTestDSN(t)
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewCatalogPG(db)
	ctx := context.Background()

	setupTestCatalog(t, ctx, db)

	t.Run("GetProduct", func(t *testing.T) {
		product, err := repo.GetProduct(ctx, "prod_001")
		if err != nil {
			t.Fatal(err)
		}
		if product == nil {
			t.Fatal("expected product, got nil")
		}
		if product.ID != "prod_001" {
			t.Errorf("expected ID=prod_001, got %s", product.ID)
		}
		if product.SKU != "SHOE-RUN-001" {
			t.Errorf("expected SKU=SHOE-RUN-001, got %s", product.SKU)
		}
		if len(product.Offers) != 2 {
			t.Errorf("expected 2 offers, got %d", len(product.Offers))
		}
	})

	t.Run("GetProduct_not_found", func(t *testing.T) {
		product, err := repo.GetProduct(ctx, "nonexistent")
		if err != nil {
			t.Fatal(err)
		}
		if product != nil {
			t.Errorf("expected nil product, got %v", product)
		}
	})

	t.Run("GetProductBySKU", func(t *testing.T) {
		product, err := repo.GetProductBySKU(ctx, "SHOE-RUN-001")
		if err != nil {
			t.Fatal(err)
		}
		if product == nil {
			t.Fatal("expected product, got nil")
		}
		if product.SKU != "SHOE-RUN-001" {
			t.Errorf("expected SKU=SHOE-RUN-001, got %s", product.SKU)
		}
	})

	t.Run("SearchProducts_by_query", func(t *testing.T) {
		products, err := repo.SearchProducts(ctx, SearchFilter{Query: "running", Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		if len(products) < 1 {
			t.Errorf("expected at least 1 product, got %d", len(products))
		}
	})

	t.Run("SearchProducts_by_category", func(t *testing.T) {
		products, err := repo.SearchProducts(ctx, SearchFilter{Category: "footwear", Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		if len(products) < 1 {
			t.Errorf("expected at least 1 product, got %d", len(products))
		}
		for _, p := range products {
			if p.Category != "footwear" {
				t.Errorf("expected category=footwear, got %s", p.Category)
			}
		}
	})

	t.Run("SearchProducts_by_max_price", func(t *testing.T) {
		maxPrice := int64(250000)
		products, err := repo.SearchProducts(ctx, SearchFilter{MaxPrice: &maxPrice, Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range products {
			for _, o := range p.Offers {
				if o.PricePaisa > maxPrice {
					t.Errorf("offer price %d exceeds max %d", o.PricePaisa, maxPrice)
				}
			}
		}
	})

	t.Run("SearchProducts_in_stock_only", func(t *testing.T) {
		products, err := repo.SearchProducts(ctx, SearchFilter{InStockOnly: true, Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range products {
			inStock := false
			for _, o := range p.Offers {
				if o.Inventory-o.ReservedCount > 0 {
					inStock = true
					break
				}
			}
			if !inStock {
				t.Errorf("product %s should have at least one offer in stock", p.ID)
			}
		}
	})

	t.Run("CheckAvailability", func(t *testing.T) {
		check, err := repo.CheckAvailability(ctx, "SHOE-RUN-001-RED-42")
		if err != nil {
			t.Fatal(err)
		}
		if check == nil {
			t.Fatal("expected check, got nil")
		}
		if check.SKU != "SHOE-RUN-001-RED-42" {
			t.Errorf("expected SKU=SHOE-RUN-001-RED-42, got %s", check.SKU)
		}
		if check.Available <= 0 {
			t.Errorf("expected available > 0, got %d", check.Available)
		}
	})

	t.Run("CheckAvailability_not_found", func(t *testing.T) {
		check, err := repo.CheckAvailability(ctx, "NONEXISTENT")
		if err != nil {
			t.Fatal(err)
		}
		if check != nil {
			t.Errorf("expected nil check, got %v", check)
		}
	})

	t.Run("ReserveInventory_success", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `UPDATE offers SET reserved_count = 0 WHERE sku = 'SHOE-RUN-001-RED-42'`)

		err := repo.ReserveInventory(ctx, "SHOE-RUN-001-RED-42", 1)
		if err != nil {
			t.Fatal(err)
		}

		check, _ := repo.CheckAvailability(ctx, "SHOE-RUN-001-RED-42")
		if check.Available != 9 {
			t.Errorf("expected available=9, got %d", check.Available)
		}
		if check.Reserved != 1 {
			t.Errorf("expected reserved=1, got %d", check.Reserved)
		}
	})

	t.Run("ReserveInventory_insufficient", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `UPDATE offers SET reserved_count = 0 WHERE sku = 'SHOE-RUN-001-RED-42'`)

		err := repo.ReserveInventory(ctx, "SHOE-RUN-001-RED-42", 10)
		if err != nil {
			t.Fatal(err)
		}

		err = repo.ReserveInventory(ctx, "SHOE-RUN-001-RED-42", 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != model.ErrInsufficientInventory {
			t.Errorf("expected ErrInsufficientInventory, got %v", err)
		}
	})

	t.Run("ReleaseInventory", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `UPDATE offers SET reserved_count = 0 WHERE sku = 'SHOE-RUN-001-RED-42'`)
		err := repo.ReserveInventory(ctx, "SHOE-RUN-001-RED-42", 2)
		if err != nil {
			t.Fatal(err)
		}

		err = repo.ReleaseInventory(ctx, "SHOE-RUN-001-RED-42", 1)
		if err != nil {
			t.Fatal(err)
		}

		check, _ := repo.CheckAvailability(ctx, "SHOE-RUN-001-RED-42")
		if check.Available != 9 {
			t.Errorf("expected available=9, got %d", check.Available)
		}
		if check.Reserved != 1 {
			t.Errorf("expected reserved=1, got %d", check.Reserved)
		}
	})

	t.Run("ReleaseInventory_not_reserved", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `UPDATE offers SET reserved_count = 0 WHERE sku = 'SHOE-RUN-001-RED-42'`)

		err := repo.ReleaseInventory(ctx, "SHOE-RUN-001-RED-42", 5)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != model.ErrInventoryNotReserved {
			t.Errorf("expected ErrInventoryNotReserved, got %v", err)
		}
	})

	t.Run("ConfirmInventory", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `UPDATE offers SET reserved_count = 0 WHERE sku = 'SHOE-RUN-001-RED-42'`)
		err := repo.ReserveInventory(ctx, "SHOE-RUN-001-RED-42", 2)
		if err != nil {
			t.Fatal(err)
		}

		err = repo.ConfirmInventory(ctx, "SHOE-RUN-001-RED-42", 2)
		if err != nil {
			t.Fatal(err)
		}

		check, _ := repo.CheckAvailability(ctx, "SHOE-RUN-001-RED-42")
		if check.Total != 8 {
			t.Errorf("expected total=8, got %d", check.Total)
		}
		if check.Reserved != 0 {
			t.Errorf("expected reserved=0, got %d", check.Reserved)
		}
		if check.Available != 8 {
			t.Errorf("expected available=8, got %d", check.Available)
		}
	})

	t.Run("GetAllProducts", func(t *testing.T) {
		products, err := repo.GetAllProducts(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(products) < 1 {
			t.Errorf("expected at least 1 product, got %d", len(products))
		}
	})
}

func setupTestCatalog(t *testing.T, ctx context.Context, db *DB) {
	t.Helper()

	_, _ = db.ExecCtx(ctx, `DELETE FROM offers`)
	_, _ = db.ExecCtx(ctx, `DELETE FROM products`)

	_, err := db.ExecCtx(ctx, `
		INSERT INTO products (id, sku, name, description, category, images, attributes, reviews, tenant_id, created_at, updated_at)
		VALUES ('prod_001', 'SHOE-RUN-001', 'Running Shoes', 'Comfortable running shoes', 'footwear',
		        '["img1.jpg", "img2.jpg"]', '{"brand": "Nike"}', '[{"rating": 5, "title": "Great!"}]',
		        'default', NOW(), NOW())
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecCtx(ctx, `
		INSERT INTO offers (id, product_id, sku, price_paisa, currency, inventory, reserved_count, size, color, tenant_id, created_at, updated_at)
		VALUES ('offer_001', 'prod_001', 'SHOE-RUN-001-RED-42', 249900, 'INR', 10, 0, '42', 'red', 'default', NOW(), NOW()),
		       ('offer_002', 'prod_001', 'SHOE-RUN-001-BLU-43', 249900, 'INR', 5, 0, '43', 'blue', 'default', NOW(), NOW())
	`)
	if err != nil {
		t.Fatal(err)
	}
}
