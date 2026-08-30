# MCP Library

The MCP library (`mcp-lib`) provides a reusable Model Context Protocol (MCP) server implementation for merchants.

## How to use

1. Import the façade:
   ```
   import mcp "github.com/razorpay/aegis/internal/pkg/mcp"
   ```
2. Create a merchant MCP service and server:
   ```
   merchantMCPService := service.NewMerchantMCPService(catalogRepo, orderRepo, gatewayService)
   handler := mcp.NewMerchantServer(merchantMCPService, tenantService, catalogRepo, logger)
   ```
3. Start the server in a goroutine (for example in your main function):
   ```
   go func() {
       if err := handler.Start(context.Background(), "localhost:8082"); err != nil && err != http.ErrServerClosed {
           log.Error("MCP server error", "error", err)
       }
   }()
   ```
   The server implements `http.Handler` and can be proxied with a reverse‑proxy as shown in `cmd/portal/main.go`.

4. Run tests to verify:
   ```
   go test ./mcp-lib/...
   ```

## License

MIT
