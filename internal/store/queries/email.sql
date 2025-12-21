-- name: AddFailedEmail :exec
INSERT INTO email_failures (error, template_params, template_name, email_name) VALUES ($1, $2, $3, $4);

-- name: GetFailedEmails :many
SELECT id, template_params, template_name, email_name, retries FROM email_failures WHERE retries > 0;

-- name: UpdateFailedEmail :exec
UPDATE email_failures SET retries = $1, error = $2; 

-- name: DeleteFailedEmail :exec
DELETE FROM email_failures WHERE id = $1::uuid;