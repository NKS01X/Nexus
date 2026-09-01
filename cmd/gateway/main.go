package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
	"github.com/razorpay/aegis/internal/pkg/mcp"
	"github.com/razorpay/aegis/internal/pkg/razorpay_mcp"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level)
	slog.SetDefault(log)

	db, err := repository.NewDB(cfg.Database.DSN)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Error("run migrations", "error", err)
		os.Exit(1)
	}

	policyRepo := repository.NewPolicyPG(db)
	catalogRepo := repository.NewCatalogPG(db)
	orderRepo := repository.NewOrderPG(db)
	auditRepo := repository.NewAuditPG(db)
	queueRepo := repository.NewApprovalQueuePG(db)

	razorpayClient, err := razorpay_mcp.NewClient(cfg.RazorpayMCP.BinaryPath, map[string]string{
		"RAZORPAY_KEY_ID":     cfg.Razorpay.KeyID,
		"RAZORPAY_KEY_SECRET": cfg.Razorpay.KeySecret,
	})
	if err != nil {
		log.Error("create razorpay mcp client", "error", err)
		os.Exit(1)
	}
	defer razorpayClient.Close()

	policyEngine := service.NewPolicyEngine(policyRepo, catalogRepo)
	auditService := service.NewAuditService(auditRepo)
	_ = service.NewApprovalQueueService(queueRepo)
	gatewayService := service.NewGatewayService(
		policyEngine,
		razorpayClient,
		auditService,
		queueRepo,
		orderRepo,
		catalogRepo,
		log,
	)

	mcpServer := mcp.NewAegisServer(gatewayService, auditService, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr := fmt.Sprintf("%s:%d", cfg.AegisGateway.Host, cfg.AegisGateway.Port)
	log.Info("starting Aegis Gateway", "addr", addr)

	if err := mcpServer.Start(ctx, addr); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}

	time.Sleep(100 * time.Millisecond)
}
