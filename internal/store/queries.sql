-- name: GetCourseById :one
SELECT id, title, description FROM courses WHERE id = $1;

-- name: AddCourse :one
INSERT INTO courses (title, description) VALUES ($1, $2) RETURNING id;