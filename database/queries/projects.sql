-- name: ProjectsByAscOffsetLimit :many
SELECT * FROM projects ORDER BY project_id ASC OFFSET $1 LIMIT $2;

-- name: ProjectsByDescOffsetLimit :many
SELECT * FROM projects ORDER BY project_id DESC OFFSET $1 LIMIT $2;

-- name: CountProjects :one
SELECT count(*) FROM projects;

-- name: InsertProject :one
INSERT INTO projects (name, description, tags, repository, website, user_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: ProjectByUUID :one
SELECT * FROM projects WHERE uuid = $1 LIMIT 1;

-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, tags = $3, repository = $4, website = $5 WHERE project_id = $6 RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE project_id = $1;

-- name: UserProjectsByAscOffsetLimit :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY project_id ASC OFFSET $2 LIMIT $3;

-- name: UserProjectsByDescOffsetLimit :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY project_id DESC OFFSET $2 LIMIT $3;

-- name: CountUserProjects :one
SELECT count(*) FROM projects WHERE user_id = $1;

-- name: ProjectsByTagsAscOffsetLimit :many
SELECT * FROM projects WHERE tags && $1 ORDER BY project_id ASC OFFSET $2 LIMIT $3;

-- name: ProjectsByTagsDescOffsetLimit :many
SELECT * FROM projects WHERE tags && $1 ORDER BY project_id DESC OFFSET $2 LIMIT $3;

-- name: CountProjectsByTags :one
SELECT count(*) FROM projects WHERE tags && $1;