package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"example.com/quotaapi/internal/database"
	"example.com/quotaapi/internal/models"
	"example.com/quotaapi/internal/openstack"
	"example.com/quotaapi/internal/services"
)

type StudentHandler struct {
	db                    *database.Database
	projectMgr            *openstack.ProjectManager
	reconciliationService *services.QuotaReconciliationService
}

func NewStudentHandler(db *database.Database, projectMgr *openstack.ProjectManager) *StudentHandler {
	reconciliationService := services.NewQuotaReconciliationService(db, projectMgr)
	return &StudentHandler{
		db:                    db,
		projectMgr:            projectMgr,
		reconciliationService: reconciliationService,
	}
}

func (h *StudentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Printf("DEBUG: StudentHandler received request: %s %s\n", r.Method, path)

	switch {
	case r.Method == "POST" && path == "/students":
		fmt.Printf("DEBUG: Routing to createStudent\n")
		h.createStudent(w, r)
	case r.Method == "GET" && path == "/students":
		fmt.Printf("DEBUG: Routing to listStudents\n")
		h.listStudents(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/students/"):
		fmt.Printf("DEBUG: Routing to getStudent\n")
		h.getStudent(w, r)
	case r.Method == "POST" && strings.Contains(path, "/enroll"):
		fmt.Printf("DEBUG: Routing to enrollStudent\n")
		h.enrollStudent(w, r)
	case r.Method == "DELETE" && strings.Contains(path, "/enroll"):
		fmt.Printf("DEBUG: Routing to unenrollStudent\n")
		h.unenrollStudent(w, r)
	case r.Method == "GET" && strings.Contains(path, "/enrollments"):
		fmt.Printf("DEBUG: Routing to getStudentEnrollments\n")
		h.getStudentEnrollments(w, r)
	default:
		fmt.Printf("DEBUG: No route matched, returning 404\n")
		http.NotFound(w, r)
	}
}

func (h *StudentHandler) createStudent(w http.ResponseWriter, r *http.Request) {
	var req models.StudentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}

	// 기본 검증
	if req.StudentID == "" || req.Name == "" || req.Email == "" || req.Department == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields"})
		return
	}

	student := &models.Student{
		StudentID:         req.StudentID,
		Name:              req.Name,
		Email:             req.Email,
		Department:        req.Department,
		KeystoneProjectID: "", // OpenStack에서 생성될 예정
		KeystoneUserID:    "", // OpenStack에서 생성될 예정
		CreatedAt:         time.Now(),
	}

	// 1. 데이터베이스에 학생 정보 저장
	if err := h.db.CreateStudent(student); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create student: " + err.Error()})
		return
	}

	// 2. OpenStack에 프로젝트 생성 (백그라운드에서 처리)
	if h.projectMgr != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := h.projectMgr.CreateStudentProject(ctx, student); err != nil {
				fmt.Printf("Warning: Failed to create OpenStack project for student %s: %v\n", student.StudentID, err)
				return
			}

			// 프로젝트 ID 업데이트
			if err := h.db.UpdateStudent(student.StudentID, map[string]interface{}{
				"keystone_project_id": student.KeystoneProjectID,
				"keystone_user_id":    student.KeystoneUserID,
			}); err != nil {
				fmt.Printf("Warning: Failed to update student with project IDs: %v\n", err)
			}
		}()
	}

	WriteJSON(w, http.StatusCreated, student)
}

func (h *StudentHandler) getStudent(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid student id"})
		return
	}

	studentID := pathParts[2]
	student, err := h.db.GetStudent(studentID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, map[string]any{"error": "student not found"})
		return
	}

	WriteJSON(w, http.StatusOK, student)
}

func (h *StudentHandler) listStudents(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get("department")

	students, err := h.db.ListStudents(department)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list students: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, students)
}

func (h *StudentHandler) enrollStudent(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid path"})
		return
	}

	studentID := pathParts[2]
	var req models.EnrollmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}

	// 과목 정보 조회하여 기간 가져오기
	course, err := h.db.GetCourse(req.CourseID)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "course not found: " + err.Error()})
		return
	}

	enrollment := &models.Enrollment{
		StudentID: studentID,
		CourseID:  req.CourseID,
		Status:    req.Status,
		StartAt:   course.StartAt, // 과목의 시작일 사용
		EndAt:     course.EndAt,   // 과목의 종료일 사용
	}

	if err := h.db.EnrollStudent(enrollment); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to enroll: " + err.Error()})
		return
	}

	// 학생 프로젝트 크기 조정 및 할당량 재조정
	if h.reconciliationService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.reconciliationService.ReconcileQuota(ctx, studentID, req.CourseID); err != nil {
				fmt.Printf("Warning: Failed to reconcile quota for student %s after enrollment: %v\n", studentID, err)
			}
		}()
	}

	WriteJSON(w, http.StatusCreated, enrollment)
}

func (h *StudentHandler) unenrollStudent(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid path"})
		return
	}

	studentID := pathParts[2]
	courseID := pathParts[4]

	if err := h.db.UnenrollStudent(studentID, courseID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to unenroll: " + err.Error()})
		return
	}

	// 수강 해제 후 쿼타 재조정
	if h.reconciliationService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := h.reconciliationService.ReconcileQuota(ctx, studentID, courseID); err != nil {
				fmt.Printf("Warning: Failed to reconcile quota for student %s after unenrollment from course %s: %v\n", studentID, courseID, err)
			}
		}()
	}

	WriteJSON(w, http.StatusOK, map[string]any{"message": "unenrolled successfully"})
}

func (h *StudentHandler) getStudentEnrollments(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid path"})
		return
	}

	studentID := pathParts[2]
	enrollments, err := h.db.GetStudentEnrollments(studentID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to get enrollments: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, enrollments)
}

// listOpenStackProjects lists all OpenStack projects
func (h *StudentHandler) ListOpenStackProjects(w http.ResponseWriter, r *http.Request) {
	if h.projectMgr == nil {
		WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "OpenStack not available"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	projects, err := h.projectMgr.ListAllProjects(ctx)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list projects: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"projects": projects,
		"count":    len(projects),
	})
}

// findStudentProject finds a specific student's project in OpenStack
func (h *StudentHandler) FindStudentProject(w http.ResponseWriter, r *http.Request) {
	if h.projectMgr == nil {
		WriteJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "OpenStack not available"})
		return
	}

	// URL에서 student_id 추출: /openstack/projects/{student_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid path"})
		return
	}
	studentID := pathParts[3]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	project, err := h.projectMgr.FindStudentProject(ctx, studentID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, map[string]any{"error": "project not found: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, project)
}
