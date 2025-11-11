-- name: GetDummyItem :one
SELECT id, name FROM dummy WHERE id = $1;
