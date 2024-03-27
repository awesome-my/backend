-- name: InsertTag :one
INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING *;