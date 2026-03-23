
-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: GetUsersAndAssignedCourses :many
SELECT
  u.id,
  u.name,
  u.email,
  json_agg(c.course_id) FILTER (WHERE c.course_id IS NOT NULL) AS course_ids
FROM users u
LEFT JOIN usercourses c ON u.id = c.user_id
GROUP BY u.id, u.name, u.email
ORDER BY u.name;
