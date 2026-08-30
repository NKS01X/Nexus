package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	appmcp "github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
	pkgmcp "github.com/razorpay/aegis/internal/pkg/mcp"
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

	catalogRepo := repository.NewCatalogPG(db)
	orderRepo := repository.NewOrderPG(db)

	aegisClient := NewMCPAegisClient(fmt.Sprintf("http://%s:%d", cfg.AegisGateway.Host, cfg.AegisGateway.Port))

	merchantService := service.NewMerchantMCPService(catalogRepo, orderRepo, aegisClient)

	mcpServer := pkgmcp.NewMerchantServer(merchantService, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := fmt.Sprintf("%s:%d", cfg.MerchantMCP.Host, cfg.MerchantMCP.Port)
	log.Info("starting Merchant MCP Server", "addr", addr)

	if err := mcpServer.Start(ctx, addr); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	time.Sleep(100 * time.Millisecond)
}

// mcpAegisClient is an MCP client for calling Aegis Gateway via JSON-RPC 2.0.
type mcpAegisClient struct {
	baseURL string
	client  *http.Client
	reqID   int64
	mu      sync.Mutex
}

// NewMCPAegisClient creates a new MCP client for Aegis Gateway.
func NewMCPAegisClient(baseURL string) *mcpAegisClient {
	return &mcpAegisClient{
		baseURL: baseURL + "/mcp",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// nextID generates the next request ID.
func (c *mcpAegisClient) nextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reqID++
	return c.reqID
}

// initialize performs the MCP initialize handshake.
func (c *mcpAegisClient) initialize(ctx context.Context) error {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextID(),
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]string{
				"name":    "merchant-mcp",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}
	return nil
}

// doRequest sends a JSON-RPC request and returns the response.
func (c *mcpAegisClient) doRequest(ctx context.Context, req map[string]any) (*mcpResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	var resp mcpResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &resp, nil
}

// callTool calls an MCP tool by name with arguments.
func (c *mcpAegisClient) callTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      c.nextID(),
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call error: %s", resp.Error.Message)
	}

	// Extract result from MCP response format: {content: [{type: "text", text: "..."}]}
	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal tool result: %w", err)
	}

	if len(result.Content) == 0 || result.Content[0].Type != "text" {
		return nil, fmt.Errorf("unexpected tool result format")
	}

	return json.RawMessage(result.Content[0].Text), nil
}

// Purchase calls the Aegis Gateway purchase tool via MCP.
func (c *mcpAegisClient) Purchase(ctx context.Context, params appmcp.AegisPurchaseParams) (*appmcp.AegisPurchaseResult, error) {

	if err := c.initialize(ctx); err != nil {

		fmt.Printf("initialize warning: %v\n", err)
	}

	args, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal purchase params: %w", err)
	}

	var argsMap map[string]any
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}

	resultJSON, err := c.callTool(ctx, appmcp.AegisToolPurchase, argsMap)
	if err != nil {
		return nil, fmt.Errorf("call purchase tool: %w", err)
	}

	var result appmcp.AegisPurchaseResult
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("unmarshal purchase result: %w", err)
	}

	return &result, nil
}

// mcpResponse represents an MCP JSON-RPC response.
type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

// mcpError represents an MCP JSON-RPC error.
type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
