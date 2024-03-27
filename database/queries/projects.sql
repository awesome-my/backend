-- name: FilteredProjectsByOffsetLimit :many
SELECT * FROM projects
    WHERE projects.name ILIKE $1 OR projects.slug ILIKE $2
    OFFSET $3 LIMIT $4;

-- name: FilteredProjectsByDescOffsetLimit :many
SELECT * FROM projects
WHERE projects.name ILIKE $1 OR projects.slug ILIKE $2
ORDER BY project_id DESC
OFFSET $3 LIMIT $4;

-- name: CountFilteredProjects :one
SELECT count(*) FROM projects WHERE projects.name ILIKE $1 OR projects.slug ILIKE $2;

-- name: ProjectsByAscOffsetLimit :many
SELECT * FROM projects ORDER BY project_id OFFSET $1 LIMIT $2;

-- name: ProjectsByDescOffsetLimit :many
SELECT * FROM projects ORDER BY project_id DESC OFFSET $1 LIMIT $2;

-- name: CountProjects :one
SELECT count(*) FROM projects;

-- name: InsertProject :one
INSERT INTO projects (name, description, repository, website, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ProjectByUUID :one
SELECT * FROM projects WHERE uuid = $1 LIMIT 1;

-- name: UpdateProject :one
UPDATE projects SET name = $1, description = $2, repository = $3, website = $4 WHERE project_id = $5 RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE project_id = $1;

-- name: UserProjectsByAscOffsetLimit :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY project_id ASC OFFSET $2 LIMIT $3;

-- name: UserProjectsByDescOffsetLimit :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY project_id DESC OFFSET $2 LIMIT $3;

-- name: CountUserProjects :one
SELECT count(*) FROM projects WHERE user_id = $1;