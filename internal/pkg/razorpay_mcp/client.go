package razorpay_mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/razorpay/aegis/internal/app/service"
)

// Client implements service.RazorpayMCPClient using the Razorpay MCP server binary.
type Client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	mu      sync.Mutex
	reqID   int
	pending map[int]chan *mcpResponse
}

// mcpRequest represents a JSON-RPC request to the MCP server.
type mcpRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// mcpResponse represents a JSON-RPC response from the MCP server.
type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewClient creates a new Razorpay MCP client.
func NewClient(binaryPath string, env map[string]string) (*Client, error) {
	cmd := exec.Command(binaryPath, "stdio")
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start mcp server: %w", err)
	}

	client := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		pending: make(map[int]chan *mcpResponse),
	}

	go client.readResponses()

	if err := client.initialize(context.Background()); err != nil {
		client.Close()
		return nil, fmt.Errorf("initialize mcp: %w", err)
	}

	return client, nil
}

// initialize sends the initialize request to the MCP server.
func (c *Client) initialize(ctx context.Context) error {
	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]string{
				"name":    "aegis-gateway",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize failed: %s", resp.Error.Message)
	}

	notifyReq := mcpRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	return c.sendNotification(notifyReq)
}

// sendRequest sends a request and waits for response.
func (c *Client) sendRequest(ctx context.Context, req mcpRequest) (*mcpResponse, error) {
	c.mu.Lock()
	respChan := make(chan *mcpResponse, 1)
	c.pending[req.ID] = respChan
	c.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case resp, ok := <-respChan:
		if !ok || resp == nil {
			return nil, fmt.Errorf("mcp server closed response stream")
		}
		return resp, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, req.ID)
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.pending, req.ID)
		c.mu.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

// sendNotification sends a notification (no response expected).
func (c *Client) sendNotification(req mcpRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write notification: %w", err)
	}
	return nil
}

// readResponses continuously reads responses from stdout.
func (c *Client) readResponses() {
	decoder := json.NewDecoder(c.stdout)
	for {
		var resp mcpResponse
		if err := decoder.Decode(&resp); err != nil {

			c.mu.Lock()
			for _, ch := range c.pending {
				close(ch)
			}
			c.pending = make(map[int]chan *mcpResponse)
			c.mu.Unlock()
			return
		}

		c.mu.Lock()
		if ch, ok := c.pending[resp.ID]; ok {
			delete(c.pending, resp.ID)
			ch <- &resp
		}
		c.mu.Unlock()
	}
}

// nextID returns the next request ID.
func (c *Client) nextID() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reqID++
	return c.reqID
}

// callTool calls a tool on the MCP server.
func (c *Client) callTool(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      c.nextID(),
		Method:  "tools/call",
		Params: map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call failed: %s", resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal tool result: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty tool result")
	}

	return json.RawMessage(result.Content[0].Text), nil
}

// CreateOrder creates a Razorpay order.
func (c *Client) CreateOrder(ctx context.Context, req service.CreateOrderRequest) (*service.CreateOrderResponse, error) {
	args := map[string]any{
		"amount":   req.AmountPaisa,
		"currency": req.Currency,
		"receipt":  req.Receipt,
		"notes":    req.Notes,
	}

	result, err := c.callTool(ctx, "create_order", args)
	if err != nil {
		return nil, err
	}

	var resp service.CreateOrderResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create_order response: %w", err)
	}

	return &resp, nil
}

// CapturePayment captures a payment for an order.
func (c *Client) CapturePayment(ctx context.Context, paymentID string) (*service.CapturePaymentResponse, error) {
	args := map[string]any{
		"payment_id": paymentID,
	}

	result, err := c.callTool(ctx, "capture_payment", args)
	if err != nil {
		return nil, err
	}

	var resp service.CapturePaymentResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal capture_payment response: %w", err)
	}

	return &resp, nil
}

// GetPayment retrieves payment details.
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*service.PaymentResponse, error) {
	args := map[string]any{
		"payment_id": paymentID,
	}

	result, err := c.callTool(ctx, "get_payment", args)
	if err != nil {
		return nil, err
	}

	var resp service.PaymentResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal get_payment response: %w", err)
	}

	return &resp, nil
}

// CreateRefund creates a refund for a payment.
func (c *Client) CreateRefund(ctx context.Context, req service.CreateRefundRequest) (*service.CreateRefundResponse, error) {
	args := map[string]any{
		"payment_id":      req.PaymentID,
		"amount":          req.AmountPaisa,
		"currency":        req.Currency,
		"idempotency_key": req.IdempotencyKey,
	}

	result, err := c.callTool(ctx, "create_refund", args)
	if err != nil {
		return nil, err
	}

	var resp service.CreateRefundResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal create_refund response: %w", err)
	}

	return &resp, nil
}

// Close closes the MCP client connection.
func (c *Client) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()
	return c.cmd.Wait()
}

// IncrementRequestCount is a no-op for the real client (velocity tracking handled by policy engine).
func (c *Client) IncrementRequestCount(ctx context.Context, buyerID, sessionID string) error {
	return nil
}
