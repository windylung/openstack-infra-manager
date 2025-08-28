package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	osapi "example.com/quotaapi/internal/openstack"
)

type ProvisionReq struct {
	Name           string   `json:"name"`
	ImageID        string   `json:"imageId"`
	FlavorID       string   `json:"flavorId"`
	NetworkID      string   `json:"networkId"`
	KeyName        string   `json:"keyName"`
	SecurityGroups []string `json:"securityGroups"`
	AssignFloating bool     `json:"assignFloatingIp"`
	FloatingIP     string   `json:"floatingIp"`
	ExternalNetID  string   `json:"externalNetworkId"`
	UserData       string   `json:"userData,omitempty"` // 선택
}

type ProvisionResp struct {
	ServerID   string `json:"serverId"`
	Status     string `json:"status"`
	FixedIP    string `json:"fixedIp,omitempty"`
	FloatingIP string `json:"floatingIp,omitempty"`
}

func NewProvisionServerHandler(osc *osapi.Clients) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProvisionReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + err.Error()})
			return
		}

		// 최소 검증
		if req.Name == "" || req.ImageID == "" || req.FlavorID == "" || req.NetworkID == "" || req.KeyName == "" {
			WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
		defer cancel()

		res, err := osc.ProvisionServer(ctx, osapi.ProvisionOpts{
			Name:           req.Name,
			ImageID:        req.ImageID,
			FlavorID:       req.FlavorID,
			NetworkID:      req.NetworkID,
			KeyName:        req.KeyName,
			SecurityGroups: req.SecurityGroups,
			UserData:       req.UserData,
			AssignFIP:      req.AssignFloating,
			FloatingIP:     req.FloatingIP,
			ExternalNetID:  req.ExternalNetID, // 비면 내부 기본값 사용 가능
		})
		if err != nil {
			WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		WriteJSON(w, http.StatusOK, ProvisionResp{
			ServerID:   res.ServerID,
			Status:     res.Status,
			FixedIP:    res.FixedIP,
			FloatingIP: res.FloatingIP,
		})
	}
}
