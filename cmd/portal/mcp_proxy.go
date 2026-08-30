package main

import (
    "fmt"
    "net/http"
    "net/http/httputil"
    "net/url"
)

// newMCPProxy builds a reverse‑proxy that forwards /mcp/* requests to the internal MCP server.
// It defaults to http://localhost:8082 if no MCP_INTERNAL_URL env var is supplied.
func newMCPProxy(internalURL string) (http.Handler, error) {
    if internalURL == "" {
        internalURL = "http://localhost:8082"
    }
    target, err := url.Parse(internalURL)
    if err != nil {
        return nil, fmt.Errorf("parse MCP_INTERNAL_URL: %w", err)
    }
    return httputil.NewSingleHostReverseProxy(target), nil
}
