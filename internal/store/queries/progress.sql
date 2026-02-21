-- name: GetProgress :one
SELECT completed_intro, completed_section_ids FROM userprogress WHERE user_id = $1 AND course_id = $2;

-- Insert section_id into completed_section_ids array if no entry exists
-- or append section_id to the existing array if it's not already present
-- name: UpdateProgress :exec
INSERT INTO userprogress (user_id, course_id, completed_section_ids)
VALUES (sqlc.arg('user_id'), sqlc.arg('course_id'), ARRAY[sqlc.arg('section_id')::uuid])
ON CONFLICT (user_id, course_id)
DO UPDATE SET completed_section_ids = array_append(userprogress.completed_section_ids, sqlc.arg('section_id')::uuid)
WHERE NOT (sqlc.arg('section_id') = ANY(userprogress.completed_section_ids));

-- name: HasCompletedCourse :one
SELECT completed_course FROM userprogress WHERE user_id = $1 AND course_id = $2;

-- If there is no existing userprogress (should not happen since user should have some progress already)
-- then insert new row with empty completed_section_ids */
-- name: SetCourseCompleted :exec
INSERT INTO userprogress (user_id, course_id, completed_section_ids, completed_course)
VALUES ($1, $2, ARRAY[]::uuid[], TRUE)
ON CONFLICT (user_id, course_id)
DO UPDATE SET completed_course = TRUE;

-- name: GetCompletedSectionIDsByUserID :many
SELECT completed_section_ids FROM userprogress WHERE user_id = $1;

-- name: GetAllProgress :many
SELECT
  COALESCE(uc.user_id, up.user_id) AS user_id,
  COALESCE(uc.course_id, up.course_id) AS course_id,
  u.name as user_name,
  u.email,
  c.title as course_title,
  up.completed_intro,
  up.completed_section_ids,
  up.completed_course
FROM usercourses uc
FULL OUTER JOIN userprogress up
  ON uc.user_id = up.user_id
  AND uc.course_id = up.course_id
LEFT JOIN users u
  ON u.id = COALESCE(uc.user_id, up.user_id)
LEFT JOIN courses c
  ON c.id = COALESCE(uc.course_id, up.course_id);
