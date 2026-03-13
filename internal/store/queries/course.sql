-- name: GetCourse :one
SELECT
  c.id,
  c.title,
  c.description,
  c.completion_title,
  c.completion_message,
  (
    SELECT json_agg(json_build_object(
      'id', v.id,
      'title', v.title,
      'storage_key', v.storage_key,
      'position', v.position
    ) ORDER BY v.position)
    FROM videosections v WHERE v.course_id = c.id
  ) AS video_sections,
  (
    SELECT json_agg(json_build_object(
      'id', qs.id,
      'position', qs.position,
      'questions', (
        SELECT json_agg(
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
        )
        FROM quizquestions qq
        WHERE qq.quiz_section_id = qs.id
      )
    ) ORDER BY qs.position)
    FROM quizsections qs WHERE qs.course_id = c.id
  ) AS quiz_sections,
  (
    SELECT json_agg(json_build_object(
      'id', m.id,
      'name', m.name,
      'storage_key', m.storage_key,
      'position', m.position
    ) ORDER BY m.position)
    FROM course_materials m WHERE m.course_id = c.id
  ) AS materials
FROM courses c
WHERE c.id = $1;

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

-- name: GetAssignedCourseTitles :many
SELECT c.id, c.title, c.description
FROM courses c
INNER JOIN usercourses uc ON uc.course_id = c.id
WHERE uc.user_id = $1
ORDER BY c.title;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;

-- name: GetAllCourses :many
SELECT
  c.id,
  c.title,
  c.description,
  c.completion_title,
  c.completion_message,
  (
    SELECT json_agg(json_build_object(
      'id', v.id,
      'title', v.title,
      'storage_key', v.storage_key,
      'position', v.position
    ) ORDER BY v.position)
    FROM videosections v WHERE v.course_id = c.id
  ) AS video_sections,
  (
    SELECT json_agg(json_build_object(
      'id', q.id,
      'position', q.position
    ) ORDER BY q.position)
    FROM quizsections q WHERE q.course_id = c.id
  ) AS quiz_sections,
  (
    SELECT json_agg(json_build_object(
      'id', m.id,
      'name', m.name,
      'storage_key', m.storage_key,
      'position', m.position
    ) ORDER BY m.position)
    FROM course_materials m WHERE m.course_id = c.id
  ) AS materials
FROM courses c
ORDER BY c.title;
