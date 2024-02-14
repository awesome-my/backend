package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
)

func (c *Client) StoreProject(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	var data struct {
		Name        string   `json:"name" validate:"required,min=8,max=191"`
		Description string   `json:"description" validate:"required,min=8,max=512"`
		Tags        []string `json:"tags" validate:"dive,min=4,max=12"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	if err := c.validator.StructCtx(r.Context(), data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request body is in malformed format.",
		})
		return
	}

	project, err := c.queries.InsertProject(r.Context(), c.database, database.InsertProjectParams{
		Name:        data.Name,
		Description: data.Description,
		Tags:        data.Tags,
		UserID:      authUser.UserID,
	})
	if err != nil {
		c.logger.Error("could not insert project", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not insert project into database.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": ProjectFromDatabase(project),
	})
}
