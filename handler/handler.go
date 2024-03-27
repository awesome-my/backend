package handler

import (
	"database/sql"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/awesome-my/backend"
	"github.com/go-chi/httprate"
	"github.com/gomodule/redigo/redis"
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
			return redis.Dial("tcp", cfg.Redis.ConnectionString())
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
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		awesomemy.RenderNotFound(w)
	})
	r.Use(
		httprate.Limit(
			50,
			1*time.Minute,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				awesomemy.Render(w, http.StatusTooManyRequests, map[string]string{
					"message": "You have hit the rate limit, try again later.",
				})
			}),
		),
		sm.LoadAndSave,
		corsMiddleware(cfg),
	)
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
