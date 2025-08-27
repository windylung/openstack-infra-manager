package http

import (
	"log"
	"net/http"

	"example.com/quotaapi/internal/config"
	osauth "example.com/quotaapi/internal/openstack"
)

// GET /auth/check
// - 헤더 X-Auth-Token 이 있으면 그 토큰을 검증
// - 없으면 서비스 토큰(환경변수 계정으로 발급된 토큰)을 검증
func NewAuthCheckHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "GET only"})
			return
		}
		provider, ident, err := osauth.NewIdentity(cfg)
		if err != nil {
			WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "keystone auth error: " + err.Error()})
			return
		}
		subject := r.Header.Get("X-Auth-Token")
		raw, err := osauth.IntrospectToken(r.Context(), ident, provider, subject)
		if err != nil {
			WriteJSON(w, http.StatusBadGateway, map[string]string{"error": "token introspection failed: " + err.Error()})
			return
		}

		// 안전 캐스팅
		token, _ := raw["token"].(map[string]any)
		user, _ := token["user"].(map[string]any)
		project, _ := token["project"].(map[string]any)
		roles, _ := token["roles"].([]any)
		log.Println(raw["token"])

		WriteJSON(w, http.StatusOK, map[string]any{
			"auth": "ok",
			"user": map[string]any{
				"id":     safe(user, "id"),
				"name":   safe(user, "name"),
				"domain": user["domain"],
			},
			"project": map[string]any{
				"id":     safe(project, "id"),
				"name":   safe(project, "name"),
				"domain": project["domain"],
			},
			"token": map[string]any{
				"issued_at":  safe(token, "issued_at"),
				"expires_at": safe(token, "expires_at"),
			},
			"roles": roles, // [{id,name}, ...]
		})
	}
}

func safe(m map[string]any, k string) any {
	if m == nil {
		return ""
	}
	if v, ok := m[k]; ok && v != nil {
		return v
	}
	return ""
}
