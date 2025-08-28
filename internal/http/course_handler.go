package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"example.com/quotaapi/internal/database"
	"example.com/quotaapi/internal/models"
)

type CourseHandler struct {
	db *database.Database
}

func NewCourseHandler(db *database.Database) *CourseHandler {
	return &CourseHandler{db: db}
}

func (h *CourseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case r.Method == "POST" && path == "/courses":
		h.createCourse(w, r)
	case r.Method == "GET" && path == "/courses":
		h.listCourses(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/courses/"):
		h.getCourse(w, r)
	case r.Method == "PUT" && strings.HasPrefix(path, "/courses/"):
		h.updateCourse(w, r)
	case r.Method == "DELETE" && strings.HasPrefix(path, "/courses/"):
		h.deleteCourse(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *CourseHandler) createCourse(w http.ResponseWriter, r *http.Request) {
	var req models.CourseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}

	// 기본 검증
	if req.CourseID == "" || req.Title == "" || req.Department == "" || req.Semester == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "missing required fields"})
		return
	}

	startAt, err := time.Parse("2006-01-02", req.StartAt)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid start_at format (use YYYY-MM-DD)"})
		return
	}

	endAt, err := time.Parse("2006-01-02", req.EndAt)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid end_at format (use YYYY-MM-DD)"})
		return
	}

	course := &models.Course{
		CourseID:     req.CourseID,
		Title:        req.Title,
		Department:   req.Department,
		Semester:     req.Semester,
		StartAt:      startAt,
		EndAt:        endAt,
		QuotaProfile: req.QuotaProfile,
		Defaults:     req.Defaults,
		CreatedAt:    time.Now(),
	}

	if err := h.db.CreateCourse(course); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create course: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusCreated, course)
}

func (h *CourseHandler) getCourse(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid course id"})
		return
	}

	courseID := pathParts[2]
	course, err := h.db.GetCourse(courseID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, map[string]any{"error": "course not found"})
		return
	}

	WriteJSON(w, http.StatusOK, course)
}

func (h *CourseHandler) listCourses(w http.ResponseWriter, r *http.Request) {
	department := r.URL.Query().Get("department")
	semester := r.URL.Query().Get("semester")

	courses, err := h.db.ListCourses(department, semester)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list courses: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, courses)
}

func (h *CourseHandler) updateCourse(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid course id"})
		return
	}

	courseID := pathParts[2]
	var req models.CourseUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Department != nil {
		updates["department"] = *req.Department
	}
	if req.Semester != nil {
		updates["semester"] = *req.Semester
	}
	if req.StartAt != nil {
		if startAt, err := time.Parse("2006-01-02", *req.StartAt); err == nil {
			updates["start_at"] = startAt
		}
	}
	if req.EndAt != nil {
		if endAt, err := time.Parse("2006-01-02", *req.EndAt); err == nil {
			updates["end_at"] = endAt
		}
	}
	if req.QuotaProfile != nil {
		updates["quota_profile"] = req.QuotaProfile
	}
	if req.Defaults != nil {
		updates["defaults"] = req.Defaults
	}

	if len(updates) == 0 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "no fields to update"})
		return
	}

	if err := h.db.UpdateCourse(courseID, updates); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to update course: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"message": "course updated successfully"})
}

func (h *CourseHandler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid course id"})
		return
	}

	courseID := pathParts[2]
	if err := h.db.DeleteCourse(courseID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to delete course: " + err.Error()})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{"message": "course deleted successfully"})
}
