package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
	"github.com/razorpay/aegis/internal/app/repository"
)

var (
	ErrOrderCreationFailed  = errors.New("order creation failed")
	ErrPaymentCaptureFailed = errors.New("payment capture failed")
	ErrApprovalNotFound     = errors.New("approval not found")
	ErrApprovalNotPending   = errors.New("approval not pending")
	ErrIdempotencyConflict  = errors.New("idempotency key conflict")
)

// GatewayServiceImpl implements the GatewayService interface.
type GatewayServiceImpl struct {
	policyEngine   PolicyEngine
	razorpayClient RazorpayMCPClient
	auditService   AuditService
	queueRepo      repository.ApprovalQueueRepository
	orderRepo      repository.OrderRepository
	catalogRepo    repository.CatalogRepository
	logger         *slog.Logger
}

// NewGatewayService creates a new GatewayServiceImpl.
func NewGatewayService(
	policyEngine PolicyEngine,
	razorpayClient RazorpayMCPClient,
	auditService AuditService,
	queueRepo repository.ApprovalQueueRepository,
	orderRepo repository.OrderRepository,
	catalogRepo repository.CatalogRepository,
	logger *slog.Logger,
) *GatewayServiceImpl {
	return &GatewayServiceImpl{
		policyEngine:   policyEngine,
		razorpayClient: razorpayClient,
		auditService:   auditService,
		queueRepo:      queueRepo,
		orderRepo:      orderRepo,
		catalogRepo:    catalogRepo,
		logger:         logger,
	}
}

// Purchase processes a purchase request through policy evaluation and payment execution.
func (s *GatewayServiceImpl) Purchase(ctx context.Context, params mcp.AegisPurchaseParams) (*mcp.AegisPurchaseResult, error) {

	req := &model.PurchaseRequest{
		BuyerID:        params.BuyerID,
		SessionID:      params.SessionID,
		ProductID:      params.ProductID,
		SKU:            params.SKU,
		Quantity:       params.Quantity,
		AmountPaisa:    params.AmountPaisa,
		IdempotencyKey: params.IdempotencyKey,
		BuyerPincode:   params.BuyerPincode,
		Metadata:       params.Metadata,
	}

	existingOrder, err := s.orderRepo.GetByIdempotencyKey(ctx, params.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("check idempotency: %w", err)
	}
	if existingOrder != nil {

		return &mcp.AegisPurchaseResult{
			Allowed:   true,
			Reason:    "duplicate request (idempotent)",
			RuleFired: model.RuleFiredNone,
			Status:    existingOrder.Status,
			OrderID:   existingOrder.ID,
			PaymentID: existingOrder.RazorpayPaymentID,
			Remaining: mcp.PolicyRemaining{},
		}, nil
	}

	if err := s.policyEngine.IncrementRequestCount(ctx, req.BuyerID, req.SessionID); err != nil {
		return nil, fmt.Errorf("increment request count: %w", err)
	}

	decision, err := s.policyEngine.Evaluate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("policy evaluation: %w", err)
	}

	auditEntry := s.buildAuditEntry(req, decision, model.AuditActionPurchaseAttempt, "")
	if err := s.auditService.Log(ctx, auditEntry); err != nil {

		s.logger.Error("audit log failed", "error", err)
	}

	if !decision.Allowed {

		approvalID, err := s.escalateToApprovalQueue(ctx, req, decision)
		if err != nil {
			return nil, fmt.Errorf("escalate to approval queue: %w", err)
		}

		auditEntry := s.buildAuditEntry(req, decision, model.AuditActionEscalated, approvalID)
		if err := s.auditService.Log(ctx, auditEntry); err != nil {
			s.logger.Error("audit log failed", "error", err)
		}

		return &mcp.AegisPurchaseResult{
			Allowed:         false,
			Reason:          decision.Reason,
			RuleFired:       decision.RuleFired,
			Status:          "BLOCKED",
			ApprovalQueueID: approvalID,
			Remaining:       convertPolicyRemaining(decision.Remaining),
		}, nil
	}

	orderID, paymentID, err := s.executePayment(ctx, req, decision)
	if err != nil {

		auditEntry := s.buildAuditEntry(req, decision, model.AuditActionPurchaseBlocked, err.Error())
		if auditErr := s.auditService.Log(ctx, auditEntry); auditErr != nil {
			s.logger.Error("audit log failed", "error", auditErr)
		}
		return nil, err
	}

	if err := s.policyEngine.RecordSpend(ctx, req.BuyerID, req.SessionID, req.SKU, req.Quantity, req.AmountPaisa); err != nil {
		return nil, fmt.Errorf("record spend: %w", err)
	}

	auditEntry = s.buildAuditEntry(req, decision, model.AuditActionPaymentExecuted, "")
	auditEntry.Response = json.RawMessage(fmt.Sprintf(`{"order_id": "%s", "payment_id": "%s"}`, orderID, paymentID))
	if err := s.auditService.Log(ctx, auditEntry); err != nil {
		s.logger.Error("audit log failed", "error", err)
	}

	return &mcp.AegisPurchaseResult{
		Allowed:   true,
		Reason:    decision.Reason,
		RuleFired: decision.RuleFired,
		Status:    "COMPLETED",
		OrderID:   orderID,
		PaymentID: paymentID,
		Remaining: convertPolicyRemaining(decision.Remaining),
	}, nil
}

