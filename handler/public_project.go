package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
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
	Repository  nulls.String `json:"repository"`
	Website     nulls.String `json:"website"`
	CreatedAt   time.Time    `json:"created_at"`
}

func ProjectFromDatabase(p database.Project) Project {
	return Project{
		Uuid:        p.Uuid,
		Name:        p.Name,
		Description: p.Description,
		Repository:  p.Repository,
		Website:     p.Website,
		CreatedAt:   p.CreatedAt,
	}
}

func (p *Public) Projects(w http.ResponseWriter, r *http.Request) {
	page, limit, offset := awesomemy.PageLimitOffsetFromRequest(r)

	//var tags []string
	//if r.URL.Query().Get("tags") != "" {
	//	tags = strings.Split(r.URL.Query().Get("tags"), ",")
	//}

	keyword := r.URL.Query().Get("keyword")

	orderBy := "desc"
	if r.URL.Query().Get("orderBy") == "asc" {
		orderBy = "asc"
	}

	var err error
	var projects []database.Project
	switch orderBy {
	case "asc":
		projects, err = p.queries.FilteredProjectsByOffsetLimit(r.Context(), p.database, database.FilteredProjectsByOffsetLimitParams{
			Name:   keyword,
			Slug:   nulls.NewString(keyword),
			Offset: int32(offset),
			Limit:  int32(limit),
		})
	case "desc":
		projects, err = p.queries.FilteredProjectsByDescOffsetLimit(r.Context(), p.database, database.FilteredProjectsByDescOffsetLimitParams{
			Name:   keyword,
			Slug:   nulls.NewString(keyword),
			Offset: int32(offset),
			Limit:  int32(limit),
		})
	}
	if err != nil {
		p.logger.Error("could not fetch projects by limit offset", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not fetch projects.",
		})
		return
	}

	total, err := p.queries.CountProjects(r.Context(), p.database)
	if err != nil {
		p.logger.Error("could not fetch projects count", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not fetch projects count.",
		})
		return
	}

	apiProjects := make([]Project, len(projects))
	for i, p := range projects {
		apiProjects[i] = ProjectFromDatabase(p)
	}

	awesomemy.Render(w, http.StatusOK, map[string]any{
		"items":      apiProjects,
		"pagination": awesomemy.NewPaginationMeta(page, len(projects), int(total)),
	})
}

func (p *Public) Project(w http.ResponseWriter, r *http.Request) {
	projectUuid, err := uuid.FromString(chi.URLParam(r, "project"))
	if err != nil {
		awesomemy.RenderNotFound(w)
		return
	}

	project, err := p.queries.ProjectByUUID(r.Context(), p.database, projectUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			awesomemy.RenderNotFound(w)
			return
		}

		p.logger.Error("could not fetch project by uuid", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not fetch project.",
		})
		return
	}

	awesomemy.Render(w, http.StatusOK, map[string]any{
		"item": ProjectFromDatabase(project),
	})
}
