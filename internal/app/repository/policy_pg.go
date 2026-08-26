package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/razorpay/aegis/internal/app/model"
)

// PolicyPG implements PolicyRepository using PostgreSQL.
type PolicyPG struct {
	db *DB
}

// NewPolicyPG creates a new PostgreSQL policy repository.
func NewPolicyPG(db *DB) *PolicyPG {
	return &PolicyPG{db: db}
}

// GetSpend retrieves the total spend for a buyer/session.
func (r *PolicyPG) GetSpend(ctx context.Context, buyerID, sessionID string) (int64, error) {
	const query = `SELECT spend_paisa FROM policy_state WHERE buyer_id = $1 AND session_id = $2 AND sku = '*'`
	var spend int64
	err := r.db.QueryRowCtx(ctx, query, buyerID, sessionID).Scan(&spend)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get spend: %w", err)
	}
	return spend, nil
}

// AddSpend atomically adds to the spend counter for a buyer/session.
func (r *PolicyPG) AddSpend(ctx context.Context, buyerID, sessionID string, amountPaisa int64) error {
	const query = `
		INSERT INTO policy_state (buyer_id, session_id, sku, spend_paisa, updated_at)
		VALUES ($1, $2, '*', $3, NOW())
		ON CONFLICT (buyer_id, session_id, sku)
		DO UPDATE SET spend_paisa = policy_state.spend_paisa + EXCLUDED.spend_paisa, updated_at = NOW()
	`
	_, err := r.db.ExecCtx(ctx, query, buyerID, sessionID, amountPaisa)
	if err != nil {
		return fmt.Errorf("add spend: %w", err)
	}
	return nil
}

// GetSKUQuantity retrieves the quantity purchased of a specific SKU for a buyer/session.
func (r *PolicyPG) GetSKUQuantity(ctx context.Context, buyerID, sessionID, sku string) (int, error) {
	const query = `SELECT sku_quantity FROM policy_state WHERE buyer_id = $1 AND session_id = $2 AND sku = $3`
	var qty int
	err := r.db.QueryRowCtx(ctx, query, buyerID, sessionID, sku).Scan(&qty)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get sku quantity: %w", err)
	}
	return qty, nil
}

// AddSKUQuantity atomically adds to the SKU quantity counter for a buyer/session.
func (r *PolicyPG) AddSKUQuantity(ctx context.Context, buyerID, sessionID, sku string, quantity int) error {
	const query = `
		INSERT INTO policy_state (buyer_id, session_id, sku, sku_quantity, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (buyer_id, session_id, sku)
		DO UPDATE SET sku_quantity = policy_state.sku_quantity + EXCLUDED.sku_quantity, updated_at = NOW()
	`
	_, err := r.db.ExecCtx(ctx, query, buyerID, sessionID, sku, quantity)
	if err != nil {
		return fmt.Errorf("add sku quantity: %w", err)
	}
	return nil
}

// GetRequestCount retrieves the request count within a time window for a buyer/session.
func (r *PolicyPG) GetRequestCount(ctx context.Context, buyerID, sessionID string, windowSeconds int) (int, error) {
	const query = `
		SELECT COALESCE(SUM(request_count), 0)
		FROM policy_state
		WHERE buyer_id = $1 AND session_id = $2 AND sku = '*' AND window_start >= NOW() - ($3 * interval '1 second')
	`
	var count int
	err := r.db.QueryRowCtx(ctx, query, buyerID, sessionID, windowSeconds).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get request count: %w", err)
	}
	return count, nil
}

// IncrementRequestCount atomically increments the request counter for a buyer/session.
func (r *PolicyPG) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	const query = `
		INSERT INTO policy_state (buyer_id, session_id, sku, request_count, window_start, updated_at)
		VALUES ($1, $2, '*', 1, NOW(), NOW())
		ON CONFLICT (buyer_id, session_id, sku)
		DO UPDATE SET request_count = policy_state.request_count + 1, updated_at = NOW()
	`
	_, err := r.db.ExecCtx(ctx, query, buyerID, sessionID)
	if err != nil {
		return fmt.Errorf("increment request count: %w", err)
	}
	return nil
}

