package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

const baseYAML = `
aegis_gateway:
  host: "localhost"
  port: 8081
merchant_mcp:
  host: "localhost"
  port: 8082
portal:
  host: "localhost"
  port: 8084
database:
  dsn: "postgres://yaml-dsn"
log:
  level: "info"
`

func TestLoadEnvOverrides(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		wantDSN       string
		wantBindHost  string
		wantPortaPort int
		wantMCPURL    string
	}{
		{
			name:          "No Env Uses YAML Values",
			env:           map[string]string{},
			wantDSN:       "postgres://yaml-dsn",
			wantBindHost:  "localhost",
			wantPortaPort: 8084,
			wantMCPURL:    "",
		},
		{
			name: "Platform Injected Values Win",
			env: map[string]string{
				"DATABASE_URL":        "postgres://render-internal/neondb",
				"HOST":                "0.0.0.0",
				"PORT":                "10000",
				"MCP_PUBLIC_BASE_URL": "https://nexus-api.onrender.com",
			},
			wantDSN:       "postgres://render-internal/neondb",
			wantBindHost:  "0.0.0.0",
			wantPortaPort: 10000,
			wantMCPURL:    "https://nexus-api.onrender.com",
		},
		{
			name: "Invalid Port Ignored Keeps YAML",
			env: map[string]string{
				"PORT": "not-a-number",
			},
			wantDSN:       "postgres://yaml-dsn",
			wantBindHost:  "localhost",
			wantPortaPort: 8084,
			wantMCPURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"DATABASE_URL", "HOST", "PORT", "MCP_PUBLIC_BASE_URL"} {
				os.Unsetenv(k)
			}
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			}()

			cfg, err := Load(writeTempConfig(t, baseYAML))
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Database.DSN != tt.wantDSN {
				t.Errorf("DSN = %q, want %q", cfg.Database.DSN, tt.wantDSN)
			}
			if cfg.Portal.Host != tt.wantBindHost {
				t.Errorf("Portal.Host = %q, want %q", cfg.Portal.Host, tt.wantBindHost)
			}
			if cfg.MerchantMCP.Host != tt.wantBindHost {
				t.Errorf("MerchantMCP.Host = %q, want %q", cfg.MerchantMCP.Host, tt.wantBindHost)
			}
			if cfg.AegisGateway.Host != tt.wantBindHost {
				t.Errorf("AegisGateway.Host = %q, want %q", cfg.AegisGateway.Host, tt.wantBindHost)
			}
			if cfg.Portal.Port != tt.wantPortaPort {
				t.Errorf("Portal.Port = %d, want %d", cfg.Portal.Port, tt.wantPortaPort)
			}
			if cfg.MerchantMCP.PublicBaseURL != tt.wantMCPURL {
				t.Errorf("PublicBaseURL = %q, want %q", cfg.MerchantMCP.PublicBaseURL, tt.wantMCPURL)
			}
		})
	}
}
