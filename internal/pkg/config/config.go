package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the Aegis platform.
type Config struct {
	AegisGateway  AegisGatewayConfig  `yaml:"aegis_gateway"`
	MerchantMCP   MerchantMCPConfig   `yaml:"merchant_mcp"`
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

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `yaml:"level"`
}

// Load loads configuration from a YAML file with environment variable overrides.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if dsn := os.Getenv("DB_PASSWORD"); dsn != "" {

	}
	if keyID := os.Getenv("RAZORPAY_KEY_ID"); keyID != "" {
		cfg.Razorpay.KeyID = keyID
	}
	if keySecret := os.Getenv("RAZORPAY_KEY_SECRET"); keySecret != "" {
		cfg.Razorpay.KeySecret = keySecret
	}

	return &cfg, nil
}
