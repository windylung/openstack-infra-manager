package database

import (
	"fmt"

	"example.com/quotaapi/internal/models"
	_ "github.com/lib/pq"
)

// CreateStudent creates a new student in the database
func (db *Database) CreateStudent(student *models.Student) error {
	query := `
		INSERT INTO students (student_id, name, email, department, keystone_user_id, keystone_project_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.db.Exec(query,
		student.StudentID, student.Name, student.Email, student.Department,
		student.KeystoneUserID, student.KeystoneProjectID)
	if err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}
	return nil
}

// GetStudent retrieves a single student by ID
func (db *Database) GetStudent(studentID string) (*models.Student, error) {
	query := "SELECT student_id, name, email, department, keystone_project_id, keystone_user_id, created_at FROM students WHERE student_id = $1"

	student := &models.Student{}
	err := db.db.QueryRow(query, studentID).Scan(
		&student.StudentID,
		&student.Name,
		&student.Email,
		&student.Department,
		&student.KeystoneProjectID,
		&student.KeystoneUserID,
		&student.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get student %s: %w", studentID, err)
	}

	return student, nil
}

// UpdateStudent updates an existing student
func (db *Database) UpdateStudent(studentID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE students SET "
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

	query += fmt.Sprintf(" WHERE student_id = $%d", argCount)
	args = append(args, studentID)

	_, err := db.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}
	return nil
}

// DeleteStudent deletes a student by ID
func (db *Database) DeleteStudent(studentID string) error {
	query := `DELETE FROM students WHERE student_id = $1`
	result, err := db.db.Exec(query, studentID)
	if err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("student not found: %s", studentID)
	}
	return nil
}

// ListStudents retrieves all students with optional filtering
func (db *Database) ListStudents(department string) ([]models.Student, error) {
	var query string
	var args []interface{}

	if department != "" {
		query = `SELECT student_id, name, email, department, keystone_user_id, keystone_project_id, created_at
				 FROM students WHERE department = $1 ORDER BY student_id`
		args = []interface{}{department}
	} else {
		query = `SELECT student_id, name, email, department, keystone_user_id, keystone_project_id, created_at
				 FROM students ORDER BY student_id`
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list students: %w", err)
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		err := rows.Scan(&student.StudentID, &student.Name, &student.Email, &student.Department,
			&student.KeystoneUserID, &student.KeystoneProjectID, &student.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student: %w", err)
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating students: %w", err)
	}

	return students, nil
}

// GetAllStudents retrieves all students from the database
func (db *Database) GetAllStudents() ([]*models.Student, error) {
	query := "SELECT student_id, name, email, department, keystone_project_id, keystone_user_id, created_at FROM students ORDER BY student_id"

	rows, err := db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query students: %w", err)
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		student := &models.Student{}
		err := rows.Scan(
			&student.StudentID,
			&student.Name,
			&student.Email,
			&student.Department,
			&student.KeystoneProjectID,
			&student.KeystoneUserID,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student: %w", err)
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return students, nil
}
