
-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: GetUsersAndAssignedCourses :many
SELECT
  u.id,
  u.name,
  u.email,
  json_agg(json_build_object('id', c.id, 'title', c.title)) AS courses
FROM users u
LEFT JOIN usercourses uc ON u.id = uc.user_id
LEFT JOIN courses c ON uc.course_id = c.id
GROUP BY u.id, u.name, u.email
ORDER BY u.name;