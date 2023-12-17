package loader

import (
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgsafemigrate/annotations"
	"testing"
)

func TestComment_IsDown(t *testing.T) {
	type fields struct {
		Content              string
		TokenIndex           int
		SQLMigrateAnnotation bool
		SQLMigrateDirection  migrate.MigrationDirection
		NoLintAnnotation     annotations.NoLint
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "when a sql-migrate command is found and marks the migration direction as down",
			fields: fields{
				SQLMigrateAnnotation: true,
				SQLMigrateDirection:  migrate.Down,
			},
			want: true,
		},
		{
			name: "when a sql-migrate command is found and marks the migration direction as up",
			fields: fields{
				SQLMigrateAnnotation: true,
				SQLMigrateDirection:  migrate.Up,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Comment{
				Content:              tt.fields.Content,
				TokenIndex:           tt.fields.TokenIndex,
				SQLMigrateAnnotation: tt.fields.SQLMigrateAnnotation,
				SQLMigrateDirection:  tt.fields.SQLMigrateDirection,
				NoLintAnnotation:     tt.fields.NoLintAnnotation,
			}
			if got := c.IsDown(); got != tt.want {
				t.Errorf("IsDown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadMigration(t *testing.T) {
	t.Run("only up migration statements", func(t *testing.T) {
		m, err := LoadMigration(`
-- +migrate Up
SELECT 1;
UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;
-- +migrate Down
-- nothing to downgrade`)

		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"SELECT 1;\n",
				`UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;` + "\n"}, m.UpStatements)
		assert.Empty(t, m.DownStatements)
	})

	t.Run("both up & down migration statements", func(t *testing.T) {
		m, err := LoadMigration(`
-- +migrate Up
SELECT 1;
UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;
-- +migrate Down
UPDATE "movies" SET updated_at = NULL;
`)

		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"SELECT 1;\n",
				`UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;` + "\n"}, m.UpStatements)
		assert.Equal(t, []string{`UPDATE "movies" SET updated_at = NULL;` + "\n"}, m.DownStatements)
	})

	t.Run("not an sql-migrate-formatted file", func(t *testing.T) {
		m, err := LoadMigration(`
		SELECT 1;
		UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;`)

		require.NoError(t, err)
		assert.Equal(t,
			[]string{
				"SELECT 1;",
				`UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP;`}, m.UpStatements)
		assert.Empty(t, m.DownStatements)
	})
}

func TestParseStatements(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "test 1",
			input: `SELECT 1; SELECT 2; -- test comment
SELECT 3; UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP WHERE id>10;`,
			want: []string{
				"SELECT 1;", "SELECT 2;", "SELECT 3;",
				`UPDATE "movies" SET updated_at = CURRENT_TIMESTAMP WHERE id>10;`,
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements, err := ParseStatements(tt.input)
			tt.wantErr(t, err)
			require.Equal(t, tt.want, statements)
		})
	}
}
