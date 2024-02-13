package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type Auth struct {
	logger         *slog.Logger
	config         awesomemy.Config
	database       *sql.DB
	queries        *database.Queries
	sessionManager *scs.SessionManager
}

func NewAuth(logger *slog.Logger, cfg awesomemy.Config, db *sql.DB, sm *scs.SessionManager) http.Handler {
	a := &Auth{
		logger:         logger,
		config:         cfg,
		database:       db,
		queries:        database.New(),
		sessionManager: sm,
	}

	r := chi.NewRouter()
	r.Route("/oauth2", func(r chi.Router) {
		r.Get("/", a.OAuth2)
		r.Get("/callback", a.OAuth2Callback)
	})

	return r
}

func (a *Auth) OAuth2(w http.ResponseWriter, r *http.Request) {
	if err := a.sessionManager.RenewToken(r.Context()); err != nil {
		a.logger.Error("could not renew request session token", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not renew the request session token.",
		})
		return
	}

	verifier := oauth2.GenerateVerifier()
	a.sessionManager.Put(r.Context(), "oauth2:verifier", verifier)

	oauth2Cfg := a.config.Authentication.OAuth2.OAuth2Config("github")
	http.Redirect(w, r, oauth2Cfg.AuthCodeURL("state", oauth2.S256ChallengeOption(verifier)), http.StatusTemporaryRedirect)
}

func (a *Auth) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	verifier := a.sessionManager.GetString(r.Context(), "oauth2:verifier")
	if verifier == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request is missing PKCE code.",
		})
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The request is missing GitHub OAuth2 code.",
		})
		return
	}

	oauth2Cfg := a.config.Authentication.OAuth2.OAuth2Config("github")
	token, err := oauth2Cfg.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "The GitHub OAuth2 token is invalid.",
		})
		return
	}

	githubEmail, err := githubOAuth2Email(r.Context(), oauth2Cfg, token)
	if err != nil {
		a.logger.Error("could not fetch github oauth2 account details", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch account details from GitHub.",
		})
		return
	}

	user, err := a.queries.UserByGithubEmail(r.Context(), a.database, githubEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = a.queries.InsertUser(r.Context(), a.database, githubEmail)
			if err != nil {
				a.logger.Error("could not insert user by github email", slog.Any("err", err))
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Could not insert user into database.",
				})
			}

			return
		}

		a.logger.Error("could not fetch user by github email", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not fetch user by GitHub email.",
		})
		return
	}

	if err := a.sessionManager.RenewToken(r.Context()); err != nil {
		a.logger.Error("could not renew request session token", slog.Any("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Could not renew the request session token.",
		})
		return
	}

	a.sessionManager.Put(r.Context(), "user:uuid", user.Uuid.String())

	http.Redirect(w, r, a.config.FrontendBaseURL, http.StatusTemporaryRedirect)
}
