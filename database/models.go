// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0

package database

import (
	"github.com/gofrs/uuid"
)

type User struct {
	UserID      int32
	Uuid        uuid.UUID
	GithubEmail string
}