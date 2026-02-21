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

-- name: GetQuizStatesByUserID :many
SELECT quiz_id, quiz_state_v2 FROM user_quiz_state WHERE user_id = $1;

-- name: UpsertQuizState :exec
INSERT INTO user_quiz_state (user_id, quiz_id, quiz_state_v2)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, quiz_id)
DO UPDATE SET quiz_state_v2 = EXCLUDED.quiz_state_v2;
