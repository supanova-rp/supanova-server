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

-- name: IsUserEnrolledInCourse :one
SELECT EXISTS(SELECT 1 FROM usercourses WHERE user_id = $1 AND course_id = $2);

-- name: EnrolInCourse :exec
INSERT INTO usercourses (user_id, course_id) VALUES ($1, $2);

-- name: DisenrolInCourse :exec
DELETE FROM usercourses WHERE user_id = $1 AND course_id = $2;