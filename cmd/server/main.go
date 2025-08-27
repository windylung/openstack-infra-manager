package main

import (
	"log"
	"net/http"
	"os"

	"example.com/quotaapi/internal/config"
	httph "example.com/quotaapi/internal/http"
	osapi "example.com/quotaapi/internal/openstack"
)

func main() {
	// 1) 환경 로드/검증
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2) 라우팅
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httph.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	http.HandleFunc("/auth/check", httph.NewAuthCheckHandler(cfg))

	osc, err := osapi.NewServiceClients(cfg)
	if err != nil {
		log.Fatal(err)
	}

	qget := &httph.QuotaGetServer{OS: osc}
	http.HandleFunc("/quota/current", qget.Current)

	srv := &httph.QuotaApplyServer{OS: osc}
	http.HandleFunc("/quota/apply", srv.QuotaApply)

	http.HandleFunc("/quota/applyProfile", httph.NewApplyProfileHandler(osc))

	// 3) 서버 시작
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
