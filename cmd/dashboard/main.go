package main

import (
	"flag"
	"fmt"
	"html/template"
	stdlog "log"
	"net/http"
	"os"

	"github.com/razorpay/aegis/internal/app/repository"
	"github.com/razorpay/aegis/internal/app/service"
	"github.com/razorpay/aegis/internal/pkg/config"
	"github.com/razorpay/aegis/internal/pkg/logger"
)

var (
	port       = flag.Int("port", 8083, "Dashboard port")
	configPath = flag.String("config", "config.yaml", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level)
	stdLog := stdlog.New(os.Stdout, "", 0)

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

	auditRepo := repository.NewAuditPG(db)
	queueRepo := repository.NewApprovalQueuePG(db)

	auditService := service.NewAuditService(auditRepo)
	approvalQueueService := service.NewApprovalQueueService(queueRepo)

	dashboard := NewDashboardHandler(approvalQueueService, auditService, stdLog)

	mux := http.NewServeMux()
	mux.HandleFunc("/", dashboard.HandleIndex)
	mux.HandleFunc("/api/queue", dashboard.HandleQueueAPI)
	mux.HandleFunc("/api/queue/approve", dashboard.HandleApprove)
	mux.HandleFunc("/api/queue/reject", dashboard.HandleReject)
	mux.HandleFunc("/api/audit", dashboard.HandleAuditAPI)
	mux.HandleFunc("/api/health", dashboard.HandleHealth)

	addr := fmt.Sprintf(":%d", *port)
	log.Info("starting dashboard", "addr", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

// DashboardHandler handles HTTP requests for the dashboard
type DashboardHandler struct {
	approvalQueueService service.ApprovalQueueService
	auditService         service.AuditService
	log                  *stdlog.Logger
	templates            *template.Template
}

func NewDashboardHandler(
	approvalQueueService service.ApprovalQueueService,
	auditService service.AuditService,
	log *stdlog.Logger,
) *DashboardHandler {

	tmpl := template.Must(template.ParseFiles(
		"web/templates/index.html",
		"web/templates/queue_row.html",
		"web/templates/audit_row.html",
	))

	return &DashboardHandler{
		approvalQueueService: approvalQueueService,
		auditService:         auditService,
		log:                  log,
		templates:            tmpl,
	}
}

func (h *DashboardHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	h.templates.ExecuteTemplate(w, "index.html", nil)
}

func (h *DashboardHandler) HandleQueueAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pending, err := h.approvalQueueService.ListPending(ctx, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	for _, item := range pending {
		h.templates.ExecuteTemplate(w, "queue_row.html", item)
	}
}

func (h *DashboardHandler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	note := r.FormValue("note")
	reviewer := "dashboard_user"

	if err := h.approvalQueueService.Approve(r.Context(), id, reviewer, note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="alert alert-success">Approved: %s</div>`, id)
}

func (h *DashboardHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	note := r.FormValue("note")
	reviewer := "dashboard_user"

	if err := h.approvalQueueService.Reject(r.Context(), id, reviewer, note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="alert alert-error">Rejected: %s</div>`, id)
}

func (h *DashboardHandler) HandleAuditAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entries, err := h.auditService.GetAll(ctx, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	for _, entry := range entries {
		h.templates.ExecuteTemplate(w, "audit_row.html", entry)
	}
}

func (h *DashboardHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status": "ok"}`)
}
