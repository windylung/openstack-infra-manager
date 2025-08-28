package models

import "time"

// Student represents a student in the system
type Student struct {
	StudentID         string    `json:"student_id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Department        string    `json:"department"`
	KeystoneUserID    string    `json:"keystone_user_id,omitempty"`
	KeystoneProjectID string    `json:"keystone_project_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// StudentCreateRequest represents the request to create a new student
type StudentCreateRequest struct {
	StudentID  string `json:"student_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Department string `json:"department" validate:"required"`
}

// StudentUpdateRequest represents the request to update a student
type StudentUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty" validate:"omitempty,email"`
	Department string `json:"department,omitempty"`
}
