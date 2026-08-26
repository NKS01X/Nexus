package repository

import (
	"context"
	"testing"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
)

func TestOrderPG(t *testing.T) {
	dsn := getTestDSN(t)
	db, err := NewDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewOrderPG(db)
	ctx := context.Background()

	_, _ = db.ExecCtx(ctx, `DELETE FROM orders`)

	t.Run("CreateOrder_and_GetOrder", func(t *testing.T) {
		order := &model.Order{
			ID:             "order_001",
			BuyerID:        "buyer_1",
			SessionID:      "session_1",
			ProductID:      "prod_001",
			SKU:            "SHOE-001",
			Quantity:       1,
			AmountPaisa:    249900,
			Currency:       "INR",
			Status:         model.OrderStatusPending,
			IdempotencyKey: "idem_001",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := repo.CreateOrder(ctx, order)
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := repo.GetOrder(ctx, "order_001")
		if err != nil {
			t.Fatal(err)
		}
		if retrieved == nil {
			t.Fatal("expected order, got nil")
		}
		if retrieved.ID != order.ID {
			t.Errorf("expected ID=%s, got %s", order.ID, retrieved.ID)
		}
		if retrieved.BuyerID != order.BuyerID {
			t.Errorf("expected BuyerID=%s, got %s", order.BuyerID, retrieved.BuyerID)
		}
		if retrieved.SKU != order.SKU {
			t.Errorf("expected SKU=%s, got %s", order.SKU, retrieved.SKU)
		}
		if retrieved.AmountPaisa != order.AmountPaisa {
			t.Errorf("expected AmountPaisa=%d, got %d", order.AmountPaisa, retrieved.AmountPaisa)
		}
		if retrieved.Status != model.OrderStatusPending {
			t.Errorf("expected status=PENDING, got %s", retrieved.Status)
		}
	})

	t.Run("CreateOrder_idempotent_replay", func(t *testing.T) {
		order := &model.Order{
			ID:             "order_002",
			BuyerID:        "buyer_1",
			SessionID:      "session_1",
			ProductID:      "prod_001",
			SKU:            "SHOE-001",
			Quantity:       1,
			AmountPaisa:    249900,
			Currency:       "INR",
			Status:         model.OrderStatusPending,
			IdempotencyKey: "idem_002",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		err := repo.CreateOrder(ctx, order)
		if err != nil {
			t.Fatal(err)
		}

		order2 := &model.Order{
			ID:             "order_002_duplicate",
			BuyerID:        "buyer_1",
			SessionID:      "session_1",
			ProductID:      "prod_001",
			SKU:            "SHOE-001",
			Quantity:       1,
			AmountPaisa:    249900,
			Currency:       "INR",
			Status:         model.OrderStatusPending,
			IdempotencyKey: "idem_002",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err = repo.CreateOrder(ctx, order2)
		if err != nil {
			t.Fatal(err)
		}

		orders, err := repo.GetOrdersByBuyer(ctx, "buyer_1", 10)
		if err != nil {
			t.Fatal(err)
		}

		if len(orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(orders))
		}
	})

	t.Run("GetByIdempotencyKey", func(t *testing.T) {
		order, err := repo.GetByIdempotencyKey(ctx, "idem_001")
		if err != nil {
			t.Fatal(err)
		}
		if order == nil {
			t.Fatal("expected order, got nil")
		}
		if order.IdempotencyKey != "idem_001" {
			t.Errorf("expected idempotency key=idem_001, got %s", order.IdempotencyKey)
		}
		if order.ID != "order_001" {
			t.Errorf("expected ID=order_001, got %s", order.ID)
		}
	})

	t.Run("GetByIdempotencyKey_not_found", func(t *testing.T) {
		order, err := repo.GetByIdempotencyKey(ctx, "nonexistent")
		if err != nil {
			t.Fatal(err)
		}
		if order != nil {
			t.Errorf("expected nil order, got %v", order)
		}
	})

	t.Run("GetOrdersByBuyer", func(t *testing.T) {
		orders, err := repo.GetOrdersByBuyer(ctx, "buyer_1", 10)
		if err != nil {
			t.Fatal(err)
		}
		if len(orders) != 2 {
			t.Errorf("expected 2 orders, got %d", len(orders))
		}
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		order := &model.Order{
			ID:             "order_003",
			BuyerID:        "buyer_2",
			SessionID:      "session_2",
			ProductID:      "prod_002",
			SKU:            "SHOE-002",
			Quantity:       2,
			AmountPaisa:    499800,
			Currency:       "INR",
			Status:         model.OrderStatusPending,
			IdempotencyKey: "idem_003",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := repo.CreateOrder(ctx, order)
		if err != nil {
			t.Fatal(err)
		}

		order.Status = model.OrderStatusPaid
		order.RazorpayOrderID = "rzp_order_123"
		order.RazorpayPaymentID = "rzp_payment_456"

		err = repo.UpdateOrder(ctx, order)
		if err != nil {
			t.Fatal(err)
		}

		retrieved, err := repo.GetOrder(ctx, "order_003")
		if err != nil {
			t.Fatal(err)
		}
		if retrieved.Status != model.OrderStatusPaid {
			t.Errorf("expected status=PAID, got %s", retrieved.Status)
		}
		if retrieved.RazorpayOrderID != "rzp_order_123" {
			t.Errorf("expected RazorpayOrderID=rzp_order_123, got %s", retrieved.RazorpayOrderID)
		}
		if retrieved.RazorpayPaymentID != "rzp_payment_456" {
			t.Errorf("expected RazorpayPaymentID=rzp_payment_456, got %s", retrieved.RazorpayPaymentID)
		}
	})

	t.Run("UpdateOrder_not_found", func(t *testing.T) {
		order := &model.Order{
			ID:        "nonexistent",
			Status:    model.OrderStatusPaid,
			UpdatedAt: time.Now(),
		}
		err := repo.UpdateOrder(ctx, order)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != model.ErrOrderNotFound {
			t.Errorf("expected ErrOrderNotFound, got %v", err)
		}
	})
}
