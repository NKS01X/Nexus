package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/razorpay/aegis/internal/app/model"
)

// CatalogPG implements CatalogRepository using PostgreSQL.
type CatalogPG struct {
	db *DB
}

// NewCatalogPG creates a new PostgreSQL catalog repository.
func NewCatalogPG(db *DB) *CatalogPG {
	return &CatalogPG{db: db}
}

// GetProduct retrieves a product by ID with its offers.
func (r *CatalogPG) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	product, err := r.getProductBase(ctx, "id = $1", id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}
	return r.attachOffers(ctx, product)
}

// GetProductBySKU retrieves a product by SKU (product SKU or offer SKU) with its offers.
func (r *CatalogPG) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {

	product, err := r.getProductBase(ctx, "sku = $1", sku)
	if err != nil {
		return nil, err
	}
	if product != nil {
		return r.attachOffers(ctx, product)
	}

	// If not found, try offer-level SKU
	const query = `
		SELECT p.id, p.sku, p.name, p.description, p.category, p.images, p.attributes, p.reviews, p.created_at, p.updated_at
		FROM products p
		JOIN offers o ON p.id = o.product_id
		WHERE o.sku = $1
	`
	var p model.Product
	var imagesJSON, attrsJSON, reviewsJSON []byte
	err = r.db.QueryRowCtx(ctx, query, sku).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category,
		&imagesJSON, &attrsJSON, &reviewsJSON, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get product by offer sku: %w", err)
	}

	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &p.Images); err != nil {
			return nil, fmt.Errorf("unmarshal images: %w", err)
		}
	}
	if len(attrsJSON) > 0 {
		if err := json.Unmarshal(attrsJSON, &p.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}
	if len(reviewsJSON) > 0 {
		if err := json.Unmarshal(reviewsJSON, &p.Reviews); err != nil {
			return nil, fmt.Errorf("unmarshal reviews: %w", err)
		}
	}

	return r.attachOffers(ctx, &p)
}

// SearchProducts searches products with filters.
func (r *CatalogPG) SearchProducts(ctx context.Context, filter SearchFilter) ([]*model.Product, error) {
	query := `SELECT id, sku, name, description, category, images, attributes, reviews, created_at, updated_at FROM products WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Query != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE price_paisa <= $%d)", argIdx)
		args = append(args, *filter.MaxPrice)
		argIdx++
	}
	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE price_paisa >= $%d)", argIdx)
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.InStockOnly {
		query += ` AND id IN (SELECT product_id FROM offers WHERE inventory - reserved_count > 0)`
	}
	if filter.Color != "" {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE color = $%d)", argIdx)
		args = append(args, filter.Color)
		argIdx++
	}
	if filter.Size != "" {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE size = $%d)", argIdx)
		args = append(args, filter.Size)
		argIdx++
	}
	if filter.Brand != "" {
		query += fmt.Sprintf(" AND attributes->>'brand' = $%d", argIdx)
		args = append(args, filter.Brand)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
	}

	rows, err := r.db.QueryCtx(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}

	for _, p := range products {
		if _, err := r.attachOffers(ctx, p); err != nil {
			return nil, err
		}
	}

	return products, nil
}

// GetAllProducts retrieves all products with offers.
func (r *CatalogPG) GetAllProducts(ctx context.Context) ([]*model.Product, error) {
	return r.SearchProducts(ctx, SearchFilter{})
}

// CheckAvailability checks inventory for a SKU.
func (r *CatalogPG) CheckAvailability(ctx context.Context, sku string) (*model.InventoryCheck, error) {
	const query = `
		SELECT sku, inventory, reserved_count, inventory - reserved_count AS available
		FROM offers WHERE sku = $1
	`
	var check model.InventoryCheck
	err := r.db.QueryRowCtx(ctx, query, sku).Scan(&check.SKU, &check.Total, &check.Reserved, &check.Available)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}
	return &check, nil
}

// ReserveInventory atomically reserves inventory using optimistic locking.
func (r *CatalogPG) ReserveInventory(ctx context.Context, sku string, quantity int) error {
	const query = `
		UPDATE offers
		SET reserved_count = reserved_count + $1, updated_at = NOW()
		WHERE sku = $2 AND inventory - reserved_count >= $1
	`
	result, err := r.db.ExecCtx(ctx, query, quantity, sku)
	if err != nil {
		return fmt.Errorf("reserve inventory: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrInsufficientInventory
	}
	return nil
}

// ReleaseInventory releases reserved inventory.
func (r *CatalogPG) ReleaseInventory(ctx context.Context, sku string, quantity int) error {
	const query = `
		UPDATE offers
		SET reserved_count = reserved_count - $1, updated_at = NOW()
		WHERE sku = $2 AND reserved_count >= $1
	`
	result, err := r.db.ExecCtx(ctx, query, quantity, sku)
	if err != nil {
		return fmt.Errorf("release inventory: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrInventoryNotReserved
	}
	return nil
}

// ConfirmInventory confirms the reservation (moves from reserved to sold).
func (r *CatalogPG) ConfirmInventory(ctx context.Context, sku string, quantity int) error {
	const query = `
		UPDATE offers
		SET inventory = inventory - $1, reserved_count = reserved_count - $1, updated_at = NOW()
		WHERE sku = $2 AND reserved_count >= $1 AND inventory >= $1
	`
	result, err := r.db.ExecCtx(ctx, query, quantity, sku)
	if err != nil {
		return fmt.Errorf("confirm inventory: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return model.ErrInsufficientInventory
	}
	return nil
}

// getProductBase retrieves a product by a WHERE clause.
func (r *CatalogPG) getProductBase(ctx context.Context, whereClause string, args ...any) (*model.Product, error) {
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category, images, attributes, reviews, created_at, updated_at
		FROM products WHERE %s
	`, whereClause)
	return r.scanProduct(r.db.QueryRowCtx(ctx, query, args...))
}

