package services

import (
	"context"
	"fmt"
	"log"

	"example.com/quotaapi/internal/database"
	"example.com/quotaapi/internal/models"
	"example.com/quotaapi/internal/openstack"
)

// QuotaReconciliationService handles bulk quota reconciliation
type QuotaReconciliationService struct {
	db         *database.Database
	projectMgr *openstack.ProjectManager
}

// NewQuotaReconciliationService creates a new reconciliation service
func NewQuotaReconciliationService(db *database.Database, projectMgr *openstack.ProjectManager) *QuotaReconciliationService {
	return &QuotaReconciliationService{
		db:         db,
		projectMgr: projectMgr,
	}
}

// StudentQuotaSummary represents a student's quota summary
type StudentQuotaSummary struct {
	StudentID      string              `json:"student_id"`
	StudentName    string              `json:"student_name"`
	BaselineQuota  models.QuotaProfile `json:"baseline_quota"`
	ActiveCourses  []models.Course     `json:"active_courses"`
	EffectiveQuota models.QuotaProfile `json:"effective_quota"`
	AppliedQuota   models.QuotaProfile `json:"applied_quota,omitempty"`
	Status         string              `json:"status"` // success, failed, pending
	ErrorMessage   string              `json:"error_message,omitempty"`
}

// BulkReconciliationResult represents the result of bulk reconciliation
type BulkReconciliationResult struct {
	TotalStudents  int                   `json:"total_students"`
	SuccessCount   int                   `json:"success_count"`
	FailedCount    int                   `json:"failed_count"`
	PendingCount   int                   `json:"pending_count"`
	StudentResults []StudentQuotaSummary `json:"student_results"`
	Summary        string                `json:"summary"`
}

// RunBulkReconciliation runs bulk quota reconciliation for all students
func (s *QuotaReconciliationService) RunBulkReconciliation(ctx context.Context) (*BulkReconciliationResult, error) {
	log.Println("Starting bulk quota reconciliation...")

	// 1. 모든 학생 조회
	students, err := s.db.GetAllStudents()
	if err != nil {
		return nil, fmt.Errorf("failed to get students: %w", err)
	}

	result := &BulkReconciliationResult{
		TotalStudents:  len(students),
		StudentResults: make([]StudentQuotaSummary, 0, len(students)),
	}

	// 2. 각 학생별로 리콘실 실행
	for _, student := range students {
		studentResult := s.reconcileStudentQuota(ctx, student)
		result.StudentResults = append(result.StudentResults, studentResult)

		// 통계 업데이트
		switch studentResult.Status {
		case "success":
			result.SuccessCount++
		case "failed":
			result.FailedCount++
		case "pending":
			result.PendingCount++
		}
	}

	// 3. 결과 요약 생성
	result.Summary = fmt.Sprintf("Reconciliation completed: %d success, %d failed, %d pending",
		result.SuccessCount, result.FailedCount, result.PendingCount)

	log.Printf("Bulk reconciliation completed: %s", result.Summary)
	return result, nil
}

// ReconcileQuota reconciles quota for a single student after enrollment changes
func (s *QuotaReconciliationService) ReconcileQuota(ctx context.Context, studentID, courseID string) error {
	// 학생 정보 조회
	student, err := s.db.GetStudent(studentID)
	if err != nil {
		return fmt.Errorf("failed to get student %s: %w", studentID, err)
	}

	// 개별 학생 리콘실 실행
	summary := s.reconcileStudentQuota(ctx, student)

	if summary.Status == "success" {
		log.Printf("Successfully reconciled quota for student %s after enrollment in course %s", studentID, courseID)
	} else {
		log.Printf("Failed to reconcile quota for student %s: %s", studentID, summary.ErrorMessage)
	}

	return nil
}

