package slicesort

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringsAscendingOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "sorts strings lexicographically",
			input: []string{"a2", "a1", "12", "c3"},
			want:  []string{"12", "a1", "a2", "c3"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, StringsAscendingOrder(tt.input))
		})
	}
}
