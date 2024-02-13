package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/awesome-my/backend"
	"github.com/awesome-my/backend/database"
	"github.com/urfave/cli/v2"
)

func newMigrateCommand() *cli.Command {
	return &cli.Command{
		Name: "migrate",
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

			logger.Info("migrating postgres database")
			if err := database.Migrate(db); err != nil {
				logger.Error("could not migrate postgres database", slog.Any("err", err))
				os.Exit(1)
			}

			return nil
		},
	}
}
