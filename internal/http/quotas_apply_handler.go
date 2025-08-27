package http

import (
	"encoding/json"
	"net/http"

	"example.com/quotaapi/internal/api"
	"example.com/quotaapi/internal/openstack"
)

type QuotaApplyServer struct {
	OS *openstack.Clients
}

func (s *QuotaApplyServer) QuotaApply(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req api.ApplyQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.ProjectID == "" {
		http.Error(w, "missing projectId", http.StatusBadRequest)
		return
	}

	// 1) Nova 적용 (존재하는 항목만)
	if req.Nova != nil {
		if req.Nova.Cores == nil && req.Nova.RAMMB == nil && req.Nova.Instances == nil {
			// 아무 필드도 없으면 스킵
		} else {
			if err := s.OS.ApplyNovaQuota(ctx, req.ProjectID, req.Nova.Cores, req.Nova.RAMMB, req.Nova.Instances); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		}
	}

	// 2) Cinder 적용
	if req.Cinder != nil {
		if req.Cinder.Volumes == nil && req.Cinder.Snapshots == nil && req.Cinder.Gigabytes == nil {
			// 아무것도 없으면 스킵
		} else {
			if err := s.OS.ApplyCinderQuota(ctx, req.ProjectID, req.Cinder.Volumes, req.Cinder.Snapshots, req.Cinder.Gigabytes); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		}
	}

	// 3) 적용 후 최신 상태를 응답으로 돌려주면 UX가 좋음
	nova, _ := s.OS.GetNovaQuotaDetail(ctx, req.ProjectID)
	cinder, _ := s.OS.GetCinderQuotaDetail(ctx, req.ProjectID)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"projectId": req.ProjectID,
		"nova":      nova,
		"cinder":    cinder,
		"status":    "applied",
	})
}
