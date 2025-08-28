package http

import (
	"context"
	"net/http"
	"time"

	"example.com/quotaapi/internal/services"
)

// ReconciliationHandler handles reconciliation-related HTTP requests
type ReconciliationHandler struct {
	reconciliationService *services.QuotaReconciliationService
}

// NewReconciliationHandler creates a new reconciliation handler
func NewReconciliationHandler(reconciliationService *services.QuotaReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{
		reconciliationService: reconciliationService,
	}
}

// ServeHTTP handles reconciliation requests
func (h *ReconciliationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "POST" && r.URL.Path == "/reconciliation/bulk":
		h.runBulkReconciliation(w, r)
	case r.Method == "GET" && r.URL.Path == "/reconciliation/status":
		h.getReconciliationStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

// runBulkReconciliation runs bulk quota reconciliation
func (h *ReconciliationHandler) runBulkReconciliation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute) // 5분 타임아웃
	defer cancel()

	result, err := h.reconciliationService.RunBulkReconciliation(ctx)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to run bulk reconciliation: " + err.Error(),
		})
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

// getReconciliationStatus gets the current reconciliation status
func (h *ReconciliationHandler) getReconciliationStatus(w http.ResponseWriter, r *http.Request) {
	// 간단한 상태 정보 반환
	WriteJSON(w, http.StatusOK, map[string]any{
		"status":    "ready",
		"message":   "Reconciliation service is ready",
		"timestamp": time.Now().UTC(),
	})
}
