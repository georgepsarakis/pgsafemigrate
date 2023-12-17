package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestColumnComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		want bool
	}{
		{
			name: "no comment for added column",
			sql:  `ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;`,
			want: true,
		},
		{
			name: "comment for added column",
			sql: `ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;
COMMENT ON COLUMN movies.released_at IS 'First release date';`,
			want: false,
		},
		{
			name: "no comment for dropped column",
			sql:  `ALTER TABLE movies DROP COLUMN released_at;`,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := ColumnComment{}
			var allNodes []*pg_query.Node
			for _, statement := range strings.Split(tt.sql, "\n") {
				allNodes = append(allNodes, parseStatement(t, statement))
			}
			assert.Equal(t, tt.want, r.Process(allNodes[0], allNodes, true))
		})
	}
}
