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

-- name: UpdateCourse :exec
UPDATE courses
SET title = $1, description = $2, completion_title = $3, completion_message = $4
WHERE id = $5;

-- name: UpsertCourseMaterial :exec
INSERT INTO course_materials (id, name, storage_key, position, course_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  storage_key = EXCLUDED.storage_key,
  position = EXCLUDED.position;

-- name: DeleteCourseMaterials :exec
DELETE FROM course_materials WHERE id = ANY($1::uuid[]);

-- name: UpdateVideoSection :exec
UPDATE videosections SET title = $1, storage_key = $2, position = $3 WHERE id = $4;

-- name: UpdateQuizSectionPosition :exec
UPDATE quizsections SET position = $1 WHERE id = $2;

-- name: UpsertQuizQuestion :exec
INSERT INTO quizquestions (id, question, position, is_multi_answer, quiz_section_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
  question = EXCLUDED.question,
  position = EXCLUDED.position,
  is_multi_answer = EXCLUDED.is_multi_answer;

-- name: UpsertQuizAnswer :exec
INSERT INTO quizanswers (id, answer, correct_answer, position, quiz_question_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
  answer = EXCLUDED.answer,
  correct_answer = EXCLUDED.correct_answer,
  position = EXCLUDED.position;

-- name: DeleteVideoSections :exec
DELETE FROM videosections WHERE id = ANY($1::uuid[]);

-- name: DeleteQuizSections :exec
DELETE FROM quizsections WHERE id = ANY($1::uuid[]);

-- name: DeleteQuizQuestions :exec
DELETE FROM quizquestions WHERE id = ANY($1::uuid[]);

-- name: DeleteQuizAnswers :exec
DELETE FROM quizanswers WHERE id = ANY($1::uuid[]);

-- name: RemoveDeletedSectionsFromProgress :exec
UPDATE userprogress
SET completed_section_ids = (
  SELECT COALESCE(array_agg(section_id), ARRAY[]::uuid[])
  FROM unnest(completed_section_ids) AS section_id
  WHERE section_id != ALL($1::uuid[])
)
WHERE completed_section_ids && $1::uuid[];

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
