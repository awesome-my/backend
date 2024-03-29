package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
	"github.com/gobuffalo/nulls"
	"github.com/gofrs/uuid"
)

type Project struct {
	Uuid        uuid.UUID    `json:"uuid"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	Repository  nulls.String `json:"repository"`
	Website     nulls.String `json:"website"`
	CreatedAt   time.Time    `json:"created_at"`
}

func ProjectFromDatabase(p database.Project) Project {
	return Project{
		Uuid:        p.Uuid,
		Name:        p.Name,
		Description: p.Description,
		Tags:        p.Tags,
		Repository:  p.Repository,
		Website:     p.Website,
		CreatedAt:   p.CreatedAt,
	}
}

func (p *Public) Projects(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := awesomemy.PageLimitOffsetFromRequest(r)

	var tags []string
	if r.URL.Query().Get("tags") != "" {
		tags = strings.Split(r.URL.Query().Get("tags"), ",")
	}

	orderBy := "desc"
	if r.URL.Query().Get("orderBy") == "asc" {
		orderBy = "asc"
	}

	var err error
	var projects []database.Project
	switch orderBy {
	case "asc":
		if len(tags) > 0 {
			projects, err = p.queries.ProjectsByTagsAscOffsetLimit(r.Context(), p.database, database.ProjectsByTagsAscOffsetLimitParams{
				Tags:   tags,
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		} else {
			projects, err = p.queries.ProjectsByAscOffsetLimit(r.Context(), p.database, database.ProjectsByAscOffsetLimitParams{
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		}
	case "desc":
		if len(tags) > 0 {
			projects, err = p.queries.ProjectsByTagsDescOffsetLimit(r.Context(), p.database, database.ProjectsByTagsDescOffsetLimitParams{
				Tags:   tags,
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		} else {
			projects, err = p.queries.ProjectsByDescOffsetLimit(r.Context(), p.database, database.ProjectsByDescOffsetLimitParams{
				Offset: int32(offset),
				Limit:  int32(limit),
			})
		}
	}
	if err != nil {
		p.logger.Error("could not fetch projects by limit offset", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch projects.",
		})
		return
	}

	var total int64
	if len(tags) > 0 {
		total, err = p.queries.CountProjectsByTags(r.Context(), p.database, tags)
	} else {
		total, err = p.queries.CountProjects(r.Context(), p.database)
	}
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