// ApproveRequest approves a pending approval and executes the purchase.
func (s *GatewayServiceImpl) ApproveRequest(ctx context.Context, approvalID, reviewerID, note string) (*mcp.AegisPurchaseResult, error) {
	approval, err := s.queueRepo.GetByID(ctx, approvalID)
	if err != nil {
		return nil, fmt.Errorf("get approval: %w", err)
	}
	if approval == nil {
		return nil, ErrApprovalNotFound
	}
	if approval.Status != model.ApprovalStatusPending {
		return nil, ErrApprovalNotPending
	}

	if err := s.queueRepo.UpdateStatus(ctx, approvalID, model.ApprovalStatusApproved, reviewerID, note); err != nil {
		return nil, fmt.Errorf("update approval status: %w", err)
	}

	params := mcp.AegisPurchaseParams{
		BuyerID:        approval.PurchaseRequest.BuyerID,
		SessionID:      approval.PurchaseRequest.SessionID,
		ProductID:      approval.PurchaseRequest.ProductID,
		SKU:            approval.PurchaseRequest.SKU,
		Quantity:       approval.PurchaseRequest.Quantity,
		AmountPaisa:    approval.PurchaseRequest.AmountPaisa,
		IdempotencyKey: approval.PurchaseRequest.IdempotencyKey,
		BuyerPincode:   approval.PurchaseRequest.BuyerPincode,
		Metadata:       approval.PurchaseRequest.Metadata,
	}

	result, err := s.Purchase(ctx, params)
	if err != nil {
		return nil, err
	}

	auditEntry := &model.AuditEntry{
		Timestamp:      time.Now(),
		TraceID:        approvalID,
		BuyerID:        approval.BuyerID,
		SessionID:      approval.SessionID,
		Action:         model.AuditActionPurchaseApproved,
		PolicyDecision: &approval.PolicyDecision,
		Request:        json.RawMessage(fmt.Sprintf(`{"approval_id": "%s", "reviewer_id": "%s"}`, approvalID, reviewerID)),
		Response:       json.RawMessage(fmt.Sprintf(`{"order_id": "%s", "payment_id": "%s"}`, result.OrderID, result.PaymentID)),
	}
	if err := s.auditService.Log(ctx, auditEntry); err != nil {
		s.logger.Error("audit log failed", "error", err)
	}

	return result, nil
}

// RejectRequest rejects a pending approval.
func (s *GatewayServiceImpl) RejectRequest(ctx context.Context, approvalID, reviewerID, note string) error {
	approval, err := s.queueRepo.GetByID(ctx, approvalID)
	if err != nil {
		return fmt.Errorf("get approval: %w", err)
	}
	if approval == nil {
		return ErrApprovalNotFound
	}
	if approval.Status != model.ApprovalStatusPending {
		return ErrApprovalNotPending
	}

	if err := s.queueRepo.UpdateStatus(ctx, approvalID, model.ApprovalStatusRejected, reviewerID, note); err != nil {
		return fmt.Errorf("update approval status: %w", err)
	}

	auditEntry := &model.AuditEntry{
		Timestamp:      time.Now(),
		TraceID:        approvalID,
		BuyerID:        approval.BuyerID,
		SessionID:      approval.SessionID,
		Action:         model.AuditActionPurchaseRejected,
		PolicyDecision: &approval.PolicyDecision,
		Request:        json.RawMessage(fmt.Sprintf(`{"approval_id": "%s", "reviewer_id": "%s"}`, approvalID, reviewerID)),
		Response:       json.RawMessage(fmt.Sprintf(`{"note": "%s"}`, note)),
	}
	if err := s.auditService.Log(ctx, auditEntry); err != nil {
		s.logger.Error("audit log failed", "error", err)
	}

	return nil
}

