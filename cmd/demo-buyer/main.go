package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	appmcp "github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
	"github.com/razorpay/aegis/internal/pkg/razorpay_mcp"
)

type Goal struct {
	Query        string  `json:"query"`
	MaxPriceINR  float64 `json:"max_price_inr,omitempty"`
	Category     string  `json:"category,omitempty"`
	Color        string  `json:"color,omitempty"`
	Size         string  `json:"size,omitempty"`
	Brand        string  `json:"brand,omitempty"`
	BuyerID      string  `json:"buyer_id"`
	BuyerPincode string  `json:"buyer_pincode"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: demo-buyer <config.yaml> [goal.json]")
		fmt.Println("Example: demo-buyer config.yaml")
		fmt.Println("         demo-buyer config.yaml goal.json")
		os.Exit(1)
	}

	cfg, err := config.Load(os.Args[1])
	if err != nil {
		log.Printf("load config: %v", err)
		os.Exit(1)
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

	catalogRepo := repository.NewCatalogPG(db)
	orderRepo := repository.NewOrderPG(db)
	policyRepo := repository.NewPolicyPG(db)
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
		log,
	)

	aegisClient := &directAegisClient{gatewayService: gatewayService}

	merchantService := service.NewMerchantMCPService(catalogRepo, orderRepo, aegisClient)

	// Parse goal
	var goal Goal
	if len(os.Args) > 2 {
		data, err := os.ReadFile(os.Args[2])
		if err != nil {
			log.Error("read goal", "error", err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &goal); err != nil {
			log.Error("parse goal", "error", err)
			os.Exit(1)
		}
	} else {

		goal = Goal{
			Query:        "running shoes",
			MaxPriceINR:  3000,
			Category:     "footwear",
			BuyerID:      "demo_buyer_1",
			BuyerPincode: "400001",
		}
	}

	ctx := context.Background()
	runDemoBuyer(ctx, merchantService, goal)
}

func runDemoBuyer(ctx context.Context, merchantService service.MerchantMCPService, goal Goal) {
	fmt.Println("=== AI BUYER DEMO ===")
	fmt.Printf("Goal: %s\n", goal.Query)
	if goal.MaxPriceINR > 0 {
		fmt.Printf("Max Price: ₹%.2f\n", goal.MaxPriceINR)
	}
	if goal.Category != "" {
		fmt.Printf("Category: %s\n", goal.Category)
	}
	if goal.Color != "" {
		fmt.Printf("Color: %s\n", goal.Color)
	}
	if goal.Size != "" {
		fmt.Printf("Size: %s\n", goal.Size)
	}
	if goal.Brand != "" {
		fmt.Printf("Brand: %s\n", goal.Brand)
	}
	fmt.Printf("Buyer: %s\n\n", goal.BuyerID)

	fmt.Println("Searching for products...")
	maxPricePaisa := int64(goal.MaxPriceINR * 100)
	searchResult, err := merchantService.SearchProducts(ctx, appmcp.SearchProductsParams{
		Query:       goal.Query,
		Category:    goal.Category,
		MaxPrice:    &maxPricePaisa,
		InStockOnly: true,
		Limit:       10,
		Color:       goal.Color,
		Size:        goal.Size,
		Brand:       goal.Brand,
	})
	if err != nil {
		log.Printf("search failed: %v", err)
		os.Exit(1)
	}

	if len(searchResult.Products) == 0 {
		fmt.Println("No products found matching criteria")
		return
	}

	fmt.Printf("Found %d products:\n", len(searchResult.Products))
	for i, p := range searchResult.Products {
		fmt.Printf("  %d. %s (SKU: %s) - ₹%.2f - %s\n",
			i+1, p.Name, p.SKU, float64(p.MinPrice)/100, p.Category)
	}

	selected := searchResult.Products[0]
	fmt.Printf("\nChecking availability for %s (%s)...\n", selected.Name, selected.SKU)

	availResult, err := merchantService.CheckAvailability(ctx, appmcp.CheckAvailabilityParams{
		SKU: selected.SKU,
	})
	if err != nil {
		log.Printf("check availability failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Available: %d, Reserved: %d\n", availResult.Available, availResult.Reserved)

	if availResult.Available == 0 {
		fmt.Println("Out of stock!")
		return
	}

	fmt.Printf("\n[INFO] Getting product details for %s...\n", selected.ID)
	product, err := merchantService.GetProduct(ctx, appmcp.GetProductParams{
		ProductID: selected.ID,
	})
	if err != nil {
		log.Printf("get product failed: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Product: %s\n", product.Name)
	fmt.Printf("Description: %s\n", product.Description)
	fmt.Printf("Category: %s\n", product.Category)
	fmt.Println("Offers:")
	for _, offer := range product.Offers {
		fmt.Printf("  SKU: %s, Price: ₹%.2f, Size: %s, Color: %s, Inventory: %d\n",
			offer.SKU, float64(offer.PricePaisa)/100, offer.Size, offer.Color, offer.Inventory)
	}

	// Select best matching offer
	var chosenOffer *model.Offer
	for _, offer := range product.Offers {
		match := true
		if goal.Size != "" && offer.Size != "" && !strings.EqualFold(offer.Size, goal.Size) {
			match = false
		}
		if goal.Color != "" && offer.Color != "" && !strings.EqualFold(offer.Color, goal.Color) {
			match = false
		}
		if match {
			chosenOffer = &offer
			break
		}
	}

	if chosenOffer == nil && len(product.Offers) > 0 {
		chosenOffer = &product.Offers[0]
	}

	if chosenOffer == nil {
		fmt.Println("No suitable offer found")
		return
	}

	fmt.Printf("\n Selected: %s (Size: %s, Color: %s) - ₹%.2f\n",
		chosenOffer.SKU, chosenOffer.Size, chosenOffer.Color, float64(chosenOffer.PricePaisa)/100)

	fmt.Println("\n Initiating purchase...")
	sessionID := fmt.Sprintf("session_%s_%d", goal.BuyerID, time.Now().Unix())
	idempotencyKey := fmt.Sprintf("idem_%s_%d", goal.BuyerID, time.Now().UnixNano())

	reasoning := fmt.Sprintf(
		"Selected %s (%s) matching buyer's %s request. "+
			"Price ₹%.2f is within ₹%.2f budget. "+
			"Choosing %s colorway, size %s based on search criteria.",
		chosenOffer.SKU, product.Name, goal.Query,
		float64(chosenOffer.PricePaisa)/100, goal.MaxPriceINR,
		chosenOffer.Color, chosenOffer.Size,
	)
	metadata, _ := json.Marshal(map[string]any{
		"reasoning":  reasoning,
		"confidence": 0.92,
		"goal_query": goal.Query,
		"budget_inr": goal.MaxPriceINR,
	})

	purchaseResult, err := merchantService.Purchase(ctx, appmcp.PurchaseParams{
		BuyerID:        goal.BuyerID,
		SessionID:      sessionID,
		ProductID:      selected.ID,
		SKU:            chosenOffer.SKU,
		Quantity:       1,
		IdempotencyKey: idempotencyKey,
		BuyerPincode:   goal.BuyerPincode,
		Metadata:       metadata,
	})
	if err != nil {
		log.Printf("purchase failed: %v", err)
		os.Exit(1)
	}

	fmt.Println("\n=== PURCHASE RESULT ===")
	fmt.Printf("Allowed: %v\n", purchaseResult.Allowed)
	fmt.Printf("Status: %s\n", purchaseResult.Status)
	fmt.Printf("Reason: %s\n", purchaseResult.Reason)

	if purchaseResult.Allowed {
		fmt.Printf("Order ID: %s\n", purchaseResult.OrderID)
		fmt.Printf("Payment ID: %s\n", purchaseResult.PaymentID)
		fmt.Printf("\n PURCHASE SUCCESSFUL!\n")

		fmt.Printf("\n Checking order status for %s...\n", purchaseResult.OrderID)
		orderStatus, err := merchantService.GetOrderStatus(ctx, appmcp.GetOrderStatusParams{
			OrderID: purchaseResult.OrderID,
		})
		if err != nil {
			log.Printf("get order status failed: %v", err)
		} else {
			fmt.Printf("Order Status: %s\n", orderStatus.Status)
			fmt.Printf("Amount: ₹%.2f\n", float64(orderStatus.AmountPaisa)/100)
			for _, item := range orderStatus.Items {
				fmt.Printf("  Item: SKU=%s, Qty=%d, Price=₹%.2f\n",
					item.SKU, item.Quantity, float64(item.PricePaisa)/100)
			}
		}
	} else if purchaseResult.Status == "PENDING_APPROVAL" {
		fmt.Printf("Approval Queue ID: %s\n", purchaseResult.ApprovalQueueID)
		fmt.Println("\n Purchase requires approval. Check approval queue dashboard.")
	} else {
		fmt.Println("\n PURCHASE BLOCKED")
	}

	fmt.Println("\n=== DEMO COMPLETE ===")
}

// directAegisClient implements AegisMCPClient using direct service call.
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
		RuleFired:       result.RuleFired,
		Status:          result.Status,
		OrderID:         result.OrderID,
		PaymentID:       result.PaymentID,
		ApprovalQueueID: result.ApprovalQueueID,
		Remaining:       result.Remaining,
	}, nil
}
