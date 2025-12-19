-- name: SetFailedEmail :exec
INSERT INTO email_failures (error, params) VALUES ($1, $2);