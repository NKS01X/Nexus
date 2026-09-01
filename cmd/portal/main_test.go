package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/razorpay/aegis/internal/pkg/config"
)

func TestMCPBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		publicURL   string
		host        string
		port        int
		expectedURL string
	}{
		{"Public Base URL Takes Precedence", "https://mcp.iamnikhil.dev", "localhost", 8082, "https://mcp.iamnikhil.dev"},
		{"Trailing Slash Trimmed", "https://mcp.iamnikhil.dev/", "localhost", 8082, "https://mcp.iamnikhil.dev"},
		{"Falls Back To Internal Host Port", "", "localhost", 8082, "http://localhost:8082"},
		{"Fallback Custom Port", "", "0.0.0.0", 9099, "http://0.0.0.0:9099"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.MerchantMCP.PublicBaseURL = tt.publicURL
			cfg.MerchantMCP.Host = tt.host
			cfg.MerchantMCP.Port = tt.port

			got := mcpBaseURL(cfg)
			if got != tt.expectedURL {
				t.Errorf("mcpBaseURL() = %q, want %q", got, tt.expectedURL)
			}
		})
	}
}

func TestNewMCPProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"path": r.URL.Path, "method": r.Method})
	}))
	defer backend.Close()

	t.Run("Forwards Path And Method To Backend", func(t *testing.T) {
		handler, err := newMCPProxy(backend.URL)
		if err != nil {
			t.Fatalf("newMCPProxy() error = %v", err)
		}
		srv := httptest.NewServer(handler)
		defer srv.Close()

		res, err := http.Post(srv.URL+"/mcp/store_123", "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatalf("request error = %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", res.StatusCode, http.StatusOK)
		}
		var body map[string]string
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["path"] != "/mcp/store_123" {
			t.Errorf("proxied path = %q, want /mcp/store_123", body["path"])
		}
		if body["method"] != http.MethodPost {
			t.Errorf("proxied method = %q, want POST", body["method"])
		}
	})

	t.Run("Empty URL Defaults Without Error", func(t *testing.T) {
		if _, err := newMCPProxy(""); err != nil {
			t.Errorf("default target should not error, got %v", err)
		}
	})

	t.Run("Invalid URL Returns Error", func(t *testing.T) {
		if _, err := newMCPProxy("http://192.168.0.%%:8082"); err == nil {
			t.Error("expected parse error for invalid URL")
		}
	})
}
