package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
)

// TenantServiceInterface is defined in service package to avoid circular imports.

// MerchantServer is the MCP server for the Merchant storefront (AI buyer facing).
type MerchantServer struct {
	merchantService service.MerchantMCPService
	tenantService   service.TenantServiceInterface
	catalogRepo     repository.CatalogRepository
	logger          *slog.Logger
	mu              sync.Mutex
	tools           map[string]ToolHandler
}

// NewMerchantServer creates a new Merchant MCP server.
func NewMerchantServer(merchantService service.MerchantMCPService, tenantService service.TenantServiceInterface, catalogRepo repository.CatalogRepository, logger *slog.Logger) *MerchantServer {
	s := &MerchantServer{
		merchantService: merchantService,
		tenantService:   tenantService,
		catalogRepo:     catalogRepo,
		logger:          logger,
		tools:           make(map[string]ToolHandler),
	}
	s.registerTools()
	return s
}

// registerTools registers all Merchant MCP tools.
func (s *MerchantServer) registerTools() {
	s.tools[mcp.MerchantToolSearchProducts] = s.handleSearchProducts
	s.tools[mcp.MerchantToolGetProduct] = s.handleGetProduct
	s.tools[mcp.MerchantToolCheckAvailability] = s.handleCheckAvailability
	s.tools[mcp.MerchantToolPurchase] = s.handlePurchase
	s.tools[mcp.MerchantToolGetOrderStatus] = s.handleGetOrderStatus
}

// Start starts the MCP server on the given address.
func (s *MerchantServer) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/mcp/", s.handleTenantMCP) // note trailing slash — catches /mcp/{store_id}
	mux.HandleFunc("/health", s.handleHealth)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting Merchant MCP server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		s.logger.Info("received signal, shutting down", "signal", sig)
	case <-ctx.Done():
		s.logger.Info("context cancelled, shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// handleMCP handles MCP protocol requests.
func (s *MerchantServer) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, req.ID, -32700, "Parse error", nil)
		return
	}

	s.logger.Debug("received MCP request", "method", req.Method, "id", req.ID)

	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req)
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolCall(w, req)
	default:
		s.writeError(w, req.ID, -32601, "Method not found", nil)
	}
}

// handleTenantMCP handles MCP protocol requests for a specific tenant.
func (s *MerchantServer) handleTenantMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract store_id from path: /mcp/{store_id}
	storeID := strings.TrimPrefix(r.URL.Path, "/mcp/")
	if storeID == "" {
		http.Error(w, "missing store_id", http.StatusBadRequest)
		return
	}

	// Resolve tenant by API key from Authorization: Bearer <key>
	apiKey := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if apiKey == "" {
		http.Error(w, "missing authorization header", http.StatusUnauthorized)
		return
	}

	// Get tenant by API key
	tenant, err := s.tenantService.GetTenantByAPIKey(r.Context(), apiKey)
	if err != nil {
		s.logger.Error("tenant lookup failed", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if tenant == nil || tenant.ID != storeID {
		http.Error(w, "tenant not found", http.StatusNotFound)
		return
	}

	// Create a tenant-scoped service
	tenantService := service.NewMerchantMCPService(
		&tenantCatalogRepo{catalogRepo: s.catalogRepo, tenantID: tenant.ID},
		s.merchantService.(*service.MerchantMCPServiceImpl).GetOrderRepo(),
		s.merchantService.(*service.MerchantMCPServiceImpl).GetAegisClient(),
	)

	// Handle the request with tenant-scoped service
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, req.ID, -32700, "Parse error", nil)
		return
	}

	s.logger.Debug("received MCP request", "method", req.Method, "id", req.ID, "tenant", storeID)

	switch req.Method {
	case "initialize":
		s.handleInitialize(w, req)
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolCallForTenant(w, req, tenantService)
	default:
		s.writeError(w, req.ID, -32601, "Method not found", nil)
	}
}

// handleToolCallForTenant handles the tools/call request for a specific tenant.
func (s *MerchantServer) handleToolCallForTenant(w http.ResponseWriter, req MCPRequest, tenantService service.MerchantMCPService) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeError(w, req.ID, -32602, "Invalid params", nil)
		return
	}

	s.logger.Info("tool call", "tool", params.Name)

	_, ok := s.tools[params.Name]
	if !ok {
		s.writeError(w, req.ID, -32601, fmt.Sprintf("Tool not found: %s", params.Name), nil)
		return
	}

	// We need to create a tenant-aware handler wrapper
	result, err := s.executeToolForTenant(req.Context(), params.Arguments, params.Name, tenantService)
	if err != nil {
		s.logger.Error("tool call failed", "tool", params.Name, "error", err)
		s.writeError(w, req.ID, -32603, err.Error(), nil)
		return
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
	s.writeResponse(w, resp)
}

