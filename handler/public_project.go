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

type Project struct {
	Uuid        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
}

func ProjectFromDatabase(p database.Project) Project {
	return Project{
		Uuid:        p.Uuid,
		Name:        p.Name,
		Description: p.Description,
		Tags:        p.Tags,
	}
}

func (p *Public) Projects(w http.ResponseWriter, r *http.Request) {
	page, offset := awesomemy.PageAndOffsetFromRequest(r)

	projects, err := p.queries.ProjectsByOffsetLimit(r.Context(), p.database, database.ProjectsByOffsetLimitParams{
		Offset: int32(offset),
		Limit:  10,
	})
	if err != nil {
		p.logger.Error("could not fetch projects by limit offset", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch projects.",
		})
		return
	}

	total, err := p.queries.CountProjects(r.Context(), p.database)
	if err != nil {
		p.logger.Error("could not fetch projects count", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch projects count.",
		})
		return
	}

	apiProjects := make([]Project, len(projects))
	for i, p := range projects {
		apiProjects[i] = ProjectFromDatabase(p)
	}

	json.NewEncoder(w).Encode(map[string]any{
		"items":      apiProjects,
		"pagination": awesomemy.NewPaginationMeta(page, len(projects), int(total)),
	})
}

func (p *Public) Project(w http.ResponseWriter, r *http.Request) {
	projectUuid, err := uuid.FromString(chi.URLParam(r, "project"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The resource you are looking for could not be found.",
		})
		return
	}

	project, err := p.queries.ProjectByUUID(r.Context(), p.database, projectUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "The resource you are looking for could not be found.",
			})
			return
		}

		p.logger.Error("could not fetch project by uuid", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch project.",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"item": ProjectFromDatabase(project),
	})
}
