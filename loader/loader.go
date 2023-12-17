package loader

import (
	"bytes"
	"fmt"
	pg_query "github.com/pganalyze/pg_query_go/v4"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/rubenv/sql-migrate/sqlparse"
	"os"
	"path/filepath"
	"pgsafemigrate/annotations"
	"strings"
)

func LoadMigration(sql string) (*sqlparse.ParsedMigration, error) {
	m, err := sqlparse.ParseMigration(bytes.NewReader([]byte(sql)))
	if err != nil {
		if strings.Contains(err.Error(), "no Up/Down annotations found") {
			statements, err := ParseStatements(sql)
			if err != nil {
				return nil, err
			}
			return &sqlparse.ParsedMigration{
				UpStatements: statements,
			}, nil
		}
		return nil, err
	}
	return m, nil
}

// ParseStatements splits a multi-statement SQL script to individual statements.
func ParseStatements(rawSQL string) ([]string, error) {
	var statements []string
	tree, err := pg_query.Parse(rawSQL)
	if err != nil {
		return nil, err
	}
	for _, s := range tree.GetStmts() {
		stmt := rawSQL[s.StmtLocation : s.StmtLocation+s.StmtLen+1]
		scanRes, err := pg_query.Scan(stmt)
		if err != nil {
			return nil, err
		}
		for _, token := range scanRes.GetTokens() {
			if token.Token == pg_query.Token_SQL_COMMENT {
				stmt = stmt[:token.GetStart()] + stmt[token.GetEnd():]
				break
			}
		}
		stmt = strings.TrimSpace(stmt)
		// ensure statement is valid after removing comment
		if _, err := pg_query.Parse(stmt); err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}
	return statements, nil
}

type MigrationFile struct {
	Contents string
	Path     string
}

func ReadStatementsFromFiles(paths ...string) ([]MigrationFile, error) {
	var statements []MigrationFile
	for _, p := range paths {
		p, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if fileInfo.IsDir() {
			return nil, fmt.Errorf("expected file and %s is a directory", p)
		}
		sql, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		statements = append(statements, MigrationFile{
			Contents: string(sql),
			Path:     p,
		})
	}
	return statements, nil
}

type Comment struct {
	Content              string
	TokenIndex           int
	SQLMigrateAnnotation bool
	SQLMigrateDirection  migrate.MigrationDirection
	NoLintAnnotation     annotations.NoLint
}

func (c Comment) IsDown() bool {
	return c.SQLMigrateAnnotation && c.SQLMigrateDirection == migrate.Down
}

func ScanCommentsFromString(sql string) ([]Comment, error) {
	scanRes, err := pg_query.Scan(sql)
	if err != nil {
		return nil, err
	}

	var (
		comments         []Comment
		currentDirection migrate.MigrationDirection
	)

	for i, token := range scanRes.Tokens {
		if token.Token != pg_query.Token_SQL_COMMENT {
			continue
		}
		c := Comment{
			Content:    sql[token.GetStart()+3 : token.GetEnd()],
			TokenIndex: i + 1,
		}
		c.SQLMigrateAnnotation = strings.HasPrefix(c.Content, "+migrate")
		if strings.HasPrefix(c.Content, "+migrate Up") {
			c.SQLMigrateDirection = migrate.Up
			currentDirection = migrate.Up
		} else if strings.HasPrefix(c.Content, "+migrate Down") {
			c.SQLMigrateDirection = migrate.Down
			currentDirection = migrate.Down
		}
		c.SQLMigrateDirection = currentDirection
		if c.SQLMigrateAnnotation {
			comments = append(comments, c)
			continue
		}
		c.NoLintAnnotation = annotations.Parse(c.Content)
		comments = append(comments, c)

	}
	// Validate positions, nolint annotations must come immediately after migration commands
	if len(comments) == 0 {
		return comments, nil
	}
	if comments[0].SQLMigrateAnnotation {
		if len(comments) > 1 {
			if !checkMigrateNoLintAreSequential(comments[0], comments[1]) {
				return nil, fmt.Errorf("nolint annotation found at line %d", comments[1].TokenIndex)
			}
		}
	}

	for i, c := range comments {
		if c.IsDown() {
			if i < len(comments)-1 {
				if !checkMigrateNoLintAreSequential(c, comments[i+1]) {
					return nil, fmt.Errorf("nolint annotation found at line %d", comments[i+1].TokenIndex)
				}
			}
		}
	}

	return comments, nil
}

func checkMigrateNoLintAreSequential(current, next Comment) bool {
	if !next.NoLintAnnotation.Valid {
		return true
	}
	return next.TokenIndex-current.TokenIndex == 1
}
