package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNestedTransaction_Process(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sql           string
		inTransaction bool
		want          bool
	}{
		{
			name:          "transaction BEGIN statement found",
			sql:           `BEGIN; ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;`,
			inTransaction: true,
			want:          true,
		},
		{
			name:          "transaction BEGIN statement found but no active transaction context",
			sql:           `BEGIN; ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;`,
			inTransaction: false,
			want:          true,
		},
		{
			name:          "no transaction management statements found",
			sql:           `ALTER TABLE movies ADD COLUMN released_at TIMESTAMP;`,
			inTransaction: true,
			want:          false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NestedTransaction{}
			var allNodes []*pg_query.Node
			for _, statement := range strings.Split(tt.sql, "\n") {
				allNodes = append(allNodes, parseStatement(t, statement))
			}
			assert.Equal(t, tt.want, r.Process(allNodes[0], allNodes, true))
		})
	}
}
