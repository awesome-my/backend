package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
)

type Public struct {
	logger   *slog.Logger
	config   awesomemy.Config
	database *sql.DB
	queries  *database.Queries
}

func NewPublic(logger *slog.Logger, cfg awesomemy.Config, db *sql.DB) http.Handler {
	p := &Public{
		logger:   logger,
		config:   cfg,
		database: db,
		queries:  database.New(),
	}

	r := chi.NewRouter()
	r.Route("/projects", func(r chi.Router) {
		r.Get("/", p.Projects)
		r.Get("/{project}", p.Project)
	})

	return r
}
