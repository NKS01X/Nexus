package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/razorpay/aegis/internal/app/service"
)

func handleListApprovals(w http.ResponseWriter, r *http.Request, gatewayService service.GatewayService) {
	approvals, err := gatewayService.GetPendingApprovals(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(approvals)
}

func handleApproveAction(w http.ResponseWriter, r *http.Request, gatewayService service.GatewayService) {
	var req struct {
		ID   string `json:"id"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	result, err := gatewayService.ApproveRequest(r.Context(), req.ID, "admin_portal", req.Note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleRejectAction(w http.ResponseWriter, r *http.Request, gatewayService service.GatewayService) {
	var req struct {
		ID   string `json:"id"`
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	err := gatewayService.RejectRequest(r.Context(), req.ID, "admin_portal", req.Note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleAuditVerify(w http.ResponseWriter, r *http.Request, auditService service.AuditService) {
	valid, err := auditService.VerifyIntegrity(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"chain_valid": valid,
	})
}

func handleAuditTrail(w http.ResponseWriter, r *http.Request, auditService service.AuditService) {
	buyerID := r.URL.Query().Get("buyer_id")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	entries, err := auditService.GetTrail(r.Context(), buyerID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func handleAuditEntries(w http.ResponseWriter, r *http.Request, auditService service.AuditService) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	entries, err := auditService.GetAll(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func handleRunRedTeam(w http.ResponseWriter, r *http.Request) {
	// Resolve binary and config paths relative to the executable so the handler
	// works regardless of the process working directory.
	exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		exeDir = "."
	}
	// Walk up from the binary dir to find the project root (contains config.yaml).
	projectRoot := exeDir
	if filepath.Base(exeDir) == "bin" {
		projectRoot = filepath.Dir(exeDir)
	}

	redteamBin := filepath.Join(projectRoot, "bin", "redteam")
	configPath := filepath.Join(projectRoot, "config.yaml")

	// Fall back to relative paths if the resolved binary doesn't exist.
	if _, statErr := os.Stat(redteamBin); statErr != nil {
		redteamBin = "./bin/redteam"
		configPath = "config.yaml"
	}

	cmd := exec.Command(redteamBin, "--json", configPath)
	cmd.Dir = projectRoot
	output, _ := cmd.CombinedOutput()

	// The redteam binary exits with code 1 even when all attacks are blocked
	// (it signals "found vulnerabilities" vs "clean"). Always try to return the
	// JSON payload — fall back to a structured error only if output isn't JSON.
	w.Header().Set("Content-Type", "application/json")
	if json.Valid(output) && len(output) > 0 {
		w.Write(output)
		return
	}

	// Output was not valid JSON — return a structured error the frontend can render.
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error":  fmt.Sprintf("redteam binary error: %s", string(output)),
		"detail": fmt.Sprintf("binary=%s config=%s", redteamBin, configPath),
	})
}
