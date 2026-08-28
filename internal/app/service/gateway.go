package service

import (
	"context"

	"github.com/razorpay/aegis/internal/app/mcp"
	"github.com/razorpay/aegis/internal/app/model"
)

// GatewayService defines the interface for the Aegis Gateway orchestration.
type GatewayService interface {
	Purchase(ctx context.Context, params mcp.AegisPurchaseParams) (*mcp.AegisPurchaseResult, error)
	ApproveRequest(ctx context.Context, approvalID, reviewerID, note string) (*mcp.AegisPurchaseResult, error)
	RejectRequest(ctx context.Context, approvalID, reviewerID, note string) error
	GetPendingApprovals(ctx context.Context) ([]*model.PendingApproval, error)
}
