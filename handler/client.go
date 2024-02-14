package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
)

type Client struct {
	logger         *slog.Logger
	config         awesomemy.Config
	database       *sql.DB
	queries        *database.Queries
	sessionManager *scs.SessionManager
	validator      *validator.Validate
}

func NewClient(logger *slog.Logger, cfg awesomemy.Config, db *sql.DB, sm *scs.SessionManager) http.Handler {
	c := &Client{
		logger:         logger,
		config:         cfg,
		database:       db,
		queries:        database.New(),
		sessionManager: sm,
		validator:      validator.New(),
	}

	r := chi.NewRouter()
	r.Use(c.AuthenticateUser)
	r.Route("/account", func(r chi.Router) {
		r.Get("/", c.Account)
	})
	r.Route("/projects", func(r chi.Router) {
		r.Post("/", c.StoreProject)
	})

	return r
}

func (c *Client) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userUuid, err := uuid.FromString(c.sessionManager.GetString(r.Context(), "user:uuid"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "You are not authorized to access this resource.",
			})
			return
		}

		user, err := c.queries.UserByUUID(r.Context(), c.database, userUuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "You are not authorized to access this resource.",
				})
				return
			}

			c.logger.Error("could not fetch user by uuid", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Could not fetch user.",
			})
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), awesomemy.CtxKeyAuthUser, user)))
	})
}
