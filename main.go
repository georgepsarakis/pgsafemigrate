package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"pgsafemigrate/cmd"
	"pgsafemigrate/reporter"
)

func main() {
	app := &cli.App{
		Name:        "pgsafemigrate",
		Version:     "v0.1.0",
		Description: "Analyzes and detects common errors in migration SQL statements",
		Commands: []*cli.Command{
			{
				Name:  "check",
				Usage: "Check SQL statements in migration files",
				Description: "The check sub-command will process each migration file separately and produce a report. " +
					"Exits with a non-zero exit code on failure. Migration file paths are given as positional arguments.",
				Flags: []cli.Flag{
					cmd.ExcludedRulesFlag(),
				},
				Action: func(ctx *cli.Context) error {
					// TODO: make reporter configurable
					return cmd.Check(ctx, ctx.Args().Slice(), ctx.StringSlice(cmd.ExcludedRulesFlag().Name), reporter.PlainText{})
				},
			},
			{
				Name:  "list-rules",
				Usage: "List available rules",
				Action: func(ctx *cli.Context) error {
					return cmd.ListRules(ctx)
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
