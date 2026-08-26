package repository

import (
	"context"
	"testing"

	"github.com/razorpay/aegis/internal/app/model"
)

func TestPolicyPG(t *testing.T) {
	dsn := getTestDSN(t)
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewPolicyPG(db)
	ctx := context.Background()

	_, _ = db.ExecCtx(ctx, `DELETE FROM policy_state`)
	_, _ = db.ExecCtx(ctx, `DELETE FROM policy_config WHERE id = 1`)

	t.Run("GetSpend_default_zero", func(t *testing.T) {
		spend, err := repo.GetSpend(ctx, "buyer_1", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		if spend != 0 {
			t.Errorf("expected spend=0, got %d", spend)
		}
	})

	t.Run("AddSpend_and_GetSpend", func(t *testing.T) {
		err := repo.AddSpend(ctx, "buyer_1", "session_1", 10000)
		if err != nil {
			t.Fatal(err)
		}

		spend, err := repo.GetSpend(ctx, "buyer_1", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		if spend != 10000 {
			t.Errorf("expected spend=10000, got %d", spend)
		}

		err = repo.AddSpend(ctx, "buyer_1", "session_1", 5000)
		if err != nil {
			t.Fatal(err)
		}

		spend, err = repo.GetSpend(ctx, "buyer_1", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		if spend != 15000 {
			t.Errorf("expected spend=15000, got %d", spend)
		}
	})

	t.Run("GetSKUQuantity_default_zero", func(t *testing.T) {
		qty, err := repo.GetSKUQuantity(ctx, "buyer_1", "session_1", "SKU-001")
		if err != nil {
			t.Fatal(err)
		}
		if qty != 0 {
			t.Errorf("expected qty=0, got %d", qty)
		}
	})

	t.Run("AddSKUQuantity_and_GetSKUQuantity", func(t *testing.T) {
		err := repo.AddSKUQuantity(ctx, "buyer_1", "session_1", "SKU-001", 1)
		if err != nil {
			t.Fatal(err)
		}

		qty, err := repo.GetSKUQuantity(ctx, "buyer_1", "session_1", "SKU-001")
		if err != nil {
			t.Fatal(err)
		}
		if qty != 1 {
			t.Errorf("expected qty=1, got %d", qty)
		}

		qty2, err := repo.GetSKUQuantity(ctx, "buyer_1", "session_1", "SKU-002")
		if err != nil {
			t.Fatal(err)
		}
		if qty2 != 0 {
			t.Errorf("expected qty2=0, got %d", qty2)
		}

		qty3, err := repo.GetSKUQuantity(ctx, "buyer_1", "session_2", "SKU-001")
		if err != nil {
			t.Fatal(err)
		}
		if qty3 != 0 {
			t.Errorf("expected qty3=0, got %d", qty3)
		}
	})

	t.Run("GetRequestCount_within_window", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `DELETE FROM policy_state WHERE buyer_id = 'buyer_1'`)

		_, err := db.ExecCtx(ctx, `
			INSERT INTO policy_state (buyer_id, session_id, sku, request_count, window_start, updated_at)
			VALUES ('buyer_1', 'session_1', '*', 5, NOW() - INTERVAL '30 seconds', NOW())
		`)
		if err != nil {
			t.Fatal(err)
		}

		count, err := repo.GetRequestCount(ctx, "buyer_1", "session_1", 60)
		if err != nil {
			t.Fatal(err)
		}
		if count != 5 {
			t.Errorf("expected count=5, got %d", count)
		}
	})

	t.Run("GetRequestCount_outside_window", func(t *testing.T) {
		_, _ = db.ExecCtx(ctx, `DELETE FROM policy_state WHERE buyer_id = 'buyer_1'`)

		_, err := db.ExecCtx(ctx, `
			INSERT INTO policy_state (buyer_id, session_id, sku, request_count, window_start, updated_at)
			VALUES ('buyer_1', 'session_1', '*', 10, NOW() - INTERVAL '5 minutes', NOW())
		`)
		if err != nil {
			t.Fatal(err)
		}

		count, err := repo.GetRequestCount(ctx, "buyer_1", "session_1", 60)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("expected count=0, got %d", count)
		}
	})

	t.Run("IncrementRequestCount", func(t *testing.T) {
		_, _ = db.ExecCtx(ctx, `DELETE FROM policy_state WHERE buyer_id = 'buyer_1'`)

		err := repo.IncrementRequestCount(ctx, "buyer_1", "session_1")
		if err != nil {
			t.Fatal(err)
		}

		err = repo.IncrementRequestCount(ctx, "buyer_1", "session_1")
		if err != nil {
			t.Fatal(err)
		}

		count, err := repo.GetRequestCount(ctx, "buyer_1", "session_1", 60)
		if err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Errorf("expected count=2, got %d", count)
		}
	})

	t.Run("GetPolicyConfig_defaults", func(t *testing.T) {
		cfg, err := repo.GetPolicyConfig(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if cfg == nil {
			t.Fatal("expected config, got nil")
		}
		if cfg.SpendCapPaisa != 300000 {
			t.Errorf("expected SpendCapPaisa=300000, got %d", cfg.SpendCapPaisa)
		}
		if cfg.VelocityCap.MaxRequests != 10 {
			t.Errorf("expected MaxRequests=10, got %d", cfg.VelocityCap.MaxRequests)
		}
		if cfg.VelocityCap.WindowSeconds != 60 {
			t.Errorf("expected WindowSeconds=60, got %d", cfg.VelocityCap.WindowSeconds)
		}
		if cfg.PerSKUCap == nil {
			t.Error("expected PerSKUCap not nil")
		}
		if cfg.AllowedCategories == nil {
			t.Error("expected AllowedCategories not nil")
		}
		if cfg.BlockedSKUs == nil {
			t.Error("expected BlockedSKUs not nil")
		}
		if cfg.GeoRules == nil {
			t.Error("expected GeoRules not nil")
		}
	})

	t.Run("UpdatePolicyConfig_and_Get", func(t *testing.T) {

		_, _ = db.ExecCtx(ctx, `INSERT INTO policy_config (id) VALUES (1) ON CONFLICT DO NOTHING`)

		newCfg := &model.PolicyConfig{
			SpendCapPaisa:     500000,
			PerSKUCap:         map[string]int{"SKU-001": 5},
			VelocityCap:       model.VelocityLimit{MaxRequests: 20, WindowSeconds: 120},
			AllowedCategories: []string{"footwear", "electronics"},
			BlockedSKUs:       []string{"BLOCKED-001"},
			GeoRules:          []model.GeoRule{{Country: "IN", Allowed: true, Pincodes: []string{"560001"}}},
		}

		err := repo.UpdatePolicyConfig(ctx, newCfg)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := repo.GetPolicyConfig(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.SpendCapPaisa != newCfg.SpendCapPaisa {
			t.Errorf("expected SpendCapPaisa=%d, got %d", newCfg.SpendCapPaisa, cfg.SpendCapPaisa)
		}
		if cfg.PerSKUCap["SKU-001"] != 5 {
			t.Errorf("expected PerSKUCap[SKU-001]=5, got %d", cfg.PerSKUCap["SKU-001"])
		}
		if cfg.VelocityCap.MaxRequests != newCfg.VelocityCap.MaxRequests {
			t.Errorf("expected MaxRequests=%d, got %d", newCfg.VelocityCap.MaxRequests, cfg.VelocityCap.MaxRequests)
		}
		if cfg.VelocityCap.WindowSeconds != newCfg.VelocityCap.WindowSeconds {
			t.Errorf("expected WindowSeconds=%d, got %d", newCfg.VelocityCap.WindowSeconds, cfg.VelocityCap.WindowSeconds)
		}
		if len(cfg.AllowedCategories) != 2 {
			t.Errorf("expected 2 allowed categories, got %d", len(cfg.AllowedCategories))
		}
		if len(cfg.BlockedSKUs) != 1 {
			t.Errorf("expected 1 blocked SKU, got %d", len(cfg.BlockedSKUs))
		}
		if len(cfg.GeoRules) != 1 {
			t.Errorf("expected 1 geo rule, got %d", len(cfg.GeoRules))
		}
	})

	t.Run("Concurrent_AddSpend", func(t *testing.T) {
		_, _ = db.ExecCtx(ctx, `DELETE FROM policy_state WHERE buyer_id = 'concurrent_buyer'`)

		done := make(chan error, 100)
		for i := 0; i < 100; i++ {
			go func() {
				err := repo.AddSpend(ctx, "concurrent_buyer", "session_1", 100)
				done <- err
			}()
		}

		for i := 0; i < 100; i++ {
			err := <-done
			if err != nil {
				t.Errorf("concurrent add failed: %v", err)
			}
		}

		spend, err := repo.GetSpend(ctx, "concurrent_buyer", "session_1")
		if err != nil {
			t.Fatal(err)
		}
		if spend != 10000 {
			t.Errorf("expected spend=10000, got %d", spend)
		}
	})
}
