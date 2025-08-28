package database

import (
	"fmt"

	"example.com/quotaapi/internal/models"
)

func (db *Database) EnrollStudent(enrollment *models.Enrollment) error {
	query := `
		INSERT INTO enrollments (student_id, course_id, status, start_at, end_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (student_id, course_id) DO UPDATE SET
		status = EXCLUDED.status,
		start_at = EXCLUDED.start_at,
		end_at = EXCLUDED.end_at
	`

	_, err := db.db.Exec(query, enrollment.StudentID, enrollment.CourseID, enrollment.Status, enrollment.StartAt, enrollment.EndAt)
	if err != nil {
		return fmt.Errorf("failed to enroll student: %w", err)
	}

	return nil
}

func (db *Database) UnenrollStudent(studentID, courseID string) error {
	query := `DELETE FROM enrollments WHERE student_id = $1 AND course_id = $2`

	result, err := db.db.Exec(query, studentID, courseID)
	if err != nil {
		return fmt.Errorf("failed to unenroll student: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enrollment not found")
	}

	return nil
}

func (db *Database) GetStudentEnrollments(studentID string) ([]models.Enrollment, error) {
	query := `
		SELECT student_id, course_id, status, start_at, end_at
		FROM enrollments
		WHERE student_id = $1
		ORDER BY start_at DESC
	`

	rows, err := db.db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		err := rows.Scan(&e.StudentID, &e.CourseID, &e.Status, &e.StartAt, &e.EndAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return enrollments, nil
}

func (db *Database) GetActiveEnrollments() ([]models.Enrollment, error) {
	query := `
		SELECT student_id, course_id, status, start_at, end_at
		FROM enrollments
		WHERE status = 'active' 
		AND now() BETWEEN start_at AND end_at
		ORDER BY start_at DESC
	`

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		err := rows.Scan(&e.StudentID, &e.CourseID, &e.Status, &e.StartAt, &e.EndAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return enrollments, nil
}

func (db *Database) GetActiveEnrollmentsByStudent(studentID string) ([]models.Enrollment, error) {
	query := `
		SELECT student_id, course_id, status, start_at, end_at
		FROM enrollments
		WHERE student_id = $1 
		AND status = 'active' 
		AND now() BETWEEN start_at AND end_at
		ORDER BY start_at DESC
	`

	rows, err := db.db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		err := rows.Scan(&e.StudentID, &e.CourseID, &e.Status, &e.StartAt, &e.EndAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return enrollments, nil
}