// reconcileStudentQuota reconciles quota for a single student
func (s *QuotaReconciliationService) reconcileStudentQuota(ctx context.Context, student *models.Student) StudentQuotaSummary {
	summary := StudentQuotaSummary{
		StudentID:     student.StudentID,
		StudentName:   student.Name,
		BaselineQuota: models.GetBasicProfile(), // 기본 프로파일
		Status:        "pending",
	}

	// 1. 학생의 활성 수강 과목 조회
	enrollments, err := s.db.GetStudentEnrollments(student.StudentID)
	if err != nil {
		summary.Status = "failed"
		summary.ErrorMessage = fmt.Sprintf("failed to get enrollments: %v", err)
		return summary
	}

	// 2. 활성 과목만 필터링
	var activeCourses []models.Course
	for _, enrollment := range enrollments {
		if enrollment.Status == "active" {
			course, err := s.db.GetCourse(enrollment.CourseID)
			if err != nil {
				log.Printf("Warning: failed to get course %s: %v", enrollment.CourseID, err)
				continue
			}
			activeCourses = append(activeCourses, *course)
		}
	}
	summary.ActiveCourses = activeCourses

	// 3. 유효 쿼터 계산: baseline + Σ(활성 과목 쿼타)
	effectiveQuota := s.calculateEffectiveQuota(summary.BaselineQuota, activeCourses)
	summary.EffectiveQuota = effectiveQuota

	// 4. OpenStack 프로젝트가 있는 경우 쿼타 적용
	if student.KeystoneProjectID != "" && s.projectMgr != nil {
		if err := s.applyQuotaToOpenStack(ctx, student.KeystoneProjectID, effectiveQuota); err != nil {
			summary.Status = "failed"
			summary.ErrorMessage = fmt.Sprintf("failed to apply quota: %v", err)
			return summary
		}
		summary.AppliedQuota = effectiveQuota
		summary.Status = "success"
	} else {
		summary.Status = "pending"
		summary.ErrorMessage = "no OpenStack project or project manager"
	}

	return summary
}

// calculateEffectiveQuota calculates effective quota by adding course quotas to baseline
func (s *QuotaReconciliationService) calculateEffectiveQuota(baseline models.QuotaProfile, courses []models.Course) models.QuotaProfile {
	effective := baseline // baseline 복사

	// 각 활성 과목의 쿼타를 baseline에 추가
	for _, course := range courses {
		effective.Instances += course.QuotaProfile.Instances
		effective.Cores += course.QuotaProfile.Cores
		effective.RAMMB += course.QuotaProfile.RAMMB
		effective.Volumes += course.QuotaProfile.Volumes
		effective.Gigabytes += course.QuotaProfile.Gigabytes
		effective.Ports += course.QuotaProfile.Ports
		effective.FloatingIPs += course.QuotaProfile.FloatingIPs
		effective.Snapshots += course.QuotaProfile.Snapshots
	}

	return effective
}

// applyQuotaToOpenStack applies the calculated quota to OpenStack project
func (s *QuotaReconciliationService) applyQuotaToOpenStack(ctx context.Context, projectID string, quota models.QuotaProfile) error {
	// Nova 쿼타 적용
	cores := quota.Cores
	ramMB := quota.RAMMB
	instances := quota.Instances
	if err := s.projectMgr.GetClients().ApplyNovaQuota(ctx, projectID, &cores, &ramMB, &instances); err != nil {
		return fmt.Errorf("failed to apply Nova quota: %w", err)
	}

	// Cinder 쿼타 적용
	volumes := quota.Volumes
	gigabytes := quota.Gigabytes
	snapshots := quota.Snapshots
	if err := s.projectMgr.GetClients().ApplyCinderQuota(ctx, projectID, &volumes, &snapshots, &gigabytes); err != nil {
		return fmt.Errorf("failed to apply Cinder quota: %w", err)
	}

	// Neutron 쿼타 적용
	ports := quota.Ports
	floatingIPs := quota.FloatingIPs
	if err := s.projectMgr.GetClients().ApplyNeutronQuota(ctx, projectID, &ports, &floatingIPs); err != nil {
		return fmt.Errorf("failed to apply Neutron quota: %w", err)
	}

	log.Printf("Applied quota to project %s: vCPU=%d, RAM=%dMB, Instances=%d, Volumes=%d, Disk=%dGB, Ports=%d, FloatingIPs=%d",
		projectID, cores, ramMB, instances, volumes, gigabytes, ports, floatingIPs)

	return nil
}