// executeToolForTenant executes a tool using the tenant-scoped service.
func (s *MerchantServer) executeToolForTenant(ctx context.Context, params json.RawMessage, toolName string, tenantService service.MerchantMCPService) (any, error) {
	switch toolName {
	case mcp.MerchantToolSearchProducts:
		var p mcp.SearchProductsParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		result, err := tenantService.SearchProducts(ctx, p)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": mustMarshalJSON(result),
				},
			},
		}, nil
	case mcp.MerchantToolGetProduct:
		var p mcp.GetProductParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		result, err := tenantService.GetProduct(ctx, p)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": mustMarshalJSON(result),
				},
			},
		}, nil
	case mcp.MerchantToolCheckAvailability:
		var p mcp.CheckAvailabilityParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		result, err := tenantService.CheckAvailability(ctx, p)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": mustMarshalJSON(result),
				},
			},
		}, nil
	case mcp.MerchantToolPurchase:
		var p mcp.PurchaseParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		result, err := tenantService.Purchase(ctx, p)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": mustMarshalJSON(result),
				},
			},
		}, nil
	case mcp.MerchantToolGetOrderStatus:
		var p mcp.GetOrderStatusParams
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		result, err := tenantService.GetOrderStatus(ctx, p)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": mustMarshalJSON(result),
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
}

// tenantCatalogRepo wraps CatalogRepository with tenant scoping.
type tenantCatalogRepo struct {
	catalogRepo repository.CatalogRepository
	tenantID    string
}

func (t *tenantCatalogRepo) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	return t.catalogRepo.GetProductByTenant(ctx, id, t.tenantID)
}

func (t *tenantCatalogRepo) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	// For simplicity, we don't implement full SKU lookup by tenant here
	// In production, you'd add GetProductBySKUByTenant to the interface
	return t.catalogRepo.GetProductBySKU(ctx, sku)
}

func (t *tenantCatalogRepo) SearchProducts(ctx context.Context, filter repository.SearchFilter) ([]*model.Product, error) {
	return t.catalogRepo.SearchProductsByTenant(ctx, filter, t.tenantID)
}

func (t *tenantCatalogRepo) GetAllProducts(ctx context.Context) ([]*model.Product, error) {
	return t.catalogRepo.SearchProductsByTenant(ctx, repository.SearchFilter{}, t.tenantID)
}

func (t *tenantCatalogRepo) CheckAvailability(ctx context.Context, sku string) (*model.InventoryCheck, error) {
	return t.catalogRepo.CheckAvailabilityByTenant(ctx, sku, t.tenantID)
}

func (t *tenantCatalogRepo) ReserveInventory(ctx context.Context, sku string, quantity int) error {
	// Use base method since inventory operations are tenant-scoped by SKU
	return t.catalogRepo.ReserveInventory(ctx, sku, quantity)
}

func (t *tenantCatalogRepo) ReleaseInventory(ctx context.Context, sku string, quantity int) error {
	return t.catalogRepo.ReleaseInventory(ctx, sku, quantity)
}

func (t *tenantCatalogRepo) ConfirmInventory(ctx context.Context, sku string, quantity int) error {
	return t.catalogRepo.ConfirmInventory(ctx, sku, quantity)
}

func (t *tenantCatalogRepo) InsertProduct(ctx context.Context, p *model.Product, tenantID string) error {
	return t.catalogRepo.InsertProduct(ctx, p, tenantID)
}

func (t *tenantCatalogRepo) SearchProductsByTenant(ctx context.Context, filter repository.SearchFilter, tenantID string) ([]*model.Product, error) {
	return t.catalogRepo.SearchProductsByTenant(ctx, filter, tenantID)
}

func (t *tenantCatalogRepo) GetProductByTenant(ctx context.Context, id string, tenantID string) (*model.Product, error) {
	return t.catalogRepo.GetProductByTenant(ctx, id, tenantID)
}

func (t *tenantCatalogRepo) CheckAvailabilityByTenant(ctx context.Context, sku string, tenantID string) (*model.InventoryCheck, error) {
	return t.catalogRepo.CheckAvailabilityByTenant(ctx, sku, tenantID)
}

// handleInitialize handles the initialize request.
func (s *MerchantServer) handleInitialize(w http.ResponseWriter, req MCPRequest) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]string{
				"name":    "merchant-storefront",
				"version": "1.0.0",
			},
		},
	}
	s.writeResponse(w, resp)
}

