package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateIndexNonConcurrently_Process(t *testing.T) {
	t.Parallel()

	type args struct {
		statement     string
		inTransaction bool
	}
	tests := []struct {
		name          string
		args          args
		assertionFunc assert.BoolAssertionFunc
	}{
		{
			name: "only CREATE INDEX statements are evaluated",
			args: args{
				statement:     `ALTER TABLE "public"."movie_ratings" ADD COLUMN "rating" BIGINT NOT NULL`,
				inTransaction: false,
			},
			assertionFunc: assert.False,
		},
		{
			name: "non-concurrent CREATE INDEX statement",
			args: args{
				statement:     `CREATE INDEX idx_created_at_rating ON "movie_ratings" ("created_at", "rating")`,
				inTransaction: false,
			},
			assertionFunc: assert.True,
		},
		{
			name: "concurrent CREATE INDEX statement",
			args: args{
				statement:     `CREATE INDEX CONCURRENTLY idx_created_at_rating ON "movie_ratings" ("created_at", "rating")`,
				inTransaction: false,
			},
			assertionFunc: assert.False,
		},
		{
			name: "non-concurrent CREATE UNIQUE INDEX statement",
			args: args{
				statement:     `CREATE UNIQUE INDEX idx_movie_id_user_id ON "movie_ratings" ("movie_id", "user_id")`,
				inTransaction: false,
			},
			assertionFunc: assert.True,
		},
		{
			name: "concurrent CREATE UNIQUE INDEX statement",
			args: args{
				statement:     `CREATE UNIQUE INDEX CONCURRENTLY idx_movie_id_user_id ON "movie_ratings" ("movie_id", "user_id")`,
				inTransaction: false,
			},
			assertionFunc: assert.False,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := CreateIndexNonConcurrently{}
			node := parseStatement(t, tt.args.statement)

			tt.assertionFunc(t, r.Process(node, nil, tt.args.inTransaction))
		})
	}
}

func parseStatement(t *testing.T, sql string) *pg_query.Node {
	t.Helper()

	node, err := pg_query.Parse(sql)
	require.NoError(t, err)

	return node.GetStmts()[0].Stmt
}
