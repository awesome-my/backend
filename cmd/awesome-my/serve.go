package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/handler"
	"github.com/urfave/cli/v2"
)

func newServeCommand() *cli.Command {
	return &cli.Command{
		Name: "serve",
		Action: func(cliCtx *cli.Context) error {
			logger := awesomemy.MustContextValue[*slog.Logger](cliCtx.Context, awesomemy.CtxKeyLogger)
			cfg := awesomemy.MustContextValue[awesomemy.Config](cliCtx.Context, awesomemy.CtxKeyConfig)

			logger.Info("opening a connection to postgres database")
			db, err := sql.Open("postgres", cfg.Postgres.DSN())
			if err != nil {
				logger.Error("could not initialize postgres database", slog.Any("err", err))
				os.Exit(1)
			}
			defer func(db *sql.DB) {
				_ = db.Close()
			}(db)

			logger.Info("listening and serving http")
			srv := &http.Server{
				Addr:    cfg.Http.Address(),
				Handler: handler.New(logger, cfg, db),
			}
			if err := srv.ListenAndServe(); err != nil {
				logger.Error("could not listen and serve http", slog.Any("err", err))
				os.Exit(1)
			}

			return nil
		},
	}
}
