package models

import (
	"time"
)

// Enrollment represents a student's enrollment in a course
type Enrollment struct {
	StudentID string    `json:"student_id"`
	CourseID  string    `json:"course_id"`
	Status    string    `json:"status"` // active, completed, dropped
	StartAt   time.Time `json:"start_at"`
	EndAt     time.Time `json:"end_at"`
}

// EnrollmentCreateRequest represents the request to enroll a student in a course
type EnrollmentCreateRequest struct {
	CourseID string `json:"course_id" validate:"required"`
	Status   string `json:"status" validate:"required,oneof=active completed dropped"`
}

// EnrollmentUpdateRequest represents the request to update an enrollment
type EnrollmentUpdateRequest struct {
	Status  *string `json:"status,omitempty" validate:"omitempty,oneof=active completed dropped"`
	StartAt *string `json:"start_at,omitempty"`
	EndAt   *string `json:"end_at,omitempty"`
}

// EnrollmentWithDetails represents an enrollment with course and student details
type EnrollmentWithDetails struct {
	Enrollment
	Course  *Course  `json:"course,omitempty"`
	Student *Student `json:"student,omitempty"`
}
