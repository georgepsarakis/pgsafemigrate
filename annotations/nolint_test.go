package annotations

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		want       NoLint
	}{
		{
			name:       "annotation with specific rule exclusion",
			annotation: "pgsafemigrate:nolint:exclusive-locking-column-type-change",
			want:       NoLint{Valid: true, RuleNames: []string{"exclusive-locking-column-type-change"}},
		},
		{
			name:       "annotation with specific multiple rule exclusions",
			annotation: "pgsafemigrate:nolint:exclusive-locking-column-type-change,high-availability-avoid-table-rename",
			want:       NoLint{Valid: true, RuleNames: []string{"exclusive-locking-column-type-change", "high-availability-avoid-table-rename"}},
		},
		{
			name:       "exclude all rules",
			annotation: "pgsafemigrate:nolint",
			want:       NoLint{Valid: true},
		},
		{
			name:       "nolint directive missing",
			annotation: "pgsafemigrate:exclusive-locking-column-type-change",
			want:       NoLint{},
		},
		{
			name:       "unknown string",
			annotation: "an sql comment",
			want:       NoLint{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Parse(tt.annotation))
		})
	}
}
