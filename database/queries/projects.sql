-- name: ProjectsByOffsetLimit :many
SELECT * FROM projects OFFSET $1 LIMIT $2;

-- name: CountProjects :one
SELECT count(*) FROM projects;

-- name: InsertProject :one
INSERT INTO projects (name, description, tags, user_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ProjectByUUID :one
SELECT * FROM projects WHERE uuid = $1 LIMIT 1;

-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, tags = $3 WHERE project_id = $4 RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE project_id = $1;