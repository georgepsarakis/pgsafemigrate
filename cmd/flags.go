package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"pgsafemigrate/rules"
	"strings"
)

// ExcludedRulesFlag defines a --excluded-rules option for providing the rule aliases
// that will be ignored for all files.
func ExcludedRulesFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name: "excluded-rules",
		Action: func(cCtx *cli.Context, inputs []string) error {
			var allAliases []string
			for _, in := range inputs {
				allAliases = append(allAliases, strings.Split(in, ",")...)
			}
			for _, alias := range allAliases {
				if !rules.All().Contains(alias) {
					return cli.Exit(fmt.Sprintf("unknown alias %q", alias), 1)
				}
			}
			return nil
		},
	}
}
