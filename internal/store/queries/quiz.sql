-- name: SaveQuizAttempt :exec
INSERT INTO quiz_attempts (user_id, quiz_id, answers, attempt_number)
VALUES (
  sqlc.arg('user_id'),
  sqlc.arg('quiz_id'),
  sqlc.arg('answers'),
  (SELECT COALESCE(MAX(attempt_number), 0) + 1 FROM quiz_attempts WHERE user_id = sqlc.arg('user_id') AND quiz_id = sqlc.arg('quiz_id'))
);

-- name: GetQuizAttemptsByUserID :many
SELECT
  qah.id,
  uqs.user_id,
  uqs.quiz_id,
  qah.answers,
  qah.attempt_number,
  uqs.attempts AS total_attempts
FROM user_quiz_state uqs
LEFT JOIN quiz_attempts qah ON qah.user_id = uqs.user_id AND qah.quiz_id = uqs.quiz_id
WHERE uqs.user_id = $1
ORDER BY uqs.quiz_id, qah.attempt_number;

-- name: IncrementAttempts :exec
INSERT INTO user_quiz_state (user_id, quiz_id, attempts)
VALUES (
  sqlc.arg('user_id'),
  sqlc.arg('quiz_id'),
  1
)
ON CONFLICT (user_id, quiz_id)
DO UPDATE SET attempts = user_quiz_state.attempts + 1;

-- name: GetCurrentQuizAnswersByUserID :many
SELECT quiz_id, quiz_answers FROM user_quiz_state WHERE user_id = $1;

-- name: GetAllQuizSections :many
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
GROUP BY qs.id, qs.position, qs.course_id
ORDER BY qs.course_id, qs.position;

-- name: DeleteUserQuizState :exec
DELETE FROM user_quiz_state WHERE user_id = $1 AND quiz_id = $2;

-- name: DeleteQuizAttempts :exec
DELETE FROM quiz_attempts WHERE user_id = $1 AND quiz_id = $2;

-- name: UpsertQuizState :exec
INSERT INTO user_quiz_state (user_id, quiz_id, quiz_answers)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, quiz_id)
DO UPDATE SET quiz_answers = EXCLUDED.quiz_answers;

-- name: GetQuizState :one
SELECT quiz_state, attempts FROM user_quiz_state WHERE user_id = $1 AND quiz_id = $2;

-- name: SetQuizState :exec
INSERT INTO user_quiz_state (user_id, quiz_id, quiz_state)
     VALUES ($1, $2, $3)
     ON CONFLICT (user_id, quiz_id)
     DO UPDATE SET quiz_state = EXCLUDED.quiz_state;
