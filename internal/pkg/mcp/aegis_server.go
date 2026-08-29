package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/service"
)

// AegisServer is the MCP server for the Aegis Gateway.
type AegisServer struct {
	gatewayService service.GatewayService
	logger         *slog.Logger
	mu             sync.Mutex
	tools          map[string]ToolHandler
}

// ToolHandler is a function that handles an MCP tool call.
type ToolHandler func(ctx context.Context, params json.RawMessage) (any, error)

// NewAegisServer creates a new Aegis MCP server.
func NewAegisServer(gatewayService service.GatewayService, logger *slog.Logger) *AegisServer {
	s := &AegisServer{
		gatewayService: gatewayService,
		logger:         logger,
		tools:          make(map[string]ToolHandler),
	}
	s.registerTools()
	return s
}

// registerTools registers all Aegis MCP tools.
func (s *AegisServer) registerTools() {
	s.tools[mcp.AegisToolPurchase] = s.handlePurchase
}

// Start starts the MCP server on the given address.
func (s *AegisServer) Start(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleMCP)
	mux.HandleFunc("/health", s.handleHealth)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting Aegis MCP server", "addr", addr)
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
func (s *AegisServer) handleMCP(w http.ResponseWriter, r *http.Request) {
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

// handleInitialize handles the initialize request.
func (s *AegisServer) handleInitialize(w http.ResponseWriter, req MCPRequest) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]string{
				"name":    "aegis-gateway",
				"version": "1.0.0",
			},
		},
	}
	s.writeResponse(w, resp)
}

// handleToolsList handles the tools/list request.
func (s *AegisServer) handleToolsList(w http.ResponseWriter, req MCPRequest) {
	tools := []map[string]any{
		{
			"name":        mcp.AegisToolPurchase,
			"description": "Process a purchase request through Aegis Gateway policy enforcement",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"buyer_id":        map[string]string{"type": "string"},
					"session_id":      map[string]string{"type": "string"},
					"product_id":      map[string]string{"type": "string"},
					"sku":             map[string]string{"type": "string"},
					"quantity":        map[string]string{"type": "integer"},
					"amount_paisa":    map[string]string{"type": "integer"},
					"idempotency_key": map[string]string{"type": "string"},
					"buyer_pincode":   map[string]string{"type": "string"},
					"metadata":        map[string]string{"type": "object"},
				},
				"required": []string{"buyer_id", "session_id", "product_id", "sku", "quantity", "amount_paisa", "idempotency_key"},
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
func (s *AegisServer) handleToolCall(w http.ResponseWriter, req MCPRequest) {
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

// handlePurchase handles the purchase tool.
func (s *AegisServer) handlePurchase(ctx context.Context, params json.RawMessage) (any, error) {
	var p mcp.AegisPurchaseParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	result, err := s.gatewayService.Purchase(ctx, p)
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
func (s *AegisServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// writeResponse writes a JSON response.
func (s *AegisServer) writeResponse(w http.ResponseWriter, resp MCPResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// writeError writes an error response.
func (s *AegisServer) writeError(w http.ResponseWriter, id any, code int, message string, data any) {
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

// MCPRequest represents an MCP JSON-RPC request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Context returns a context for the request.
func (r *MCPRequest) Context() context.Context {
	return context.Background()
}

// MCPResponse represents an MCP JSON-RPC response.
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

// MCPError represents an MCP JSON-RPC error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// mustMarshalJSON marshals to JSON, panicking on error.
func mustMarshalJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
