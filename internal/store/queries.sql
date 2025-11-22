-- name: GetCourseById :one
SELECT id, title, description FROM courses WHERE id = $1;

-- name: AddCourse :one
INSERT INTO courses (title, description) VALUES ($1, $2) RETURNING id;

-- name: GetProgressById :one
SELECT completed_intro, completed_section_ids FROM userprogress WHERE user_id = $1 AND course_id = $2;