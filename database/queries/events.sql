-- name: EventsByAscOffsetLimit :many
SELECT * FROM events ORDER BY event_id ASC OFFSET $1 LIMIT $2;

-- name: EventsByDescOffsetLimit :many
SELECT * FROM events ORDER BY event_id DESC OFFSET $1 LIMIT $2;

-- name: CountEvents :one
SELECT count(*) FROM events;

-- name: InsertEvent :one
INSERT INTO events (name, description, tags, website, starts_at, ends_at, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: EventByUUID :one
SELECT * FROM events WHERE uuid = $1 LIMIT 1;

-- name: UpdateEvent :one
UPDATE events SET name = $1, description = $2, tags = $3, website = $4, starts_at = $5, ends_at = $6 WHERE event_id = $7 RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE event_id = $1;

-- name: UserEventsByAscOffsetLimit :many
SELECT * FROM events WHERE user_id = $1 ORDER BY event_id ASC OFFSET $2 LIMIT $3;

-- name: UserEventsByDescOffsetLimit :many
SELECT * FROM events WHERE user_id = $1 ORDER BY event_id DESC OFFSET $2 LIMIT $3;

-- name: CountUserEvents :one
SELECT count(*) FROM events WHERE user_id = $1;