package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	migrate "github.com/rubenv/sql-migrate"
	"pgsafemigrate/annotations"
	"pgsafemigrate/loader"
)

type MigrationContext struct {
	AllStatements []*pg_query.Node
	Direction     migrate.MigrationDirection
	FilePath      string
	InTransaction bool
	RawSQL        string
}

func ProcessMigration(migrationFile loader.MigrationFile, excludedRules []string) ([]StatementResult, error) {
	migration, err := loader.LoadMigration(migrationFile.Contents)
	if err != nil {
		return nil, err
	}

	nl, err := noLint(migrationFile.Contents)
	if err != nil {
		panic(err)
	}

	var results []StatementResult
	upRules := All().Except(append(excludedRules, nl[migrate.Up].RuleNames...)...)
	upResults, err := upRules.ProcessAll(MigrationContext{
		InTransaction: !migration.DisableTransactionUp,
		Direction:     migrate.Up,
		FilePath:      migrationFile.Path,
	}, migration.UpStatements)
	if err != nil {
		return nil, err
	}
	results = append(results, upResults...)

	downRules := All().Except(append(excludedRules, nl[migrate.Down].RuleNames...)...)
	downResults, err := downRules.ProcessAll(MigrationContext{
		InTransaction: !migration.DisableTransactionDown,
		Direction:     migrate.Down,
		FilePath:      migrationFile.Path,
	}, migration.DownStatements)
	if err != nil {
		return nil, err
	}

	results = append(results, downResults...)
	return results, nil
}

func noLint(migrationFileContents string) (map[migrate.MigrationDirection]annotations.NoLint, error) {
	comments, err := loader.ScanCommentsFromString(migrationFileContents)
	if err != nil {
		panic(err)
	}
	a := make(map[migrate.MigrationDirection]annotations.NoLint, 0)
	for _, c := range comments {
		if c.NoLintAnnotation.Valid {
			a[c.SQLMigrateDirection] = c.NoLintAnnotation
		}
	}
	return a, nil
}
