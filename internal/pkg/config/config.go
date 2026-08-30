package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the Aegis platform.
type Config struct {
	Groq GroqConfig `yaml:"groq"`
	AegisGateway  AegisGatewayConfig  `yaml:"aegis_gateway"`
	MerchantMCP   MerchantMCPConfig   `yaml:"merchant_mcp"`
	Portal        PortalConfig        `yaml:"portal"`
	Database      DatabaseConfig      `yaml:"database"`
	Razorpay      RazorpayConfig      `yaml:"razorpay"`
	RazorpayMCP   RazorpayMCPConfig   `yaml:"razorpay_mcp"`
	Policy        PolicyConfig        `yaml:"policy"`
	Audit         AuditConfig         `yaml:"audit"`
	ApprovalQueue ApprovalQueueConfig `yaml:"approval_queue"`
	Log           LogConfig           `yaml:"log"`
}

// AegisGatewayConfig holds Aegis Gateway server configuration.
type AegisGatewayConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// MerchantMCPConfig holds Merchant MCP server configuration.
type MerchantMCPConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	// PublicBaseURL is the externally reachable base URL for the MCP server
	// (e.g. https://mcp.example.com). When empty, http://<host>:<port> is used.
	PublicBaseURL string `yaml:"public_base_url"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

// RazorpayConfig holds Razorpay credentials.
type RazorpayConfig struct {
	KeyID     string `yaml:"key_id"`
	KeySecret string `yaml:"key_secret"`
}

// RazorpayMCPConfig holds Razorpay MCP client configuration.
type RazorpayMCPConfig struct {
	BinaryPath string `yaml:"binary_path"`
}

// PolicyConfig holds policy configuration.
type PolicyConfig struct {
	SpendCapPaisa     int64          `yaml:"spend_cap_paisa"`
	PerSKUCap         map[string]int `yaml:"per_sku_cap"`
	VelocityCap       VelocityLimit  `yaml:"velocity_cap"`
	AllowedCategories []string       `yaml:"allowed_categories"`
	BlockedSKUs       []string       `yaml:"blocked_skus"`
	GeoRules          []GeoRule      `yaml:"geo_rules"`
}

// VelocityLimit defines request rate limiting configuration.
type VelocityLimit struct {
	MaxRequests   int           `yaml:"max_requests"`
	Window        time.Duration `yaml:"-"`
	WindowSeconds int           `yaml:"window_seconds"`
}

// GetWindow returns the velocity window as a time.Duration.
func (v *VelocityLimit) GetWindow() time.Duration {
	if v.Window == 0 && v.WindowSeconds > 0 {
		v.Window = time.Duration(v.WindowSeconds) * time.Second
	}
	return v.Window
}

// GeoRule defines geographic restrictions for purchases.
type GeoRule struct {
	Country  string   `yaml:"country"`
	Allowed  bool     `yaml:"allowed"`
	Pincodes []string `yaml:"pincodes,omitempty"`
}

// AuditConfig holds audit log configuration.
type AuditConfig struct {
	RetentionDays int `yaml:"retention_days"`
}

// ApprovalQueueConfig holds approval queue configuration.
type ApprovalQueueConfig struct {
	DefaultTTLHours int `yaml:"default_ttl_hours"`
}

// PortalConfig holds Portal server configuration.
type PortalConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	AdminKey string `yaml:"admin_key"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
    Level string `yaml:"level"`
}

// GroqConfig holds Groq LLM client configuration.
type GroqConfig struct {
    APIKey string `yaml:"api_key"`
    Model  string `yaml:"model"`
    }


// Load loads configuration from a YAML file with environment variable overrides.
// Environment variables take precedence, enabling portable deployments where the
// platform (Render, Heroku, containers) injects binding and credential values.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		cfg.Database.DSN = dsn
	}
	if host := os.Getenv("HOST"); host != "" {
		cfg.AegisGateway.Host = host
		cfg.MerchantMCP.Host = host
		cfg.Portal.Host = host
	}
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Portal.Port = p
		}
	}
if publicURL := os.Getenv("MCP_PUBLIC_BASE_URL"); publicURL != "" {
        cfg.MerchantMCP.PublicBaseURL = publicURL
    }
    // Groq configuration – read from env if set.
    if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
        cfg.Groq.APIKey = apiKey
    }
    if model := os.Getenv("GROQ_MODEL"); model != "" {
        cfg.Groq.Model = model
    }
	if keyID := os.Getenv("RAZORPAY_KEY_ID"); keyID != "" {
		cfg.Razorpay.KeyID = keyID
	}
	if keySecret := os.Getenv("RAZORPAY_KEY_SECRET"); keySecret != "" {
		cfg.Razorpay.KeySecret = keySecret
	}
	if mcpBin := os.Getenv("RAZORPAY_MCP_BINARY_PATH"); mcpBin != "" {
		cfg.RazorpayMCP.BinaryPath = mcpBin
	}
	if adminKey := os.Getenv("NEXUS_ADMIN_KEY"); adminKey != "" {
		cfg.Portal.AdminKey = adminKey
	}
	if cfg.Portal.AdminKey == "" {
		cfg.Portal.AdminKey = "nexus_admin_default"
	}

	return &cfg, nil
}
