package rules

import (
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"pgsafemigrate/loader"
	"testing"
)

func TestProcessMigration(t *testing.T) {
	type args struct {
		migrationFile loader.MigrationFile
		excludedRules []string
	}
	tests := []struct {
		name    string
		args    args
		want    []StatementResult
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "not sql-migrate formatted without violations",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;
COMMENT ON COLUMN movies.released_at IS 'First release date';`,
				},
			},
			want: []StatementResult{
				{
					Passed:    true,
					Direction: migrate.Up,
				},
				{
					Passed:    true,
					Direction: migrate.Up,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "sql-migrate formatted without violations",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `
-- +migrate Up
ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;
COMMENT ON COLUMN movies.released_at IS 'First release date';

-- +migrate Down
ALTER TABLE movies DROP COLUMN released_at;
`,
				},
			},
			want: []StatementResult{
				{
					Passed:    true,
					Direction: migrate.Up,
				},
				{
					Passed:    true,
					Direction: migrate.Up,
				},
				{
					Passed:    true,
					Direction: migrate.Down,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid SQL",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `
-- +migrate Up
ALTER TABLE movies ADD COLUMN TIMESTAMP;
COMMENT ON COLUMN movies.released_at IS 'First release date';

-- +migrate Down
ALTER TABLE movies DROP COLUMN released_at;
`,
				},
			},
			want: []StatementResult{
				{
					Passed:    false,
					Direction: migrate.Up,
					Errors:    []ReportedError{ParseError{message: "syntax error at or near \";\"", statement: "ALTER TABLE movies ADD COLUMN TIMESTAMP;\n"}},
				},
				{
					Passed:    true,
					Direction: migrate.Up,
				},
				{
					Passed:    true,
					Direction: migrate.Down,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "statements with violations",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `
CREATE INDEX test_idx ON movies(title);
`,
				},
			},
			want: []StatementResult{
				{
					Passed:    false,
					Direction: migrate.Up,
					Errors: []ReportedError{
						Violation{
							rule:      All()["high-availability-avoid-non-concurrent-index-creation"],
							statement: "CREATE INDEX test_idx ON movies(title);",
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "statements with violations with matching no-lint annotation",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `-- pgsafemigrate:nolint:high-availability-avoid-non-concurrent-index-creation
CREATE INDEX test_idx ON movies(title);
`,
				},
			},
			want: []StatementResult{
				{
					Passed:    true,
					Direction: migrate.Up,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "statements with violations with not matching no-lint annotation",
			args: args{
				migrationFile: loader.MigrationFile{
					Path: "test1.sql",
					Contents: `-- pgsafemigrate:nolint:high-availability-another-rule
CREATE INDEX test_idx ON movies(title);
`,
				},
			},
			want: []StatementResult{
				{
					Passed:    false,
					Direction: migrate.Up,
					Errors: []ReportedError{
						Violation{
							rule:      All()["high-availability-avoid-non-concurrent-index-creation"],
							statement: "CREATE INDEX test_idx ON movies(title);",
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ProcessMigration(tt.args.migrationFile, tt.args.excludedRules)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
