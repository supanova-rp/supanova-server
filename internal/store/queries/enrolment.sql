-- name: IsUserEnrolledInCourse :one
SELECT EXISTS(SELECT 1 FROM usercourses WHERE user_id = $1 AND course_id = $2);

-- name: EnrolInCourse :exec
INSERT INTO usercourses (user_id, course_id) VALUES ($1, $2);

-- name: DisenrolInCourse :exec
DELETE FROM usercourses WHERE user_id = $1 AND course_id = $2;