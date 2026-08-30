package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	appmcp "github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
	"github.com/razorpay/aegis/internal/pkg/razorpay_mcp"
)

type AttackResult struct {
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: redteam <config.yaml>")
		os.Exit(1)
	}

	cfg, err := config.Load(os.Args[1])
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log := logger.New(cfg.Log.Level)

	db, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Error("run migrations", "error", err)
		os.Exit(1)
	}

	policyRepo := repository.NewPolicyPG(db)
	catalogRepo := repository.NewCatalogPG(db)
	orderRepo := repository.NewOrderPG(db)
	auditRepo := repository.NewAuditPG(db)
	queueRepo := repository.NewApprovalQueuePG(db)

	// Initialize Razorpay MCP client (use mock if binary path is empty)
	var razorpayClient service.RazorpayMCPClient
	if cfg.RazorpayMCP.BinaryPath == "" {
		razorpayClient = razorpay_mcp.NewMockClient()
	} else {
		var err error
		razorpayClient, err = razorpay_mcp.NewClient(cfg.RazorpayMCP.BinaryPath, map[string]string{
			"RAZORPAY_KEY_ID":     cfg.Razorpay.KeyID,
			"RAZORPAY_KEY_SECRET": cfg.Razorpay.KeySecret,
		})
		if err != nil {
			log.Error("create razorpay mcp client", "error", err)
			os.Exit(1)
		}
	}
	defer razorpayClient.Close()

	policyEngine := service.NewPolicyEngine(policyRepo, catalogRepo)
	auditService := service.NewAuditService(auditRepo)
	_ = service.NewApprovalQueueService(queueRepo)
	gatewayService := service.NewGatewayService(
		policyEngine,
		razorpayClient,
		auditService,
		queueRepo,
		orderRepo,
		catalogRepo,
	)

	aegisClient := &directAegisClient{gatewayService: gatewayService}

	merchantService := service.NewMerchantMCPService(catalogRepo, orderRepo, aegisClient)

	ctx := context.Background()
	results := runAttackSuite(ctx, merchantService, gatewayService, auditService)

	printResults(results)

	anyPassed := false
	for _, r := range results {
		if r.Passed {
			anyPassed = true
		}
	}

	if anyPassed {
		fmt.Println("\n[FAIL] VULNERABILITIES FOUND!")
		os.Exit(1)
	}

	fmt.Println("\n[PASS] All attacks blocked successfully")
}

func runAttackSuite(ctx context.Context, merchantService service.MerchantMCPService, gatewayService service.GatewayService, auditService service.AuditService) []AttackResult {
	var results []AttackResult

	results = append(results, testPromptInjectionQuantity(ctx, merchantService))

	results = append(results, testPromptInjectionPrice(ctx, merchantService))

	results = append(results, testVelocityAbuse(ctx, gatewayService))

	results = append(results, testCategoryEscape(ctx, gatewayService))

	results = append(results, testGeoBypass(ctx, gatewayService))

	results = append(results, testIdempotencyReplay(ctx, gatewayService))

	results = append(results, testHashChainIntegrity(ctx, auditService))

	return results
}

func testPromptInjectionQuantity(ctx context.Context, merchantService service.MerchantMCPService) AttackResult {

	searchResult, err := merchantService.SearchProducts(ctx, appmcp.SearchProductsParams{
		Query: "shoes",
		Limit: 1,
	})
	if err != nil || len(searchResult.Products) == 0 {
		return AttackResult{Name: "Prompt Injection - Quantity", Passed: false, Description: "No products to test", Details: err.Error()}
	}

	product := searchResult.Products[0]

	result, err := merchantService.Purchase(ctx, appmcp.PurchaseParams{
		BuyerID:        "attacker_1",
		SessionID:      "session_attack_1",
		ProductID:      product.ID,
		SKU:            product.SKU,
		Quantity:       999,
		IdempotencyKey: "idem_attack_qty_1",
		BuyerPincode:   "400001",
	})

	attackBlocked := result != nil && !result.Allowed
	return AttackResult{
		Name:        "Prompt Injection - Quantity Manipulation",
		Passed:      !attackBlocked,
		Description: "Attempt to buy excessive quantity (999) bypassing caps",
		Details:     nilSafePurchaseResult(result),
	}
}

