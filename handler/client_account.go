package handler

import (
	"github.com/gobuffalo/nulls"
	"net/http"
	"time"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/gofrs/uuid"
)

type User struct {
	Uuid        uuid.UUID    `json:"uuid"`
	GitHubEmail nulls.String `json:"github_email"`
	GoogleEmail nulls.String `json:"google_email"`
	CreatedAt   time.Time    `json:"created_at"`
}

func UserFromDatabase(u database.User) User {
	return User{
		Uuid:        u.Uuid,
		GitHubEmail: u.GithubEmail,
		GoogleEmail: u.GoogleEmail,
		CreatedAt:   u.CreatedAt,
	}
}

func (c *Client) Account(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	awesomemy.Render(w, http.StatusOK, map[string]any{
		"item": UserFromDatabase(authUser),
	})
}
