package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"example.com/quotaapi/internal/models"
	"example.com/quotaapi/internal/openstack"
)

// ApplyProfileReq represents the request to apply a quota profile
type ApplyProfileReq struct {
	ProjectID   string `json:"projectId"`
	Profile     string `json:"profile"`
	DryRun      bool   `json:"dryRun"`
	IncludeDiff bool   `json:"includeDiff"`
}

// ApplyProfileServer handles profile-based quota application
type ApplyProfileServer struct {
	OS *openstack.Clients
}

// NewApplyProfileHandler creates a new ApplyProfileHandler
func NewApplyProfileHandler(os *openstack.Clients) http.HandlerFunc {
	s := &ApplyProfileServer{OS: os}
	return s.Handle
}

// Handle processes the profile application request
func (s *ApplyProfileServer) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}

	var req ApplyProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json: " + err.Error()})
		return
	}

	if strings.TrimSpace(req.ProjectID) == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "projectId is required"})
		return
	}

	// models.Profiles에서 프로파일 조회
	profile, ok := models.Profiles[strings.ToLower(req.Profile)]
	if !ok {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown profile (use 'basic' or 'lab')"})
		return
	}

	ctx := r.Context()

	// 현재 쿼터 조회
	novaCurr, nErr := s.OS.GetNovaQuotaDetail(ctx, req.ProjectID)
	cinderCurr, cErr := s.OS.GetCinderQuotaDetail(ctx, req.ProjectID)
	if nErr != nil || cErr != nil {
		WriteJSON(w, http.StatusBadGateway, map[string]any{
			"error":  "failed to read current quotas",
			"nova":   errString(nErr),
			"cinder": errString(cErr),
		})
		return
	}

	// 기본 응답(계획/현재 상태)
	resp := map[string]any{
		"projectId": req.ProjectID,
		"profile":   strings.ToLower(req.Profile),
		"plan":      profile,
		"applied":   false,
		"dryRun":    req.DryRun,
		"current": map[string]any{
			"nova": map[string]any{
				"cores":     novaCurr.Cores,
				"ramMB":     novaCurr.RAMMB,
				"instances": novaCurr.Instances,
			},
			"cinder": map[string]any{
				"gigabytes": cinderCurr.Gigabytes,
				"volumes":   cinderCurr.Volumes,
				"snapshots": cinderCurr.Snapshots,
			},
		},
	}

	// diff 포함 요청 시, 계획(limit)과 현재(limit)의 차이를 계산
	if req.IncludeDiff {
		resp["diff"] = map[string]any{
			"nova": map[string]int{
				"cores":     profile.Cores - novaCurr.Cores.Limit,
				"ramMB":     profile.RAMMB - novaCurr.RAMMB.Limit,
				"instances": profile.Instances - novaCurr.Instances.Limit,
			},
			"cinder": map[string]int{
				"gigabytes": profile.Gigabytes - cinderCurr.Gigabytes.Limit,
				"volumes":   profile.Volumes - cinderCurr.Volumes.Limit,
				"snapshots": profile.Snapshots - cinderCurr.Snapshots.Limit,
			},
		}
	}

	// 실제 적용(dryRun=false)
	if !req.DryRun {
		// Nova 쿼터 적용
		cores, ram, inst := profile.Cores, profile.RAMMB, profile.Instances
		if err := s.OS.ApplyNovaQuota(ctx, req.ProjectID, &cores, &ram, &inst); err != nil {
			WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "nova apply failed: " + err.Error()})
			return
		}

		// Cinder 쿼터 적용
		vols, snaps, gigs := profile.Volumes, profile.Snapshots, profile.Gigabytes
		if err := s.OS.ApplyCinderQuota(ctx, req.ProjectID, &vols, &snaps, &gigs); err != nil {
			WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "cinder apply failed: " + err.Error()})
			return
		}

		resp["applied"] = true

		// 적용 후 최신 상태 재조회
		novaCurr, _ = s.OS.GetNovaQuotaDetail(ctx, req.ProjectID)
		cinderCurr, _ = s.OS.GetCinderQuotaDetail(ctx, req.ProjectID)
		resp["current"] = map[string]any{
			"nova": map[string]any{
				"cores":     novaCurr.Cores,
				"ramMB":     novaCurr.RAMMB,
				"instances": novaCurr.Instances,
			},
			"cinder": map[string]any{
				"gigabytes": cinderCurr.Gigabytes,
				"volumes":   cinderCurr.Volumes,
				"snapshots": cinderCurr.Snapshots,
			},
		}
	}

	WriteJSON(w, http.StatusOK, resp)
}