// scanProduct scans a product from a Row or Rows.
func (r *CatalogPG) scanProduct(scanner interface{ Scan(dest ...any) error }) (*model.Product, error) {
	var p model.Product
	var imagesJSON, attrsJSON, reviewsJSON []byte
	err := scanner.Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Category,
		&imagesJSON, &attrsJSON, &reviewsJSON, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan product: %w", err)
	}
	if len(imagesJSON) > 0 {
		_ = json.Unmarshal(imagesJSON, &p.Images)
	}
	if len(attrsJSON) > 0 {
		p.Attributes = json.RawMessage(attrsJSON)
	}
	if len(reviewsJSON) > 0 {
		_ = json.Unmarshal(reviewsJSON, &p.Reviews)
	}
	return &p, nil
}

// attachOffers loads and attaches offers to a product.
func (r *CatalogPG) attachOffers(ctx context.Context, p *model.Product) (*model.Product, error) {
	const query = `
		SELECT id, product_id, sku, price_paisa, currency, inventory, reserved_count,
		       size, color, valid_from, valid_until, created_at, updated_at
		FROM offers WHERE product_id = $1
	`
	rows, err := r.db.QueryCtx(ctx, query, p.ID)
	if err != nil {
		return nil, fmt.Errorf("query offers: %w", err)
	}
	defer rows.Close()

	var offers []model.Offer
	for rows.Next() {
		var o model.Offer
		var validFrom, validUntil sql.NullTime
		if err := rows.Scan(
			&o.ID, &o.ProductID, &o.SKU, &o.PricePaisa, &o.Currency,
			&o.Inventory, &o.ReservedCount, &o.Size, &o.Color,
			&validFrom, &validUntil, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		if validFrom.Valid {
			o.ValidFrom = &validFrom.Time
		}
		if validUntil.Valid {
			o.ValidUntil = &validUntil.Time
		}
		offers = append(offers, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}

	p.Offers = offers
	return p, nil
}

// InsertProduct inserts a product with tenant scoping.
func (r *CatalogPG) InsertProduct(ctx context.Context, p *model.Product, tenantID string) error {
	imagesJSON, _ := json.Marshal(p.Images)
	attrsJSON, _ := json.Marshal(p.Attributes)
	reviewsJSON, _ := json.Marshal(p.Reviews)

	const productQuery = `
		INSERT INTO products (id, sku, name, description, category, images, attributes, reviews, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			sku = EXCLUDED.sku,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			category = EXCLUDED.category,
			images = EXCLUDED.images,
			attributes = EXCLUDED.attributes,
			reviews = EXCLUDED.reviews,
			tenant_id = EXCLUDED.tenant_id,
			updated_at = NOW()
	`
	_, err := r.db.ExecCtx(ctx, productQuery, p.ID, p.SKU, p.Name, p.Description, p.Category,
		imagesJSON, attrsJSON, reviewsJSON, tenantID)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	// Insert offers for this product
	for _, offer := range p.Offers {
		if err := r.insertOffer(ctx, p.ID, &offer, tenantID); err != nil {
			return fmt.Errorf("insert offer: %w", err)
		}
	}
	return nil
}

// insertOffer is a helper to insert an offer with tenant scoping.
func (r *CatalogPG) insertOffer(ctx context.Context, productID string, o *model.Offer, tenantID string) error {
	const query = `
		INSERT INTO offers (id, product_id, sku, price_paisa, currency, inventory, reserved_count,
		                   size, color, valid_from, valid_until, tenant_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			sku = EXCLUDED.sku,
			price_paisa = EXCLUDED.price_paisa,
			currency = EXCLUDED.currency,
			inventory = EXCLUDED.inventory,
			reserved_count = EXCLUDED.reserved_count,
			size = EXCLUDED.size,
			color = EXCLUDED.color,
			valid_from = EXCLUDED.valid_from,
			valid_until = EXCLUDED.valid_until,
			tenant_id = EXCLUDED.tenant_id,
			updated_at = NOW()
	`
	_, err := r.db.ExecCtx(ctx, query, o.ID, productID, o.SKU, o.PricePaisa, o.Currency,
		o.Inventory, o.ReservedCount, o.Size, o.Color, o.ValidFrom, o.ValidUntil, tenantID)
	if err != nil {
		return fmt.Errorf("insert offer: %w", err)
	}
	return nil
}

// SearchProductsByTenant searches products scoped to a tenant.
func (r *CatalogPG) SearchProductsByTenant(ctx context.Context, filter SearchFilter, tenantID string) ([]*model.Product, error) {
	query := `SELECT id, sku, name, description, category, images, attributes, reviews, created_at, updated_at FROM products WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

	if filter.Query != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.MaxPrice != nil {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE price_paisa <= $%d AND tenant_id = $%d)", argIdx, argIdx+1)
		args = append(args, *filter.MaxPrice, tenantID)
		argIdx += 2
	}
	if filter.MinPrice != nil {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE price_paisa >= $%d AND tenant_id = $%d)", argIdx, argIdx+1)
		args = append(args, *filter.MinPrice, tenantID)
		argIdx += 2
	}
	if filter.InStockOnly {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE inventory - reserved_count > 0 AND tenant_id = $%d)", argIdx)
		args = append(args, tenantID)
		argIdx++
	}
	if filter.Color != "" {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE color = $%d AND tenant_id = $%d)", argIdx, argIdx+1)
		args = append(args, filter.Color, tenantID)
		argIdx += 2
	}
	if filter.Size != "" {
		query += fmt.Sprintf(" AND id IN (SELECT product_id FROM offers WHERE size = $%d AND tenant_id = $%d)", argIdx, argIdx+1)
		args = append(args, filter.Size, tenantID)
		argIdx += 2
	}
	if filter.Brand != "" {
		query += fmt.Sprintf(" AND attributes->>'brand' = $%d", argIdx)
		args = append(args, filter.Brand)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
	}

	rows, err := r.db.QueryCtx(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p, err := r.scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}

	for _, p := range products {
		if _, err := r.attachOffersByTenant(ctx, p, tenantID); err != nil {
			return nil, err
		}
	}

	return products, nil
}

// GetProductByTenant retrieves a product by ID scoped to a tenant.
func (r *CatalogPG) GetProductByTenant(ctx context.Context, id string, tenantID string) (*model.Product, error) {
	product, err := r.getProductBaseByTenant(ctx, "id = $1 AND tenant_id = $2", id, tenantID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, nil
	}
	return r.attachOffersByTenant(ctx, product, tenantID)
}

// CheckAvailabilityByTenant checks inventory for a SKU scoped to a tenant.
func (r *CatalogPG) CheckAvailabilityByTenant(ctx context.Context, sku string, tenantID string) (*model.InventoryCheck, error) {
	const query = `
		SELECT sku, inventory, reserved_count, inventory - reserved_count AS available
		FROM offers WHERE sku = $1 AND tenant_id = $2
	`
	var check model.InventoryCheck
	err := r.db.QueryRowCtx(ctx, query, sku, tenantID).Scan(&check.SKU, &check.Total, &check.Reserved, &check.Available)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}
	return &check, nil
}

// getProductBaseByTenant retrieves a product by a WHERE clause with tenant scoping.
func (r *CatalogPG) getProductBaseByTenant(ctx context.Context, whereClause string, args ...any) (*model.Product, error) {
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category, images, attributes, reviews, created_at, updated_at
		FROM products WHERE %s
	`, whereClause)
	return r.scanProduct(r.db.QueryRowCtx(ctx, query, args...))
}

// attachOffersByTenant loads and attaches offers to a product scoped to a tenant.
func (r *CatalogPG) attachOffersByTenant(ctx context.Context, p *model.Product, tenantID string) (*model.Product, error) {
	const query = `
		SELECT id, product_id, sku, price_paisa, currency, inventory, reserved_count,
		       size, color, valid_from, valid_until, created_at, updated_at
		FROM offers WHERE product_id = $1 AND tenant_id = $2
	`
	rows, err := r.db.QueryCtx(ctx, query, p.ID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("query offers: %w", err)
	}
	defer rows.Close()

	var offers []model.Offer
	for rows.Next() {
		var o model.Offer
		var validFrom, validUntil sql.NullTime
		if err := rows.Scan(
			&o.ID, &o.ProductID, &o.SKU, &o.PricePaisa, &o.Currency,
			&o.Inventory, &o.ReservedCount, &o.Size, &o.Color,
			&validFrom, &validUntil, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		if validFrom.Valid {
			o.ValidFrom = &validFrom.Time
		}
		if validUntil.Valid {
			o.ValidUntil = &validUntil.Time
		}
		offers = append(offers, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iterate: %w", err)
	}

	p.Offers = offers
	return p, nil
}

// BuildDynamicWhere builds a WHERE clause for search filters (exported for tests).
func BuildDynamicWhere(filter SearchFilter) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	argIdx := 1

	if filter.Query != "" {
		clauses = append(clauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Query+"%")
		argIdx++
	}
	if filter.Category != "" {
		clauses = append(clauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.MaxPrice != nil {
		clauses = append(clauses, fmt.Sprintf("id IN (SELECT product_id FROM offers WHERE price_paisa <= $%d)", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}
	if filter.MinPrice != nil {
		clauses = append(clauses, fmt.Sprintf("id IN (SELECT product_id FROM offers WHERE price_paisa >= $%d)", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.InStockOnly {
		clauses = append(clauses, `id IN (SELECT product_id FROM offers WHERE inventory - reserved_count > 0)`)
	}
	if filter.Color != "" {
		clauses = append(clauses, fmt.Sprintf("id IN (SELECT product_id FROM offers WHERE color = $%d)", argIdx))
		args = append(args, filter.Color)
		argIdx++
	}
	if filter.Size != "" {
		clauses = append(clauses, fmt.Sprintf("id IN (SELECT product_id FROM offers WHERE size = $%d)", argIdx))
		args = append(args, filter.Size)
		argIdx++
	}
	if filter.Brand != "" {
		clauses = append(clauses, fmt.Sprintf("attributes->>'brand' = $%d", argIdx))
		args = append(args, filter.Brand)
		argIdx++
	}

	return strings.Join(clauses, " AND "), args
}