// GetPolicyConfig retrieves the current policy configuration.
func (r *PolicyPG) GetPolicyConfig(ctx context.Context) (*model.PolicyConfig, error) {
	const query = `
		SELECT spend_cap_paisa, per_sku_cap, velocity_max_requests, velocity_window_seconds,
		       allowed_categories, blocked_skus, geo_rules
		FROM policy_config WHERE id = 1
	`
	var cfg model.PolicyConfig
	var perSKUCapJSON, allowedCatsJSON, blockedSKUsJSON, geoRulesJSON []byte

	err := r.db.QueryRowCtx(ctx, query).Scan(
		&cfg.SpendCapPaisa,
		&perSKUCapJSON,
		&cfg.VelocityCap.MaxRequests,
		&cfg.VelocityCap.WindowSeconds,
		&allowedCatsJSON,
		&blockedSKUsJSON,
		&geoRulesJSON,
	)
	if err == sql.ErrNoRows {

		return &model.PolicyConfig{
			SpendCapPaisa:     300000,
			PerSKUCap:         map[string]int{},
			VelocityCap:       model.VelocityLimit{MaxRequests: 10, WindowSeconds: 60},
			AllowedCategories: []string{},
			BlockedSKUs:       []string{},
			GeoRules:          []model.GeoRule{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get policy config: %w", err)
	}

	if len(perSKUCapJSON) > 0 {
		_ = json.Unmarshal(perSKUCapJSON, &cfg.PerSKUCap)
	}
	if len(allowedCatsJSON) > 0 {
		_ = json.Unmarshal(allowedCatsJSON, &cfg.AllowedCategories)
	}
	if len(blockedSKUsJSON) > 0 {
		_ = json.Unmarshal(blockedSKUsJSON, &cfg.BlockedSKUs)
	}
	if len(geoRulesJSON) > 0 {
		_ = json.Unmarshal(geoRulesJSON, &cfg.GeoRules)
	}

	if cfg.PerSKUCap == nil {
		cfg.PerSKUCap = map[string]int{}
	}
	if cfg.AllowedCategories == nil {
		cfg.AllowedCategories = []string{}
	}
	if cfg.BlockedSKUs == nil {
		cfg.BlockedSKUs = []string{}
	}
	if cfg.GeoRules == nil {
		cfg.GeoRules = []model.GeoRule{}
	}

	return &cfg, nil
}

// UpdatePolicyConfig updates the policy configuration.
func (r *PolicyPG) UpdatePolicyConfig(ctx context.Context, config *model.PolicyConfig) error {
	perSKUCapJSON, _ := json.Marshal(config.PerSKUCap)
	allowedCatsJSON, _ := json.Marshal(config.AllowedCategories)
	blockedSKUsJSON, _ := json.Marshal(config.BlockedSKUs)
	geoRulesJSON, _ := json.Marshal(config.GeoRules)

	const query = `
		UPDATE policy_config
		SET spend_cap_paisa = $1, per_sku_cap = $2, velocity_max_requests = $3,
		    velocity_window_seconds = $4, allowed_categories = $5, blocked_skus = $6,
		    geo_rules = $7, updated_at = NOW()
		WHERE id = 1
	`
	_, err := r.db.ExecCtx(ctx, query,
		config.SpendCapPaisa,
		perSKUCapJSON,
		config.VelocityCap.MaxRequests,
		config.VelocityCap.WindowSeconds,
		allowedCatsJSON,
		blockedSKUsJSON,
		geoRulesJSON,
	)
	if err != nil {
		return fmt.Errorf("update policy config: %w", err)
	}
	return nil
}
