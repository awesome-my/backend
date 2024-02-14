package handler

import (
	"encoding/json"
	"net/http"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/gofrs/uuid"
)

type User struct {
	Uuid        uuid.UUID `json:"uuid"`
	GitHubEmail string    `json:"github_email"`
}

func UserFromDatabase(u database.User) User {
	return User{
		Uuid:        u.Uuid,
		GitHubEmail: u.GithubEmail,
	}
}

func (c *Client) Account(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	json.NewEncoder(w).Encode(map[string]any{
		"item": UserFromDatabase(authUser),
	})
}
