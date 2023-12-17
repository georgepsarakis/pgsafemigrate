package cmd

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"pgsafemigrate/loader"
	"pgsafemigrate/reporter"
	"pgsafemigrate/rules"
	"strings"
	"text/tabwriter"
)

// Check processes the migration files at the given paths and produces a report.
// The returned error will signal a non-zero exit code for the CLI.
func Check(_ *cli.Context, paths []string, excludedRules []string, output reporter.Reporter) error {
	migrationFiles, err := loader.ReadStatementsFromFiles(paths...)
	if err != nil {
		return err
	}
	var failed bool
	for _, m := range migrationFiles {
		results, err := rules.ProcessMigration(m, excludedRules)
		if err != nil {
			panic(err)
		}
		var reports []reporter.Report
		for _, r := range results {
			failed = failed || len(r.Errors) > 0
			reports = append(reports, reporter.NewReport(m.Path, r.Errors))
		}

		fmt.Println(output.Print(reports))
	}
	if failed {
		return cli.Exit("\u274c Problems found.", 1)
	} else {
		fmt.Println("\u2713 No problems found!")
	}
	return nil
}

// ListRules list all available rules sorted by alias.
// The output includes the category and the documentation guide for each rule.
func ListRules(_ *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	toTitle := cases.Title(language.AmericanEnglish).String
	aliasToTitle := func(alias string) string {
		return toTitle(strings.ReplaceAll(rules.CategoryFromAlias(alias), "-", " "))
	}
	for _, rule := range rules.All().SortedSlice() {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n",
			aliasToTitle(rule.Alias()), rule.Alias(), rule.Documentation()); err != nil {
			return err
		}
	}
	return w.Flush()
}