// handleToolsList handles the tools/list request.
func (s *MerchantServer) handleToolsList(w http.ResponseWriter, req MCPRequest) {
	tools := []map[string]any{
		{
			"name":        mcp.MerchantToolSearchProducts,
			"description": "Search products in the merchant catalog",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query":           map[string]string{"type": "string"},
					"category":        map[string]string{"type": "string"},
					"max_price_paisa": map[string]string{"type": "integer"},
					"min_price_paisa": map[string]string{"type": "integer"},
					"in_stock_only":   map[string]string{"type": "boolean"},
					"limit":           map[string]string{"type": "integer"},
					"color":           map[string]string{"type": "string"},
					"size":            map[string]string{"type": "string"},
					"brand":           map[string]string{"type": "string"},
				},
			},
		},
		{
			"name":        mcp.MerchantToolGetProduct,
			"description": "Get detailed product information by ID",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"product_id": map[string]string{"type": "string"},
				},
				"required": []string{"product_id"},
			},
		},
		{
			"name":        mcp.MerchantToolCheckAvailability,
			"description": "Check inventory availability for a SKU",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sku": map[string]string{"type": "string"},
				},
				"required": []string{"sku"},
			},
		},
		{
			"name":        mcp.MerchantToolPurchase,
			"description": "Initiate a purchase through Aegis Gateway",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"buyer_id":        map[string]string{"type": "string"},
					"session_id":      map[string]string{"type": "string"},
					"product_id":      map[string]string{"type": "string"},
					"sku":             map[string]string{"type": "string"},
					"quantity":        map[string]string{"type": "integer"},
					"idempotency_key": map[string]string{"type": "string"},
					"buyer_pincode":   map[string]string{"type": "string"},
				},
				"required": []string{"buyer_id", "session_id", "product_id", "sku", "quantity", "idempotency_key"},
			},
		},
		{
			"name":        mcp.MerchantToolGetOrderStatus,
			"description": "Get order status by order ID",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"order_id": map[string]string{"type": "string"},
				},
				"required": []string{"order_id"},
			},
		},
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"tools": tools},
	}
	s.writeResponse(w, resp)
}

// handleToolCall handles the tools/call request.
func (s *MerchantServer) handleToolCall(w http.ResponseWriter, req MCPRequest) {
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeError(w, req.ID, -32602, "Invalid params", nil)
		return
	}

	s.logger.Info("tool call", "tool", params.Name)

	handler, ok := s.tools[params.Name]
	if !ok {
		s.writeError(w, req.ID, -32601, fmt.Sprintf("Tool not found: %s", params.Name), nil)
		return
	}

	result, err := handler(req.Context(), params.Arguments)
	if err != nil {
		s.logger.Error("tool call failed", "tool", params.Name, "error", err)
		s.writeError(w, req.ID, -32603, err.Error(), nil)
		return
	}

	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
	s.writeResponse(w, resp)
}

// handleSearchProducts handles the search_products tool.
func (s *MerchantServer) handleSearchProducts(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.SearchProductsParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.merchantService.SearchProducts(ctx, p)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": mustMarshalJSON(result),
			},
		},
	}, nil
}

// handleGetProduct handles the get_product tool.
func (s *MerchantServer) handleGetProduct(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.GetProductParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.merchantService.GetProduct(ctx, p)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": mustMarshalJSON(result),
			},
		},
	}, nil
}

// handleCheckAvailability handles the check_availability tool.
func (s *MerchantServer) handleCheckAvailability(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.CheckAvailabilityParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.merchantService.CheckAvailability(ctx, p)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": mustMarshalJSON(result),
			},
		},
	}, nil
}

// handlePurchase handles the purchase tool.
func (s *MerchantServer) handlePurchase(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.PurchaseParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.merchantService.Purchase(ctx, p)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": mustMarshalJSON(result),
			},
		},
	}, nil
}

// handleGetOrderStatus handles the get_order_status tool.
func (s *MerchantServer) handleGetOrderStatus(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.GetOrderStatusParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.merchantService.GetOrderStatus(ctx, p)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": mustMarshalJSON(result),
			},
		},
	}, nil
}

// handleHealth handles health check.
func (s *MerchantServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// writeResponse writes a JSON response.
func (s *MerchantServer) writeResponse(w http.ResponseWriter, resp MCPResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func (s *MerchantServer) writeError(w http.ResponseWriter, id any, code int, message string, data any) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.writeResponse(w, resp)
}
