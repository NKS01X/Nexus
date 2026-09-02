package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"syscall"
	"time"

	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
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

	log := logger.New(cfg.Log.Level)
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

    // AI Completion endpoint – Groq LLM
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
