package mcp

import (
    "net/http"

	internalService "github.com/razorpay/aegis/internal/app/service"
	internalMcp "github.com/razorpay/aegis/internal/pkg/mcp"
)

// Exported MCP request/response types via type aliases.
// This keeps the public API stable while the implementation lives in internal packages.

type (
    SearchProductsParams     = internalMcp.SearchProductsParams
    GetProductParams        = internalMcp.GetProductParams
    CheckAvailabilityParams = internalMcp.CheckAvailabilityParams
    PurchaseParams          = internalMcp.PurchaseParams
    PurchaseResult          = internalMcp.PurchaseResult
    // AI prompt structs (if needed by demo)
    AIPromptParams          = internalMcp.AIPromptParams
    AIPromptResult          = internalMcp.AIPromptResult
)

// NewServer builds the HTTP handler that serves the MCP endpoints.
// The caller provides a fully‑implemented MerchantMCPService.
func NewServer(svc internalService.MerchantMCPService) http.Handler {
    return internalMcp.NewMerchantServer(svc)
}