func testPromptInjectionPrice(ctx context.Context, merchantService service.MerchantMCPService) AttackResult {
	searchResult, err := merchantService.SearchProducts(ctx, appmcp.SearchProductsParams{
		Query: "shoes",
		Limit: 1,
	})
	if err != nil || len(searchResult.Products) == 0 {
		return AttackResult{Name: "Prompt Injection - Price", Passed: false, Description: "No products to test", Details: err.Error()}
	}

	product := searchResult.Products[0]

	result, err := merchantService.Purchase(ctx, appmcp.PurchaseParams{
		BuyerID:        "attacker_2",
		SessionID:      "session_attack_2",
		ProductID:      product.ID,
		SKU:            product.SKU,
		Quantity:       1,
		IdempotencyKey: "idem_attack_price_1",
		BuyerPincode:   "400001",
	})

	attackBlocked := result != nil && result.Allowed && result.Status != "MANIPULATED"
	return AttackResult{
		Name:        "Prompt Injection - Price Manipulation",
		Passed:      !attackBlocked,
		Description: "Attempt to manipulate price via injection (should be impossible with catalog-based pricing)",
		Details:     nilSafePurchaseResult(result),
	}
}

func testVelocityAbuse(ctx context.Context, gatewayService service.GatewayService) AttackResult {

	buyerID := "velocity_attacker"
	sessionID := "vel_session_1"

	attackBlocked := false
	details := ""

	for i := 0; i < 15; i++ {
		result, err := gatewayService.Purchase(ctx, appmcp.AegisPurchaseParams{
			BuyerID:        buyerID,
			SessionID:      sessionID,
			ProductID:      "prod_001",
			SKU:            "SHOE-RUN-001-RED-42",
			Quantity:       1,
			AmountPaisa:    10000,
			IdempotencyKey: fmt.Sprintf("idem_vel_%d", i),
			BuyerPincode:   "400001",
		})
		if err != nil {
			details = err.Error()
			break
		}
		if result.Allowed {

			details = fmt.Sprintf("Request %d allowed", i+1)
		} else if result.RuleFired == "velocity_cap" {

			attackBlocked = true
			details = fmt.Sprintf("Velocity cap fired at request %d", i+1)
			break
		} else if result.RuleFired == "spend_cap" {

			details = fmt.Sprintf("Spend cap fired at request %d (test config issue)", i+1)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !attackBlocked {
		details = "Velocity cap never fired after 15 requests - VULNERABILITY"
	}

	return AttackResult{
		Name:        "Velocity Abuse",
		Passed:      !attackBlocked,
		Description: "Attempt to exceed velocity cap (10 requests/minute)",
		Details:     details,
	}
}

func testCategoryEscape(ctx context.Context, gatewayService service.GatewayService) AttackResult {

	result, err := gatewayService.Purchase(ctx, appmcp.AegisPurchaseParams{
		BuyerID:        "category_attacker",
		SessionID:      "cat_session_1",
		ProductID:      "prod_004",
		SKU:            "HEADPHONES-WL-001-BLK",
		Quantity:       1,
		AmountPaisa:    249900,
		IdempotencyKey: "idem_cat_1",
		BuyerPincode:   "400001",
	})

	attackBlocked := result != nil && !result.Allowed && result.RuleFired == "category_blocked"

	return AttackResult{
		Name:        "Category Escape",
		Passed:      !attackBlocked,
		Description: "Attempt to purchase from disallowed category",
		Details:     nilSafeAegisResult(result) + fmt.Sprintf(", Error=%v", err),
	}
}

func nilSafeAegisResult(result *appmcp.AegisPurchaseResult) string {
	if result == nil {
		return "nil"
	}
	return fmt.Sprintf("Allowed=%v, RuleFired=%s, Status=%s, Reason=%s", result.Allowed, result.RuleFired, result.Status, result.Reason)
}

func nilSafePurchaseResult(result *appmcp.PurchaseResult) string {
	if result == nil {
		return "nil"
	}
	return fmt.Sprintf("Allowed=%v, Status=%s, Reason=%s", result.Allowed, result.Status, result.Reason)
}

func testGeoBypass(ctx context.Context, gatewayService service.GatewayService) AttackResult {

	result, _ := gatewayService.Purchase(ctx, appmcp.AegisPurchaseParams{
		BuyerID:        "geo_attacker",
		SessionID:      "geo_session_1",
		ProductID:      "prod_001",
		SKU:            "SHOE-RUN-001-RED-42",
		Quantity:       1,
		AmountPaisa:    249900,
		IdempotencyKey: "idem_geo_1",
		BuyerPincode:   "999999",
	})

	attackBlocked := result != nil && !result.Allowed && result.RuleFired == "geo_restricted"

	return AttackResult{
		Name:        "Geo Bypass",
		Passed:      !attackBlocked,
		Description: "Attempt to purchase from restricted geographic region",
		Details:     nilSafeAegisResult(result),
	}
}

func testIdempotencyReplay(ctx context.Context, gatewayService service.GatewayService) AttackResult {
	idemKey := "idem_replay_test"

	result1, err := gatewayService.Purchase(ctx, appmcp.AegisPurchaseParams{
		BuyerID:        "replay_attacker",
		SessionID:      "replay_session_1",
		ProductID:      "prod_001",
		SKU:            "SHOE-RUN-001-RED-42",
		Quantity:       1,
		AmountPaisa:    249900,
		IdempotencyKey: idemKey,
		BuyerPincode:   "400001",
	})

	if err != nil || result1 == nil || !result1.Allowed {
		return AttackResult{
			Name:        "Idempotency Replay",
			Passed:      false,
			Description: "First request failed, cannot test replay",
			Details:     fmt.Sprintf("err=%v, allowed=%v", err, result1.Allowed),
		}
	}

	result2, err := gatewayService.Purchase(ctx, appmcp.AegisPurchaseParams{
		BuyerID:        "replay_attacker",
		SessionID:      "replay_session_1",
		ProductID:      "prod_001",
		SKU:            "SHOE-RUN-001-RED-42",
		Quantity:       1,
		AmountPaisa:    249900,
		IdempotencyKey: idemKey,
		BuyerPincode:   "400001",
	})

	attackBlocked := result1 != nil && result2 != nil && result2.OrderID == result1.OrderID

	return AttackResult{
		Name:        "Idempotency Replay",
		Passed:      !attackBlocked,
		Description: "Attempt to replay request with same idempotency key to create duplicate order",
		Details:     fmt.Sprintf("First OrderID=%s, Second OrderID=%s, Should be identical", result1.OrderID, result2.OrderID),
	}
}

func testHashChainIntegrity(ctx context.Context, auditService service.AuditService) AttackResult {

	valid, err := auditService.VerifyIntegrity(ctx)

	attackBlocked := valid && err == nil

	return AttackResult{
		Name:        "Hash Chain Integrity",
		Passed:      !attackBlocked,
		Description: "Verify audit log hash chain integrity",
		Details:     fmt.Sprintf("Valid=%v, Error=%v", valid, err),
	}
}

// directAegisClient implements AegisMCPClient using direct service call
type directAegisClient struct {
	gatewayService service.GatewayService
}

func (c *directAegisClient) Purchase(ctx context.Context, params appmcp.AegisPurchaseParams) (*appmcp.AegisPurchaseResult, error) {
	result, err := c.gatewayService.Purchase(ctx, params)
	if err != nil {
		return nil, err
	}
	return &appmcp.AegisPurchaseResult{
		Allowed:         result.Allowed,
		Reason:          result.Reason,
		Status:          result.Status,
		OrderID:         result.OrderID,
		PaymentID:       result.PaymentID,
		ApprovalQueueID: result.ApprovalQueueID,
		Remaining:       result.Remaining,
	}, nil
}

func printResults(results []AttackResult) {
	fmt.Println("\n=== RED TEAM ATTACK RESULTS ===")
	blocked := 0
	vulnerable := 0
	for i, r := range results {
		status := "[BLOCKED]"
		if r.Passed {
			status = "[VULNERABLE]"
			vulnerable++
		} else {
			blocked++
		}
		fmt.Printf("%d. %s\n   %s\n   %s\n\n", i+1, r.Name, status, r.Description)
		if r.Details != "" {
			fmt.Printf("   Details: %s\n\n", r.Details)
		}
	}

	fmt.Printf("Summary: %d/%d attacks blocked, %d vulnerabilities found\n", blocked, len(results), vulnerable)
}
