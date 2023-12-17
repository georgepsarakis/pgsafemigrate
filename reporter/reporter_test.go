package reporter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockError struct {
	alias         string
	statement     string
	documentation string
}

func (m mockError) Alias() string         { return m.alias }
func (m mockError) Documentation() string { return m.documentation }
func (m mockError) Statement() string     { return m.statement }

func TestPlainText_Print(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		reports []Report
		want    string
	}{
		{
			name: "no errors in reports",
			reports: []Report{
				{
					FilePath: "sql/migration-1.sql",
					Errors:   ValidationErrors{},
				},
				{
					FilePath: "sql/migration-1.sql",
					Errors:   ValidationErrors{},
				},
				{
					FilePath: "sql/migration-2.sql",
					Errors:   ValidationErrors{},
				},
			},
			want: "",
		},
		{
			name: "errors in reports",
			reports: []Report{
				{
					FilePath: "sql/migration-1.sql",
					Errors: ValidationErrors{
						mockError{alias: "test-rule-1", statement: "SELECT 1", documentation: "test docs #1"},
						mockError{alias: "test-rule-1", statement: "SELECT 2", documentation: "test docs #2"},
					},
				},
				{
					FilePath: "sql/migration-1.sql",
					Errors: ValidationErrors{
						mockError{alias: "test-rule-2", statement: "SELECT 3", documentation: "test docs #3"},
					},
				},
				{
					FilePath: "sql/migration-2.sql",
					Errors: ValidationErrors{
						mockError{alias: "test-rule-2", statement: "SELECT 4", documentation: "test docs #4"},
					},
				},
			},
			want: "File sql/migration-1.sql Results:\n" +
				"\tRule test-rule-1 violation found for statement:\n" +
				"\t  SELECT 1\n\tExplanation: test docs #1\n\n" +
				"\tRule test-rule-1 violation found for statement:\n\t  SELECT 2\n" +
				"\tExplanation: test docs #2\n\n" +
				"\tRule test-rule-2 violation found for statement:\n\t  SELECT 3\n" +
				"\tExplanation: test docs #3\n\nFile sql/migration-2.sql Results:\n" +
				"\tRule test-rule-2 violation found for statement:\n\t  SELECT 4\n" +
				"\tExplanation: test docs #4",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := PlainText{}

			assert.Equal(t, tt.want, p.Print(tt.reports))
		})
	}
}