// GetPendingApprovals returns all pending approvals.
func (s *GatewayServiceImpl) GetPendingApprovals(ctx context.Context) ([]*model.PendingApproval, error) {
	return s.queueRepo.ListPending(ctx, 100)
}

// executePayment creates order and captures payment via Razorpay MCP.
func (s *GatewayServiceImpl) executePayment(ctx context.Context, req *model.PurchaseRequest, decision *model.PolicyDecision) (string, string, error) {

	orderReq := CreateOrderRequest{
		AmountPaisa: req.AmountPaisa,
		Currency:    "INR",
		Receipt:     req.IdempotencyKey,
		Notes: map[string]string{
			"buyer_id":   req.BuyerID,
			"session_id": req.SessionID,
			"product_id": req.ProductID,
			"sku":        req.SKU,
			"quantity":   fmt.Sprintf("%d", req.Quantity),
		},
	}

	orderResp, err := s.razorpayClient.CreateOrder(ctx, orderReq)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrOrderCreationFailed, err)
	}

	captureResp, err := s.razorpayClient.CapturePayment(ctx, orderResp.OrderID)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrPaymentCaptureFailed, err)
	}

	order := &model.Order{
		ID:                orderResp.OrderID,
		BuyerID:           req.BuyerID,
		SessionID:         req.SessionID,
		ProductID:         req.ProductID,
		SKU:               req.SKU,
		Quantity:          req.Quantity,
		AmountPaisa:       req.AmountPaisa,
		Currency:          "INR",
		Status:            "PAID",
		RazorpayOrderID:   orderResp.OrderID,
		RazorpayPaymentID: captureResp.PaymentID,
		IdempotencyKey:    req.IdempotencyKey,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return "", "", fmt.Errorf("create order: %w", err)
	}

	return orderResp.OrderID, captureResp.PaymentID, nil
}

// escalateToApprovalQueue adds a blocked request to the approval queue.
func (s *GatewayServiceImpl) escalateToApprovalQueue(ctx context.Context, req *model.PurchaseRequest, decision *model.PolicyDecision) (string, error) {
	approval := &model.PendingApproval{
		ID:              generateApprovalID(),
		BuyerID:         req.BuyerID,
		SessionID:       req.SessionID,
		PurchaseRequest: *req,
		PolicyDecision:  *decision,
		BuyerReasoning:  extractBuyerReasoning(req.Metadata),
		Status:          model.ApprovalStatusPending,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}

	if err := s.queueRepo.Enqueue(ctx, approval); err != nil {
		return "", fmt.Errorf("enqueue approval: %w", err)
	}

	return approval.ID, nil
}

// buildAuditEntry creates an audit entry from request and decision.
func (s *GatewayServiceImpl) buildAuditEntry(req *model.PurchaseRequest, decision *model.PolicyDecision, action model.AuditAction, errorMsg string) *model.AuditEntry {
	requestJSON, _ := json.Marshal(req)
	responseJSON, _ := json.Marshal(decision)

	return &model.AuditEntry{
		Timestamp:      time.Now(),
		TraceID:        req.IdempotencyKey,
		BuyerID:        req.BuyerID,
		SessionID:      req.SessionID,
		Action:         action,
		PolicyDecision: decision,
		Request:        requestJSON,
		Response:       responseJSON,
		BuyerReasoning: extractBuyerReasoning(req.Metadata),
		Error:          errorMsg,
	}
}

// extractBuyerReasoning extracts reasoning from metadata for injection capture.
func extractBuyerReasoning(metadata json.RawMessage) string {
	if len(metadata) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(metadata, &m); err != nil {
		return ""
	}
	if reasoning, ok := m["reasoning"].(string); ok {
		return reasoning
	}
	return ""
}

// convertPolicyRemaining converts model.PolicyRemaining to mcp.PolicyRemaining.
func convertPolicyRemaining(r model.PolicyRemaining) mcp.PolicyRemaining {
	return mcp.PolicyRemaining{
		SpendCapPaisa:     r.SpendCapPaisa,
		PerSKUCap:         r.PerSKUCap,
		VelocityRemaining: r.VelocityRemaining,
	}
}

// generateApprovalID generates a unique approval ID.
func generateApprovalID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "appr_" + hex.EncodeToString(b)
}
