package http

import (
	"net/http"

	osapi "example.com/quotaapi/internal/openstack"
)

type QuotaGetServer struct {
	OS *osapi.Clients
}

// GET /quota/current?projectId=xxxx
func (s *QuotaGetServer) Current(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "GET only"})
		return
	}

	projectID := r.URL.Query().Get("projectId")
	if projectID == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing projectId"})
		return
	}

	nova, nErr := s.OS.GetNovaQuotaDetail(r.Context(), projectID)
	cind, cErr := s.OS.GetCinderQuotaDetail(r.Context(), projectID)
	if nErr != nil || cErr != nil {
		WriteJSON(w, http.StatusBadGateway, map[string]any{
			"error":  "quota read failed",
			"nova":   errString(nErr),
			"cinder": errString(cErr),
		})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"projectId": projectID,
		"nova":      nova,
		"cinder":    cind,
	})
}
