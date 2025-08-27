package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"example.com/quotaapi/internal/openstack"
)

// plan: 프로파일이 지정하는 목표 쿼터 값 집합
// - Nova: Cores, RAMMB, Instances
// - Cinder: Gigabytes, Volumes, Snapshots
// 실제 적용은 요청(dryRun=false)일 때 수행됨
type plan struct {
	Cores     int `json:"cores"`
	RAMMB     int `json:"ramMB"`
	Instances int `json:"instances"`
	Gigabytes int `json:"gigabytes"`
	Volumes   int `json:"volumes"`
	Snapshots int `json:"snapshots"`
}

// profiles: 사전에 정의된 프로파일 목록
// - 사용 예: profile=basic | lab
// - 필요한 경우 향후 설정 파일/DB로 분리 가능
var profiles = map[string]plan{
	"basic": {Cores: 8, RAMMB: 16384, Instances: 10, Gigabytes: 100, Volumes: 10, Snapshots: 10},
	"lab":   {Cores: 16, RAMMB: 32768, Instances: 20, Gigabytes: 200, Volumes: 20, Snapshots: 20},
}

// applyProfileReq: 프로파일 기반 쿼터 적용 요청 바디
// - dryRun=true  → 적용하지 않고 계획/차이만 반환
// - includeDiff → 현재 쿼터와의 차이값(diff) 포함 여부
type applyProfileReq struct {
	ProjectID   string `json:"projectId"`
	Profile     string `json:"profile"`
	DryRun      bool   `json:"dryRun"`
	IncludeDiff bool   `json:"includeDiff"`
}

// ApplyProfileServer: OpenStack 클라이언트를 주입받아 동작하는 핸들러 서버
type ApplyProfileServer struct {
	OS *openstack.Clients
}

// NewApplyProfileHandler: POST /quota/applyProfile 핸들러 생성
// - 요청을 검증하고 프로파일을 확인한 뒤
// - 현재 쿼터 조회, 필요 시 diff 계산, dryRun=false면 실제 적용
func NewApplyProfileHandler(os *openstack.Clients) http.HandlerFunc {
	s := &ApplyProfileServer{OS: os}
	return s.Handle
}

// Handle: /quota/applyProfile 처리 엔드포인트
// 요청 플로우
// 1) 메서드/바디/필수 필드(projectId, profile) 검증
// 2) 프로파일 조회
// 3) 현재 쿼터 조회(Nova/Cinder)
// 4) includeDiff=true면 계획과 현재의 차이 계산
// 5) dryRun=false면 계획대로 Nova/Cinder 쿼터 적용 후 최신 상태 재조회
func (s *ApplyProfileServer) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}

	var req applyProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json: " + err.Error()})
		return
	}

	if strings.TrimSpace(req.ProjectID) == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "projectId is required"})
		return
	}

	p, ok := profiles[strings.ToLower(req.Profile)]
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
		"plan":      p,
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
				"cores":     p.Cores - novaCurr.Cores.Limit,
				"ramMB":     p.RAMMB - novaCurr.RAMMB.Limit,
				"instances": p.Instances - novaCurr.Instances.Limit,
			},
			"cinder": map[string]int{
				"gigabytes": p.Gigabytes - cinderCurr.Gigabytes.Limit,
				"volumes":   p.Volumes - cinderCurr.Volumes.Limit,
				"snapshots": p.Snapshots - cinderCurr.Snapshots.Limit,
			},
		}
	}

	// 실제 적용(dryRun=false)
	if !req.DryRun {
		// Nova 쿼터 적용
		cores, ram, inst := p.Cores, p.RAMMB, p.Instances
		if err := s.OS.ApplyNovaQuota(ctx, req.ProjectID, &cores, &ram, &inst); err != nil {
			WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "nova apply failed: " + err.Error()})
			return
		}

		// Cinder 쿼터 적용
		vols, snaps, gigs := p.Volumes, p.Snapshots, p.Gigabytes
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
