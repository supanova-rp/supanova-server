-- name: GetCourse :one
SELECT
  id,
  title,
  description,
  completion_title,
  completion_message
FROM courses WHERE id = $1;

-- name: GetCoursesOverview :many
SELECT id, title, description FROM courses ORDER BY title;

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

-- name: AddCourse :one
INSERT INTO courses (title, description, completion_title, completion_message)
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: InsertCourseMaterial :exec
INSERT INTO course_materials (id, name, storage_key, position, course_id)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertVideoSection :exec
INSERT INTO videosections (title, storage_key, position, course_id)
VALUES ($1, $2, $3, $4);

-- name: InsertQuizSection :one
INSERT INTO quizsections (position, course_id) VALUES ($1, $2) RETURNING id;

-- name: InsertQuizQuestion :one
INSERT INTO quizquestions (question, position, is_multi_answer, quiz_section_id)
VALUES ($1, $2, $3, $4) RETURNING id;

-- name: InsertQuizAnswer :exec
INSERT INTO quizanswers (answer, correct_answer, position, quiz_question_id)
VALUES ($1, $2, $3, $4);
