package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
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

func (c *Client) UpdateProject(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	projectUuid, err := uuid.FromString(chi.URLParam(r, "project"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	project, err := c.queries.ProjectByUUID(r.Context(), c.database, projectUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		c.logger.Error("could not fetch project by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch project.",
		})
		return
	}

	if project.UserID != authUser.UserID {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

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

	project, err = c.queries.UpdateProject(r.Context(), c.database, database.UpdateProjectParams{
		Name:        data.Name,
		Description: data.Description,
		Tags:        data.Tags,
	})
	if err != nil {
		c.logger.Error("could not update project", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not update project.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": ProjectFromDatabase(project),
	})
}

func (c *Client) DeleteProject(w http.ResponseWriter, r *http.Request) {
	authUser := awesomemy.MustContextValue[database.User](r.Context(), awesomemy.CtxKeyAuthUser)

	projectUuid, err := uuid.FromString(chi.URLParam(r, "project"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	project, err := c.queries.ProjectByUUID(r.Context(), c.database, projectUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		c.logger.Error("could not fetch project by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch project.",
		})
		return
	}

	if project.UserID != authUser.UserID {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	if err := c.queries.DeleteProject(r.Context(), c.database, project.ProjectID); err != nil {
		c.logger.Error("could not delete project", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not delete project.",
		})
		return
	}
}
