// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: tags.sql

package database

import (
	"context"

	"github.com/gobuffalo/nulls"
)

const insertTag = `-- name: InsertTag :one
INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING tag_id, uuid, name, slug
`

type InsertTagParams struct {
	Name string
	Slug nulls.String
}

func (q *Queries) InsertTag(ctx context.Context, db DBTX, arg InsertTagParams) (Tag, error) {
	row := db.QueryRowContext(ctx, insertTag, arg.Name, arg.Slug)
	var i Tag
	err := row.Scan(
		&i.TagID,
		&i.Uuid,
		&i.Name,
		&i.Slug,
	)
	return i, err
}
