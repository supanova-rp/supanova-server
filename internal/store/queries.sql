-- name: GetCourseById :one
SELECT id, title, description FROM courses WHERE id = $1;
