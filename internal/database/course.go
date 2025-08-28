package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"example.com/quotaapi/internal/models"
)

// CreateCourse creates a new course in the database
func (db *Database) CreateCourse(course *models.Course) error {
	query := `
		INSERT INTO courses (course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	quotaProfileJSON, err := json.Marshal(course.QuotaProfile)
	if err != nil {
		return fmt.Errorf("failed to marshal quota profile: %w", err)
	}

	var defaultsJSON []byte
	if course.Defaults != nil {
		// 디버깅: 실제 데이터 확인
		fmt.Printf("DEBUG: course.Defaults = %+v\n", course.Defaults)
		defaultsJSON, err = json.Marshal(course.Defaults)
		if err != nil {
			return fmt.Errorf("failed to marshal defaults: %w", err)
		}
		fmt.Printf("DEBUG: defaultsJSON = %s\n", string(defaultsJSON))
	} else {
		fmt.Printf("DEBUG: course.Defaults is nil\n")
	}

	_, err = db.db.Exec(query,
		course.CourseID,
		course.Title,
		course.Department,
		course.Semester,
		course.StartAt,
		course.EndAt,
		quotaProfileJSON,
		defaultsJSON,
		course.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create course: %w", err)
	}

	return nil
}

// GetCourse retrieves a course by ID
func (db *Database) GetCourse(courseID string) (*models.Course, error) {
	query := `
		SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
		FROM courses WHERE course_id = $1
	`
	row := db.db.QueryRow(query, courseID)

	var course models.Course
	var quotaProfileJSON, defaultsJSON []byte

	err := row.Scan(&course.CourseID, &course.Title, &course.Department, &course.Semester,
		&course.StartAt, &course.EndAt, &quotaProfileJSON, &defaultsJSON, &course.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("course not found: %s", courseID)
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	// Parse quota profile JSON
	if err := json.Unmarshal(quotaProfileJSON, &course.QuotaProfile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quota profile: %w", err)
	}

	// Parse defaults JSON if present
	if defaultsJSON != nil {
		course.Defaults = &models.CourseDefaults{}
		if err := json.Unmarshal(defaultsJSON, course.Defaults); err != nil {
			return nil, fmt.Errorf("failed to unmarshal defaults: %w", err)
		}
	}

	return &course, nil
}

// UpdateCourse updates an existing course
func (db *Database) UpdateCourse(courseID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE courses SET "
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		if argCount > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", field, argCount)
		args = append(args, value)
		argCount++
	}

	query += fmt.Sprintf(", updated_at = $%d WHERE course_id = $%d", argCount, argCount+1)
	args = append(args, time.Now(), courseID)

	_, err := db.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update course: %w", err)
	}
	return nil
}

// DeleteCourse deletes a course by ID
func (db *Database) DeleteCourse(courseID string) error {
	query := `DELETE FROM courses WHERE course_id = $1`
	result, err := db.db.Exec(query, courseID)
	if err != nil {
		return fmt.Errorf("failed to delete course: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("course not found: %s", courseID)
	}
	return nil
}

// ListCourses retrieves all courses with optional filtering
func (db *Database) ListCourses(department, semester string) ([]models.Course, error) {
	var query string
	var args []interface{}

	if department != "" && semester != "" {
		query = `SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
				 FROM courses WHERE department = $1 AND semester = $2 ORDER BY start_at DESC`
		args = []interface{}{department, semester}
	} else if department != "" {
		query = `SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
				 FROM courses WHERE department = $1 ORDER BY start_at DESC`
		args = []interface{}{department}
	} else if semester != "" {
		query = `SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
				 FROM semester = $1 ORDER BY start_at DESC`
		args = []interface{}{semester}
	} else {
		query = `SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
				 FROM courses ORDER BY start_at DESC`
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list courses: %w", err)
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var course models.Course
		var quotaProfileJSON, defaultsJSON []byte

		err := rows.Scan(&course.CourseID, &course.Title, &course.Department, &course.Semester,
			&course.StartAt, &course.EndAt, &quotaProfileJSON, &defaultsJSON, &course.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}

		// Parse quota profile JSON
		if err := json.Unmarshal(quotaProfileJSON, &course.QuotaProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quota profile: %w", err)
		}

		// Parse defaults JSON if present
		if defaultsJSON != nil {
			course.Defaults = &models.CourseDefaults{}
			if err := json.Unmarshal(defaultsJSON, course.Defaults); err != nil {
				return nil, fmt.Errorf("failed to unmarshal defaults: %w", err)
			}
		}

		courses = append(courses, course)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return courses, nil
}

// GetActiveCourses retrieves courses that are currently active
func (db *Database) GetActiveCourses() ([]models.Course, error) {
	query := `
		SELECT course_id, title, department, semester, start_at, end_at, quota_profile, defaults, created_at
		FROM courses 
		WHERE start_at <= $1 AND end_at >= $1
		ORDER BY start_at DESC
	`
	now := time.Now()

	rows, err := db.db.Query(query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active courses: %w", err)
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var course models.Course
		var quotaProfileJSON, defaultsJSON []byte

		err := rows.Scan(&course.CourseID, &course.Title, &course.Department, &course.Semester,
			&course.StartAt, &course.EndAt, &quotaProfileJSON, &defaultsJSON, &course.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}

		// Parse quota profile JSON
		if err := json.Unmarshal(quotaProfileJSON, &course.QuotaProfile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quota profile: %w", err)
		}

		// Parse defaults JSON if present
		if defaultsJSON != nil {
			course.Defaults = &models.CourseDefaults{}
			if err := json.Unmarshal(defaultsJSON, course.Defaults); err != nil {
				return nil, fmt.Errorf("failed to unmarshal defaults: %w", err)
			}
		}

		courses = append(courses, course)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating courses: %w", err)
	}

	return courses, nil
}
