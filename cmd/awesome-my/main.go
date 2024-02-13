package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/awesome-my/backend"
	"github.com/urfave/cli/v2"
)

func main() {
	(&cli.App{
		Name: "awesome-my",
		Before: func(cliCtx *cli.Context) error {
			level := slog.LevelInfo
			if cliCtx.Bool("debug") {
				level = slog.LevelDebug
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: level,
			}))

			cfg, err := awesomemy.ParseConfigFromFile("./config.yaml")
			if err != nil {
				logger.Error("could not parse configuration from file", slog.Any("err", err))
				os.Exit(1)
			}

			cliCtx.Context = context.WithValue(cliCtx.Context, awesomemy.CtxKeyLogger, logger)
			cliCtx.Context = context.WithValue(cliCtx.Context, awesomemy.CtxKeyConfig, cfg)

			return nil
		},
		Commands: []*cli.Command{
			newServeCommand(),
			newMigrateCommand(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "record level that will be logged.",
				Value: false,
			},
		},
	}).Run(os.Args)
}
