-- name: AddEmailFailure :exec
INSERT INTO email_failures (error, template_params, template_name) VALUES ($1, $2, $3);

-- name: GetEmailFailures :many
SELECT id, template_params, template_name, email_name FROM email_failures WHERE retries > 0;

-- name: UpdateEmailFailure :exec
UPDATE email_failures SET updated_at = $1, retries = $2, error = $3; 

-- name: DeleteEmailFailures :exec
DELETE FROM email_failures WHERE id = ANY($1::uuid[]) OR retries <= 0;