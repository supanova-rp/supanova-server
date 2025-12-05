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

-- name: GetCourseMaterials :many
SELECT
  id, name, position, storage_key
FROM course_materials
WHERE course_id = $1
ORDER BY position;

-- name: GetCourseVideoSections :many
SELECT
  id, title, position, storage_key
FROM videosections
WHERE course_id = $1
ORDER BY position;

-- name: GetCourseQuizSections :many
SELECT
  qs.id,
  qs.position,
  qs.course_id,
  json_agg(
    json_build_object(
      'id', qq.id,
      'question', qq.question,
      'position', qq.position,
      'is_multi_answer', qq.is_multi_answer,
      'answers', (
        SELECT json_agg(
          json_build_object(
            'id', qa.id,
            'answer', qa.answer,
            'correct_answer', qa.correct_answer,
            'position', qa.position
          ) ORDER BY qa.position
        )
        FROM quizanswers qa
        WHERE qa.quiz_question_id = qq.id
      )
    ) ORDER BY qq.position
  ) AS questions
FROM quizsections qs
LEFT JOIN quizquestions qq ON qq.quiz_section_id = qs.id
WHERE qs.course_id = $1
GROUP BY qs.id, qs.position, qs.course_id
ORDER BY qs.position;

-- Insert section_id into completed_section_ids array if no entry exists
-- or append section_id to the existing array if it's not already present
-- name: UpdateProgress :exec
INSERT INTO userprogress (user_id, course_id, completed_section_ids)
VALUES (sqlc.arg('user_id'), sqlc.arg('course_id'), ARRAY[sqlc.arg('section_id')::uuid])
ON CONFLICT (user_id, course_id)
DO UPDATE SET completed_section_ids = array_append(userprogress.completed_section_ids, sqlc.arg('section_id')::uuid)
WHERE NOT (sqlc.arg('section_id') = ANY(userprogress.completed_section_ids));
