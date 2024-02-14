package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"slices"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/awesome-my/backend"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

func New(logger *slog.Logger, cfg awesomemy.Config, db *sql.DB) http.Handler {
	sameSite := http.SameSiteLaxMode
	switch cfg.Authentication.Session.SameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	sm := scs.New()
	sm.Store = redisstore.NewWithPrefix(&redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.Redis.ConnnectionString())
		},
	}, cfg.Authentication.Session.Prefix)
	sm.Lifetime = cfg.Authentication.Session.Lifetime.Duration
	sm.Cookie = scs.SessionCookie{
		Name:     cfg.Authentication.Session.Name,
		Persist:  cfg.Authentication.Session.Persist,
		SameSite: sameSite,
		Secure:   cfg.Authentication.Session.Secure,
		Path:     "/",
	}

	r := chi.NewRouter()
	r.Use(sm.LoadAndSave, corsMiddleware(cfg))
	r.Mount("/public", NewPublic(logger, cfg, db))
	r.Mount("/auth", NewAuth(logger, cfg, db, sm))
	r.Mount("/client", NewClient(logger, cfg, db, sm))

	return r
}

func corsMiddleware(cfg awesomemy.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if slices.Contains(cfg.Http.Cors.Origin, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "7200")

			if r.Method == http.MethodOptions {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
