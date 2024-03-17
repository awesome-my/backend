-- name: UserByGithubEmail :one
SELECT * FROM users WHERE github_email = $1 LIMIT 1;

-- name: UserByUUID :one
SELECT * FROM users WHERE uuid = $1 LIMIT 1;

-- name: InsertUser :one
INSERT INTO users (github_email) VALUES ($1) RETURNING *;