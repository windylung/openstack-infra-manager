package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"example.com/quotaapi/internal/config"
	"example.com/quotaapi/internal/database"
	httph "example.com/quotaapi/internal/http"
	osapi "example.com/quotaapi/internal/openstack"
	"example.com/quotaapi/internal/services"

	flavors "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	keypairs "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	images "github.com/gophercloud/gophercloud/v2/openstack/image/v2/images"
	secgroups "github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
)

func main() {
	// 1) 환경 로드/검증
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2) 데이터베이스 연결
	db, err := database.NewDatabase("postgres://quota_user:quota_password@localhost:5432/quota_db?sslmode=disable")
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	// 3) OpenStack 클라이언트 초기화
	osc, err := osapi.NewServiceClients(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// 4) 라우팅
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httph.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	http.HandleFunc("/auth/check", httph.NewAuthCheckHandler(cfg))

	// 4-1) 시작 시 토큰/기본 자원 목록 출력
	printBasics(osc)

	// 4-2) 기존 쿼타 API
	qget := &httph.QuotaGetServer{OS: osc}
	http.HandleFunc("/quota/current", qget.Current)

	srv := &httph.QuotaApplyServer{OS: osc}
	http.HandleFunc("/quota/apply", srv.QuotaApply)

	http.HandleFunc("/quota/applyProfile", httph.NewApplyProfileHandler(osc))

	// 4-3) 프로비저닝 엔드포인트
	provision := httph.NewProvisionServerHandler(osc)
	http.HandleFunc("/provision/server", provision)

	// 4-4) 새로운 학생/수업/수강 관리 API
	var studentHandler *httph.StudentHandler
	if osc != nil {
		// OpenStack 클라이언트가 있을 때만 ProjectManager 생성
		projectMgr := osapi.NewProjectManager(osc)
		studentHandler = httph.NewStudentHandler(db, projectMgr)

		// 4-5) 리콘실 서비스 및 핸들러
		reconciliationService := services.NewQuotaReconciliationService(db, projectMgr)
		reconciliationHandler := httph.NewReconciliationHandler(reconciliationService)
		http.HandleFunc("/reconciliation/", reconciliationHandler.ServeHTTP)
	} else {
		// OpenStack 클라이언트가 없을 때는 nil로 전달
		studentHandler = httph.NewStudentHandler(db, nil)
	}

	http.HandleFunc("/students", studentHandler.ServeHTTP)
	http.HandleFunc("/students/", studentHandler.ServeHTTP) // Handles /students/{id}, /students/{id}/enroll, /students/{id}/enrollments

	// OpenStack 프로젝트 정보 조회 엔드포인트
	http.HandleFunc("/openstack/projects", studentHandler.ListOpenStackProjects)
	http.HandleFunc("/openstack/projects/", studentHandler.FindStudentProject)

	courseHandler := httph.NewCourseHandler(db)
	http.HandleFunc("/courses", courseHandler.ServeHTTP)
	http.HandleFunc("/courses/", courseHandler.ServeHTTP)

	// 5) 서버 시작
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func printBasics(osc *osapi.Clients) {
	ctx := context.Background()
	fmt.Printf("Issued Token ID: %s\n", osc.Provider.TokenID)

	// Images (Glance)
	fmt.Println("--- Images ---")
	if pages, err := images.List(osc.ImageV2, images.ListOpts{}).AllPages(ctx); err == nil {
		if list, err := images.ExtractImages(pages); err == nil {
			for _, img := range list {
				fmt.Printf("ID: %s, Name: %s\n", img.ID, img.Name)
			}
		} else {
			fmt.Printf("images extract error: %v\n", err)
		}
	} else {
		fmt.Printf("images list error: %v\n", err)
	}

	// Flavors (Nova)
	fmt.Println()
	fmt.Println("--- Flavors ---")
	if pages, err := flavors.ListDetail(osc.ComputeV2, flavors.ListOpts{}).AllPages(ctx); err == nil {
		if list, err := flavors.ExtractFlavors(pages); err == nil {
			for _, f := range list {
				fmt.Printf("ID: %s, Name: %s, vCPUs: %d, RAM: %dMB, Disk: %dGB\n", f.ID, f.Name, f.VCPUs, f.RAM, f.Disk)
			}
		} else {
			fmt.Printf("flavors extract error: %v\n", err)
		}
	} else {
		fmt.Printf("flavors list error: %v\n", err)
	}

	// Keypairs (Nova)
	fmt.Println()
	fmt.Println("--- Keypairs ---")
	if pages, err := keypairs.List(osc.ComputeV2, keypairs.ListOpts{}).AllPages(ctx); err == nil {
		if list, err := keypairs.ExtractKeyPairs(pages); err == nil {
			for _, k := range list {
				fmt.Printf("Name: %s, Fingerprint: %s\n", k.Name, k.Fingerprint)
			}
		} else {
			fmt.Printf("keypairs extract error: %v\n", err)
		}
	} else {
		fmt.Printf("keypairs list error: %v\n", err)
	}

	// Security Groups (Neutron)
	fmt.Println()
	fmt.Println("--- Security Groups ---")
	if pages, err := secgroups.List(osc.NetworkV2, secgroups.ListOpts{}).AllPages(ctx); err == nil {
		if list, err := secgroups.ExtractGroups(pages); err == nil {
			for _, g := range list {
				fmt.Printf("ID: %s, Name: %s, Description: %s\n", g.ID, g.Name, g.Description)
			}
		} else {
			fmt.Printf("secgroups extract error: %v\n", err)
		}
	} else {
		fmt.Printf("secgroups list error: %v\n", err)
	}
}
