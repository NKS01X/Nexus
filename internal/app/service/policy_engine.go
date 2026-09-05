package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

var (
	ErrPolicyConfigNotFound = errors.New("policy config not found")
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidRequest       = errors.New("invalid purchase request")
)

// PolicyEngineImpl implements the PolicyEngine interface with deterministic policy evaluation.
type PolicyEngineImpl struct {
	policyRepo  repository.PolicyRepository
	catalogRepo repository.CatalogRepository
}

// NewPolicyEngine creates a new PolicyEngineImpl.
func NewPolicyEngine(policyRepo repository.PolicyRepository, catalogRepo repository.CatalogRepository) *PolicyEngineImpl {
	return &PolicyEngineImpl{
		policyRepo:  policyRepo,
		catalogRepo: catalogRepo,
	}
}

// Evaluate performs deterministic policy evaluation against all configured caps.
// Order of evaluation: Spend Cap → Per-SKU Cap → Velocity Cap → Category Allowlist → Blocked SKUs → Geo Rules.
func (e *PolicyEngineImpl) Evaluate(ctx context.Context, req *model.PurchaseRequest) (*model.PolicyDecision, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrInvalidRequest)
	}
	if req.BuyerID == "" || req.SessionID == "" {
		return nil, fmt.Errorf("%w: buyer_id and session_id are required", ErrInvalidRequest)
	}
	if req.SKU == "" {
		return nil, fmt.Errorf("%w: sku is required", ErrInvalidRequest)
	}
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be positive", ErrInvalidRequest)
	}
	if req.AmountPaisa <= 0 {
		return nil, fmt.Errorf("%w: amount_paisa must be positive", ErrInvalidRequest)
	}

	cfg, err := e.policyRepo.GetPolicyConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get policy config: %w", err)
	}
	if cfg == nil {
		return nil, ErrPolicyConfigNotFound
	}

	product, err := e.catalogRepo.GetProductBySKU(ctx, req.SKU)
	if err != nil {
		return nil, fmt.Errorf("get product by sku: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("%w: sku=%s", ErrProductNotFound, req.SKU)
	}

	currentSpend, err := e.policyRepo.GetSpend(ctx, req.BuyerID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("get spend: %w", err)
	}

	currentVelocity, err := e.policyRepo.GetRequestCount(ctx, req.BuyerID, req.SessionID, cfg.VelocityCap.WindowSeconds)
	if err != nil {
		return nil, fmt.Errorf("get request count: %w", err)
	}

	if cfg.SpendCapPaisa > 0 && currentSpend+req.AmountPaisa > cfg.SpendCapPaisa {
		remaining := cfg.SpendCapPaisa - currentSpend
		if remaining < 0 {
			remaining = 0
		}
		return &model.PolicyDecision{
			Allowed:   false,
			Reason:    fmt.Sprintf("exceeds session spend cap of %d paisa (current: %d, requested: %d)", cfg.SpendCapPaisa, currentSpend, req.AmountPaisa),
			RuleFired: model.RuleFiredSpendCap,
			Details:   json.RawMessage(fmt.Sprintf(`{"current_spend": %d, "requested": %d, "cap": %d}`, currentSpend, req.AmountPaisa, cfg.SpendCapPaisa)),
			Remaining: model.PolicyRemaining{
				SpendCapPaisa:     remaining,
				PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
				VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
			},
		}, nil
	}

	// Check per-SKU cap against both offer SKU and product SKU
	skuToCheck := []string{req.SKU}
	if product.SKU != req.SKU {
		skuToCheck = append(skuToCheck, product.SKU)
	}
	for _, sku := range skuToCheck {
		if cap, exists := cfg.PerSKUCap[sku]; exists && cap > 0 {
			// Get current quantity for this specific SKU
			currentSKUQty, err := e.policyRepo.GetSKUQuantity(ctx, req.BuyerID, req.SessionID, sku)
			if err != nil {
				return nil, fmt.Errorf("get sku quantity: %w", err)
			}
			if currentSKUQty+req.Quantity > cap {
				remaining := cap - currentSKUQty
				if remaining < 0 {
					remaining = 0
				}
				return &model.PolicyDecision{
					Allowed:   false,
					Reason:    fmt.Sprintf("exceeds per-SKU cap of %d for SKU %s (current: %d, requested: %d)", cap, sku, currentSKUQty, req.Quantity),
					RuleFired: model.RuleFiredPerSKUCap,
					Details:   json.RawMessage(fmt.Sprintf(`{"sku": "%s", "current_qty": %d, "requested_qty": %d, "cap": %d}`, sku, currentSKUQty, req.Quantity, cap)),
					Remaining: model.PolicyRemaining{
						SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend,
						PerSKUCap:         map[string]int{sku: remaining},
						VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
					},
				}, nil
			}
		}
	}

	if cfg.VelocityCap.MaxRequests > 0 && currentVelocity >= cfg.VelocityCap.MaxRequests {
		return &model.PolicyDecision{
			Allowed:   false,
			Reason:    fmt.Sprintf("exceeds velocity cap of %d requests per %d seconds (current: %d)", cfg.VelocityCap.MaxRequests, cfg.VelocityCap.WindowSeconds, currentVelocity),
			RuleFired: model.RuleFiredVelocityCap,
			Details:   json.RawMessage(fmt.Sprintf(`{"current_count": %d, "cap": %d, "window_seconds": %d}`, currentVelocity, cfg.VelocityCap.MaxRequests, cfg.VelocityCap.WindowSeconds)),
			Remaining: model.PolicyRemaining{
				SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend,
				PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
				VelocityRemaining: 0,
			},
		}, nil
	}

	if len(cfg.AllowedCategories) > 0 {
		allowed := false
		for _, cat := range cfg.AllowedCategories {
			if cat == product.Category {
				allowed = true
				break
			}
		}
		if !allowed {
			details, _ := json.Marshal(map[string]any{
				"product_category":   product.Category,
				"allowed_categories": cfg.AllowedCategories,
			})
			return &model.PolicyDecision{
				Allowed:   false,
				Reason:    fmt.Sprintf("product category '%s' not in allowed categories: %v", product.Category, cfg.AllowedCategories),
				RuleFired: model.RuleFiredCategoryBlocked,
				Details:   details,
				Remaining: model.PolicyRemaining{
					SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend,
					PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
					VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
				},
			}, nil
		}
	}

	for _, blockedSKU := range cfg.BlockedSKUs {
		if blockedSKU == req.SKU {
			return &model.PolicyDecision{
				Allowed:   false,
				Reason:    fmt.Sprintf("SKU %s is blocked", req.SKU),
				RuleFired: model.RuleFiredSKUBlocked,
				Details:   json.RawMessage(fmt.Sprintf(`{"blocked_sku": "%s"}`, req.SKU)),
				Remaining: model.PolicyRemaining{
					SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend,
					PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
					VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
				},
			}, nil
		}
	}

	if len(cfg.GeoRules) > 0 && req.BuyerPincode != "" {
		allowed := false
		for _, rule := range cfg.GeoRules {
			if rule.Allowed {

				if len(rule.Pincodes) == 0 {
					allowed = true
					break
				}
				for _, pc := range rule.Pincodes {
					if pc == req.BuyerPincode {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			}
		}
		if !allowed {
			details, _ := json.Marshal(map[string]any{
				"buyer_pincode": req.BuyerPincode,
				"geo_rules":     cfg.GeoRules,
			})
			return &model.PolicyDecision{
				Allowed:   false,
				Reason:    fmt.Sprintf("buyer pincode %s not allowed by geo rules", req.BuyerPincode),
				RuleFired: model.RuleFiredGeoRestricted,
				Details:   details,
				Remaining: model.PolicyRemaining{
					SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend,
					PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
					VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
				},
			}, nil
		}
	}

	return &model.PolicyDecision{
		Allowed:   true,
		Reason:    "within all policy limits",
		RuleFired: model.RuleFiredNone,
		Remaining: model.PolicyRemaining{
			SpendCapPaisa:     cfg.SpendCapPaisa - currentSpend - req.AmountPaisa,
			PerSKUCap:         e.computeRemainingPerSKUCap(ctx, cfg, req.BuyerID, req.SessionID),
			VelocityRemaining: e.computeRemainingVelocity(cfg, currentVelocity),
		},
	}, nil
}

// computeRemainingPerSKUCap calculates remaining quantity for SKUs with configured caps.
func (e *PolicyEngineImpl) computeRemainingPerSKUCap(ctx context.Context, cfg *model.PolicyConfig, buyerID, sessionID string) map[string]int {
	result := make(map[string]int)
	for sku, cap := range cfg.PerSKUCap {
		if cap > 0 {
			current, _ := e.policyRepo.GetSKUQuantity(ctx, buyerID, sessionID, sku)
			remaining := cap - current
			if remaining < 0 {
				remaining = 0
			}
			result[sku] = remaining
		}
	}
	return result
}

// computeRemainingVelocity calculates remaining requests in current velocity window.
func (e *PolicyEngineImpl) computeRemainingVelocity(cfg *model.PolicyConfig, currentCount int) int {
	if cfg.VelocityCap.MaxRequests <= 0 {
		return -1
	}
	remaining := cfg.VelocityCap.MaxRequests - currentCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RecordSpend records a successful purchase by updating spend and SKU quantity.
// Note: Request count is already incremented in gateway service before policy evaluation.
func (e *PolicyEngineImpl) RecordSpend(ctx context.Context, buyerID, sessionID, sku string, quantity int, amountPaisa int64) error {
	if err := e.policyRepo.AddSpend(ctx, buyerID, sessionID, amountPaisa); err != nil {
		return fmt.Errorf("add spend: %w", err)
	}
	if err := e.policyRepo.AddSKUQuantity(ctx, buyerID, sessionID, sku, quantity); err != nil {
		return fmt.Errorf("add sku quantity: %w", err)
	}
	return nil
}

// RollbackSpend rolls back a purchase by decrementing spend, SKU quantity, and velocity counters.
// Note: This is best-effort; velocity counter is not decremented to prevent abuse.
func (e *PolicyEngineImpl) RollbackSpend(ctx context.Context, buyerID, sessionID, sku string, quantity int, amountPaisa int64) error {

	if amountPaisa > 0 {

		if err := e.policyRepo.AddSpend(ctx, buyerID, sessionID, -amountPaisa); err != nil {
			return fmt.Errorf("rollback spend: %w", err)
		}
	}
	if quantity > 0 {
		if err := e.policyRepo.AddSKUQuantity(ctx, buyerID, sessionID, sku, -quantity); err != nil {
			return fmt.Errorf("rollback sku quantity: %w", err)
		}
	}
	return nil
}

// IncrementRequestCount increments the request counter for velocity tracking.
func (e *PolicyEngineImpl) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	return e.policyRepo.IncrementRequestCount(ctx, buyerID, sessionID)
}
