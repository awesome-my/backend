package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gobuffalo/nulls"
	"github.com/google/go-github/v55/github"
	"google.golang.org/api/option"
	"log/slog"
	"net/http"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
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
	r.Route("/oauth2/{provider}", func(r chi.Router) {
		r.Get("/", a.OAuth2)
		r.Get("/callback", a.OAuth2Callback)
	})
	r.Post("/logout", a.Logout)

	return r
}

func (a *Auth) OAuth2(w http.ResponseWriter, r *http.Request) {
	if err := a.sessionManager.RenewToken(r.Context()); err != nil {
		a.logger.Error("could not renew request session token", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not renew the request session token.",
		})
		return
	}

	provider := chi.URLParam(r, "provider")
	verifier := oauth2.GenerateVerifier()
	a.sessionManager.Put(r.Context(), "oauth2:verifier", verifier)

	oauth2Cfg := a.config.Authentication.OAuth2.OAuth2Config(provider)
	if oauth2Cfg == nil {
		awesomemy.Render(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid provider was presented.",
		})
		return
	}

	http.Redirect(w, r, oauth2Cfg.AuthCodeURL("state", oauth2.S256ChallengeOption(verifier)), http.StatusTemporaryRedirect)
}

func (a *Auth) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	verifier := a.sessionManager.GetString(r.Context(), "oauth2:verifier")
	if verifier == "" {
		awesomemy.Render(w, http.StatusBadRequest, map[string]string{
			"message": "The request is missing OAuth2 PKCE code.",
		})
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		awesomemy.Render(w, http.StatusBadRequest, map[string]string{
			"message": "The request is missing OAuth2 code.",
		})
		return
	}

	provider := chi.URLParam(r, "provider")
	oauth2Cfg := a.config.Authentication.OAuth2.OAuth2Config(provider)
	if oauth2Cfg == nil {
		awesomemy.Render(w, http.StatusBadRequest, map[string]string{
			"message": "Invalid provider was presented.",
		})
		return
	}

	token, err := oauth2Cfg.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		awesomemy.Render(w, http.StatusBadRequest, map[string]string{
			"message": "The OAuth2 token is invalid.",
		})
		return
	}

	email, err := oauth2Email(r.Context(), provider, oauth2Cfg, token)
	if err != nil {
		a.logger.Error("could not fetch oauth2 account details", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not fetch OAuth2 account details.",
		})
		return
	}

	var user database.User
	switch provider {
	case "github":
		user, err = a.queries.UserByGithubEmail(r.Context(), a.database, nulls.NewString(email))
	case "google":
		user, err = a.queries.UserByGoogleEmail(r.Context(), a.database, nulls.NewString(email))
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			switch provider {
			case "github":
				user, err = a.queries.InsertUser(r.Context(), a.database, database.InsertUserParams{
					GithubEmail: nulls.NewString(email),
				})
			case "google":
				user, err = a.queries.InsertUser(r.Context(), a.database, database.InsertUserParams{
					GoogleEmail: nulls.NewString(email),
				})
			}
			if err != nil {
				a.logger.Error("could not insert user by oauth2 email", slog.Any("err", err))
				awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
					"message": "Could not insert user by OAuth2 email.",
				})
				return
			}
		} else {
			a.logger.Error("could not fetch user by oauth2 email", slog.Any("err", err))
			awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
				"message": "Could not fetch user by OAuth2 email.",
			})
			return
		}
	}

	if err := a.sessionManager.RenewToken(r.Context()); err != nil {
		a.logger.Error("could not renew request session token", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not renew the request session token.",
		})
		return
	}

	a.sessionManager.Put(r.Context(), "user:uuid", user.Uuid.String())

	http.Redirect(w, r, a.config.FrontendBaseURL, http.StatusTemporaryRedirect)
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if err := a.sessionManager.RenewToken(r.Context()); err != nil {
		a.logger.Error("could not renew request session token", slog.Any("err", err))
		awesomemy.Render(w, http.StatusInternalServerError, map[string]string{
			"message": "Could not renew the request session token.",
		})
		return
	}

	a.sessionManager.Put(r.Context(), "user:uuid", "")
}

func oauth2Email(ctx context.Context, provider string, cfg *oauth2.Config, token *oauth2.Token) (string, error) {
	switch provider {
	case "github":
		return githubOAuth2Email(ctx, cfg, token)
	case "google":
		return googleOAuth2Email(ctx, cfg, token)
	}

	return "", fmt.Errorf("invalid provider was provided")
}

func githubOAuth2Email(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, error) {
	githubEmails, _, err := github.NewClient(cfg.Client(ctx, token)).
		WithAuthToken(token.AccessToken).Users.ListEmails(ctx, &github.ListOptions{})
	if err != nil {
		return "", err
	}

	var primaryEmail string
	for _, ge := range githubEmails {
		if ge.GetPrimary() {
			primaryEmail = ge.GetEmail()
		}
	}

	return primaryEmail, nil
}

func googleOAuth2Email(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, error) {
	srv, err := googleoauth2.NewService(ctx, option.WithTokenSource(cfg.TokenSource(ctx, token)))
	if err != nil {
		return "", err
	}

	info, err := srv.Tokeninfo().Do()
	if err != nil {
		return "", err
	}

	return info.Email, nil
}
