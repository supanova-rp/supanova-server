-- name: GetCourse :one
SELECT
  id,
  title,
  description,
  completion_title,
  completion_message
FROM courses WHERE id = $1;

-- name: AddCourse :one
INSERT INTO courses (title, description) VALUES ($1, $2) RETURNING id;

-- name: GetProgress :one
SELECT completed_intro, completed_section_ids FROM userprogress WHERE user_id = $1 AND course_id = $2;

-- name: IsUserEnrolledInCourse :one
SELECT EXISTS(SELECT 1 FROM usercourses WHERE user_id = $1 AND course_id = $2);
