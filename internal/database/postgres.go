package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) DB() *sql.DB {
	return d.db
}

func initSchema(db *sql.DB) error {
	schema := `
	-- students 테이블
	CREATE TABLE IF NOT EXISTS students (
		student_id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		department TEXT NOT NULL,
		keystone_project_id TEXT NOT NULL,
		keystone_user_id TEXT,
		created_at TIMESTAMPTZ DEFAULT now()
	);

	-- courses 테이블
	CREATE TABLE IF NOT EXISTS courses (
		course_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		department TEXT NOT NULL,
		semester TEXT NOT NULL,
		start_at TIMESTAMPTZ NOT NULL,
		end_at TIMESTAMPTZ NOT NULL,
		quota_profile JSONB NOT NULL,
		defaults JSONB,
		created_at TIMESTAMPTZ DEFAULT now()
	);

	-- enrollments 테이블
	CREATE TABLE IF NOT EXISTS enrollments (
		student_id TEXT NOT NULL REFERENCES students(student_id) ON DELETE CASCADE,
		course_id TEXT NOT NULL REFERENCES courses(course_id) ON DELETE CASCADE,
		status TEXT NOT NULL CHECK (status IN ('active', 'completed', 'dropped')),
		start_at TIMESTAMPTZ NOT NULL,
		end_at TIMESTAMPTZ NOT NULL,
		PRIMARY KEY (student_id, course_id)
	);

	-- 인덱스 생성
	CREATE INDEX IF NOT EXISTS idx_enrollments_student_id ON enrollments(student_id);
	CREATE INDEX IF NOT EXISTS idx_enrollments_course_id ON enrollments(course_id);
	CREATE INDEX IF NOT EXISTS idx_courses_date_range ON courses(start_at, end_at);
	`

	_, err := db.Exec(schema)
	return err
}
