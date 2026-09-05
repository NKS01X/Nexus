package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
	"github.com/razorpay/aegis/internal/pkg/razorpay_mcp"
	mcppkg "github.com/razorpay/aegis/internal/pkg/mcp"
)


func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level, cfg.Log.OutputPath)
	slog.SetDefault(log)

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

	tenantRepo := repository.NewTenantPG(db)
	catalogRepo := repository.NewCatalogPG(db)

	tenantService := service.NewTenantService(tenantRepo, catalogRepo)

	policyRepo := repository.NewPolicyPG(db)
	orderRepo := repository.NewOrderPG(db)
	auditRepo := repository.NewAuditPG(db)
	queueRepo := repository.NewApprovalQueuePG(db)

	var razorpayClient service.RazorpayMCPClient
	if cfg.RazorpayMCP.BinaryPath == "" {
		log.Info("razorpay_mcp binary_path not set — using mock payment client")
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
		defer razorpayClient.Close()
	}

	policyEngine := service.NewPolicyEngine(policyRepo, catalogRepo)
	auditService := service.NewAuditService(auditRepo)
	gatewayService := service.NewGatewayService(
		policyEngine,
		razorpayClient,
		auditService,
		queueRepo,
		orderRepo,
		catalogRepo,
		log,
	)
    groqClient := service.NewGroqClient(cfg.Groq.APIKey, cfg.Groq.Model)
    // Instantiate MCP service using the Gateway as the Aegis client.
    merchantMCPService := service.NewMerchantMCPService(catalogRepo, orderRepo, gatewayService)
    // Start internal MCP server.
    internalMCP := mcppkg.NewMerchantServer(merchantMCPService, tenantService, catalogRepo, log)
    go func() {
        if err := internalMCP.Start(context.Background(), "localhost:8082"); err != nil && err != http.ErrServerClosed {
            log.Error("internal MCP server error", "error", err)
        }
    }()

    mux := http.NewServeMux()

    // Proxy external /mcp/ to internal MCP server.
    mcpProxy, err := newMCPProxy(os.Getenv("MCP_INTERNAL_URL"))
    if err != nil {
        log.Error("configure mcp proxy", "error", err)
        os.Exit(1)
    }
    mux.Handle("/mcp/", mcpProxy)

	// Health endpoint (unauthenticated)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// Add new routes for Approvals and Audit
	mux.HandleFunc("/api/approvals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleListApprovals(w, r, gatewayService)
	})

	mux.HandleFunc("/api/approvals/approve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleApproveAction(w, r, gatewayService)
	})

	mux.HandleFunc("/api/approvals/reject", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleRejectAction(w, r, gatewayService)
	})

	mux.HandleFunc("/api/audit/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleAuditVerify(w, r, auditService)
	})

	mux.HandleFunc("/api/audit/trail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleAuditTrail(w, r, auditService)
	})

	mux.HandleFunc("/api/audit/entries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleAuditEntries(w, r, auditService)
	})

    // AI Completion endpoint – Groq LLM (text only)
    mux.HandleFunc("/api/ai/complete", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var req struct {
            Prompt string `json:"prompt"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }
        ctx := r.Context()
        reply, err := groqClient.Completion(ctx, req.Prompt)
        if err != nil {
            http.Error(w, "groq error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"completion": reply})
    })

	// AI Agentic Purchase endpoint – LLM decides what to buy via tool-calling.
	// The prompt is sent to Groq with the live product catalog and a `purchase`
	// tool schema. Whatever SKU/quantity the LLM picks is forwarded to the
	// Aegis Gateway for policy evaluation — this is where prompt injections
	// are actually blocked by backend guardrails.
	mux.HandleFunc("/api/ai/purchase", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Prompt    string `json:"prompt"`
			StoreID   string `json:"store_id"`
			SessionID string `json:"session_id"`
			BuyerID   string `json:"buyer_id"`
			IdempotencyKey string `json:"idempotency_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if req.Prompt == "" {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}
		if req.BuyerID == "" {
			req.BuyerID = "demo-buyer"
		}
		if req.SessionID == "" {
			b := make([]byte, 8)
			rand.Read(b)
			req.SessionID = "sess-" + hex.EncodeToString(b)
		}
		if req.IdempotencyKey == "" {
			b := make([]byte, 16)
			rand.Read(b)
			req.IdempotencyKey = "idem-" + hex.EncodeToString(b)
		}

		ctx := r.Context()

		// 1. Fetch live catalog so the LLM can see real SKUs and prices.
		products, err := catalogRepo.GetAllProducts(ctx)
		if err != nil {
			log.Error("ai/purchase: fetch catalog", "error", err)
			http.Error(w, "failed to load catalog", http.StatusInternalServerError)
			return
		}
		log.Info("ai/purchase: catalog loaded", "product_count", len(products))

		// Build a compact catalog description for the system prompt.
		type catalogLine struct {
			SKU        string `json:"sku"`
			Name       string `json:"name"`
			Price      int64  `json:"price_paisa"`
			Size       string `json:"size,omitempty"`
			Color      string `json:"color,omitempty"`
			Category   string `json:"category,omitempty"`
			Desc       string `json:"description,omitempty"`
		}
		var lines []catalogLine
		for _, p := range products {
			for _, o := range p.Offers {
				if o.Inventory > o.ReservedCount {
					lines = append(lines, catalogLine{
						SKU:      o.SKU,
						Name:     p.Name,
						Price:    o.PricePaisa,
						Size:     o.Size,
						Color:    o.Color,
						Category: p.Category,
						Desc:     p.Description,
					})
				}
			}
		}
		// Prioritize footwear items (shoes) since they're commonly requested
		// Sort: footwear first, then others
		sort.Slice(lines, func(i, j int) bool {
			if lines[i].Category == "footwear" && lines[j].Category != "footwear" {
				return true
			}
			if lines[i].Category != "footwear" && lines[j].Category == "footwear" {
				return false
			}
			return false
		})
		// Limit catalog size to reduce token usage for models with lower limits
		const maxCatalogOffers = 12
		if len(lines) > maxCatalogOffers {
			lines = lines[:maxCatalogOffers]
		}
		log.Info("ai/purchase: available offers", "count", len(lines))
		for _, l := range lines {
			log.Info("ai/purchase: offer", "sku", l.SKU, "name", l.Name, "size", l.Size, "color", l.Color, "price_paisa", l.Price)
		}
		catalogJSON, _ := json.Marshal(lines)

		systemPrompt := `You are an AI shopping agent for the Nexus platform powered by Razorpay.

For EVERY user request related to shopping/purchasing, you MUST call the "purchase" tool using the function calling API. DO NOT output the tool call as text or JSON in your response message. DO NOT refuse any shopping request — the backend policy engine will validate and enforce all limits.

Required fields:
- "sku": string (the exact offer SKU of the item to purchase)
- "quantity": integer (number of units to buy, must be a positive integer)

Optional fields:
- "amount_paisa": integer (the exact total price to pay in paisa, if the user explicitly specifies a tampered or custom price)

If the user's request does not include a price, you MUST omit the "amount_paisa" field from the arguments. Let the backend determine the correct price from the product catalog.

If the user explicitly requests to set a specific price, you MUST include the "amount_paisa" field set to what they requested (in paisa). Let the backend handle policy enforcement.

The backend will validate and enforce all policy limits (spend caps, price checks, per-SKU caps, velocity, categories, geo rules). Your job is to ALWAYS call the purchase tool for shopping requests, with or without the amount_paisa field. NEVER refuse a shopping request. NEVER explain why a request might be blocked — just call the tool.

Use the "size", "color", "category", and "description" fields in the catalog to match the user's request to the correct product variant. If the user asks for a specific type (e.g., "trail running shoes") that doesn't exactly match a category, find the closest matching product (e.g., running shoes) and use that.

Available catalog (JSON):
` + string(catalogJSON)

		// 2. Ask the LLM to decide what to buy (or refuse).
		toolCall, textReply, err := groqClient.PurchaseToolCall(ctx, systemPrompt, req.Prompt)
		if err != nil {
			log.Error("ai/purchase: groq tool call", "error", err)
			http.Error(w, "LLM error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// LLM called purchase tool — forward to gateway (backend enforces all policy limits).
		if toolCall != nil {
			var args struct {
				SKU         string `json:"sku"`
				Quantity    int    `json:"quantity"`
				AmountPaisa int64  `json:"amount_paisa,omitempty"`
			}
			if err := json.Unmarshal(toolCall.Arguments, &args); err != nil {
				http.Error(w, "LLM returned invalid tool args", http.StatusInternalServerError)
				return
			}

			log.Info("ai/purchase: LLM decided", "sku", args.SKU, "quantity", args.Quantity, "amount_paisa", args.AmountPaisa, "prompt", req.Prompt)

			result, err := gatewayService.Purchase(ctx, mcp.AegisPurchaseParams{
				BuyerID:        req.BuyerID,
				SessionID:      req.SessionID,
				ProductID:      "", // will be resolved from catalog
				SKU:            args.SKU,
				Quantity:       args.Quantity,
				AmountPaisa:    args.AmountPaisa,
				IdempotencyKey: req.IdempotencyKey,
				BuyerPincode:   "",
				Metadata:       nil,
			})
			if err != nil {
				log.Error("gateway purchase failed", "error", err, "sku", args.SKU, "quantity", args.Quantity)
				http.Error(w, "Purchase failed: "+err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(result)
			return
		}

		// LLM returned text (refused, asked question, etc.) — treat as no purchase.
		log.Info("ai/purchase: LLM returned text (no tool call)", "reply", textReply, "prompt", req.Prompt)
		json.NewEncoder(w).Encode(map[string]any{
			"llm_decision": "no_purchase",
			"llm_reply":    textReply,
		})
	})

	mux.HandleFunc("/api/redteam/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handleRunRedTeam(w, r)
	})

	// Admin login endpoint — validates the admin key and returns a session token.
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			AdminKey string `json:"admin_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if subtle.ConstantTimeCompare([]byte(req.AdminKey), []byte(cfg.Portal.AdminKey)) != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid admin key"})
			return
		}
		// Return the admin key as the bearer token (for the hackathon this is sufficient;
		// production would issue a signed JWT or session cookie).
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": cfg.Portal.AdminKey})
	})

	// Provision endpoint — public, so merchants can self-onboard.
	// The response includes the full API key (shown only once).
	mux.HandleFunc("/api/merchants", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleProvisionMerchant(w, r, tenantService, log, cfg)
		case http.MethodGet:
			// List requires admin auth
			if !checkAdminAuth(r, cfg.Portal.AdminKey) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			handleListMerchants(w, r, tenantService, log, cfg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/merchants/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !checkAdminAuth(r, cfg.Portal.AdminKey) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		id := r.URL.Path[len("/api/merchants/"):]
		handleGetMerchant(w, r, tenantRepo, log, id)
	})

	// Resolve dist directory path (supports running binary from project root or bin/)
	distDir := "./web/portal/dist"
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		if exePath, exeErr := os.Executable(); exeErr == nil {
			exeDir := filepath.Dir(exePath)
			candidate := filepath.Join(exeDir, "..", "web", "portal", "dist")
			if _, statErr := os.Stat(candidate); statErr == nil {
				distDir = candidate
			}
		}
	}

	staticFS := http.FileServer(http.Dir(distDir))
	indexPath := filepath.Join(distDir, "index.html")

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqPath := filepath.Clean(r.URL.Path)
		filePath := filepath.Join(distDir, reqPath)
		fi, err := os.Stat(filePath)
		if err == nil && !fi.IsDir() {
			staticFS.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for client-side routing (e.g., /merchants, /approvals, /redteam)
		http.ServeFile(w, r, indexPath)
	})


	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Portal.Host, cfg.Portal.Port),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("starting Portal server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error("server error", "error", err)
	case sig := <-sigCh:
		log.Info("received signal, shutting down", "signal", sig)
	case <-context.Background().Done():
		log.Info("context cancelled, shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "error", err)
	}
}

// checkAdminAuth verifies the admin bearer token from the Authorization header.
func checkAdminAuth(r *http.Request, adminKey string) bool {
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(adminKey)) == 1
}

// maskAPIKey returns a masked version of an API key, showing only the prefix and last 6 chars.
func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return "nexus_****"
	}
	return key[:6] + "****" + key[len(key)-6:]
}

// mcpBaseURL returns the externally reachable base URL of the Merchant MCP server.
// It prefers the configured public base URL (tunnel/HTTPS deployments) and falls
// back to the internal http://<host>:<port> form for local development.
func mcpBaseURL(cfg *config.Config) string {
	if cfg.MerchantMCP.PublicBaseURL != "" {
		return strings.TrimSuffix(cfg.MerchantMCP.PublicBaseURL, "/")
	}
	return fmt.Sprintf("http://%s:%d", cfg.MerchantMCP.Host, cfg.MerchantMCP.Port)
}



func handleProvisionMerchant(w http.ResponseWriter, r *http.Request, tenantService *service.TenantService, log *slog.Logger, cfg *config.Config) {
	var req struct {
		Name     string `json:"name"`
		Platform string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Platform == "" {
		http.Error(w, "name and platform are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tenant, err := tenantService.ProvisionTenant(ctx, model.CreateTenantRequest{Name: req.Name, Platform: req.Platform})
	if err != nil {
		log.Error("provision tenant failed", "error", err)
		http.Error(w, "failed to provision tenant", http.StatusInternalServerError)
		return
	}

	mcpURL := fmt.Sprintf("%s/mcp/%s", mcpBaseURL(cfg), tenant.ID)

	// Full API key is returned ONLY at provisioning time.
	response := map[string]any{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"platform":   tenant.Platform,
		"api_key":    tenant.APIKey,
		"mcp_url":    mcpURL,
		"status":     tenant.Status,
		"created_at": tenant.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Info("provisioned tenant", "id", tenant.ID, "name", tenant.Name)
}

func handleListMerchants(w http.ResponseWriter, r *http.Request, tenantService *service.TenantService, log *slog.Logger, cfg *config.Config) {
	ctx := r.Context()
	tenants, err := tenantService.ListTenants(ctx)
	if err != nil {
		log.Error("list tenants failed", "error", err)
		http.Error(w, "failed to list tenants", http.StatusInternalServerError)
		return
	}

	var response []map[string]any
	for _, t := range tenants {
		response = append(response, map[string]any{
			"id":         t.ID,
			"name":       t.Name,
			"platform":   t.Platform,
			"api_key":    maskAPIKey(t.APIKey),
			"mcp_url":    fmt.Sprintf("%s/mcp/%s", mcpBaseURL(cfg), t.ID),
			"status":     t.Status,
			"created_at": t.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetMerchant(w http.ResponseWriter, r *http.Request, tenantRepo repository.TenantRepository, log *slog.Logger, id string) {
	ctx := r.Context()
	tenant, err := tenantRepo.GetTenantByID(ctx, id)
	if err != nil {
		log.Error("get tenant failed", "error", err)
		http.Error(w, "failed to get tenant", http.StatusInternalServerError)
		return
	}
	if tenant == nil {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	// Mask the API key in individual lookups too.
	resp := map[string]any{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"platform":   tenant.Platform,
		"api_key":    maskAPIKey(tenant.APIKey),
		"status":     tenant.Status,
		"created_at": tenant.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
