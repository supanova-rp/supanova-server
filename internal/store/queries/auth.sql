-- name: InsertUser :one
INSERT INTO users (id, name, email) VALUES ($1, $2, $3) RETURNING id;