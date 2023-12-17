package rules

import (
	pg_query "github.com/pganalyze/pg_query_go/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockRule struct {
	alias string
}

func (m mockRule) Alias() string {
	return m.alias
}
func (m mockRule) Documentation() string                                     { return "" }
func (m mockRule) Process(_ *pg_query.Node, _ []*pg_query.Node, _ bool) bool { return false }

func TestRuleSet_Contains(t *testing.T) {
	t.Parallel()

	set := RuleSet{
		"test-1": mockRule{alias: "test-1"},
		"test-2": mockRule{alias: "test-2"},
	}
	tests := []struct {
		name          string
		ruleSet       RuleSet
		input         string
		assertionFunc assert.BoolAssertionFunc
	}{
		{
			name:          "contains a rule with the given alias",
			ruleSet:       set,
			input:         "test-1",
			assertionFunc: assert.True,
		},
		{
			name:          "does not contain a rule with the given alias",
			ruleSet:       set,
			input:         "test-3",
			assertionFunc: assert.False,
		},
		{
			name:          "empty set",
			ruleSet:       RuleSet{},
			input:         "test-1",
			assertionFunc: assert.False,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.assertionFunc(t, tt.ruleSet.Contains(tt.input))
		})
	}
}

func TestRuleSet_SortedSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    RuleSet
		want []Rule
	}{
		{
			name: "sorts according to the alias",
			r: func() RuleSet {
				set := RuleSet{}
				set.Add(mockRule{alias: "test-b"})
				set.Add(mockRule{alias: "test-a"})
				set.Add(mockRule{alias: "test-z"})
				return set
			}(),
			want: []Rule{
				mockRule{alias: "test-a"},
				mockRule{alias: "test-b"},
				mockRule{alias: "test-z"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.r.SortedSlice(), "SortedSlice()")
		})
	}
}

func TestRuleSet_Except(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		r          RuleSet
		exceptions []string
		want       RuleSet
	}{
		{
			name: "removes the rules with matching aliases",
			r: func() RuleSet {
				rs := NewRuleSet()
				rs.Add(mockRule{alias: "test-A"})
				rs.Add(mockRule{alias: "test-B"})
				rs.Add(mockRule{alias: "test-C"})
				return rs
			}(),
			exceptions: []string{"test-B"},
			want: func() RuleSet {
				rs := NewRuleSet()
				rs.Add(mockRule{alias: "test-A"})
				rs.Add(mockRule{alias: "test-C"})
				return rs
			}(),
		},
		{
			name: "ignores exceptions with no matching aliases",
			r: func() RuleSet {
				rs := NewRuleSet()
				rs.Add(mockRule{alias: "test-A"})
				rs.Add(mockRule{alias: "test-B"})
				rs.Add(mockRule{alias: "test-C"})
				return rs
			}(),
			exceptions: []string{"test-Z"},
			want: func() RuleSet {
				rs := NewRuleSet()
				rs.Add(mockRule{alias: "test-A"})
				rs.Add(mockRule{alias: "test-B"})
				rs.Add(mockRule{alias: "test-C"})
				return rs
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.r.Except(tt.exceptions...))
		})
	}
}
