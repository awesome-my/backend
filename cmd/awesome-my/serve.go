package main

import (
	"github.com/urfave/cli/v2"
)

func newServeCommand() *cli.Command {
	return &cli.Command{
		Name: "serve",
		Action: func(cliCtx *cli.Context) error {
			return nil
		},
	}
}
