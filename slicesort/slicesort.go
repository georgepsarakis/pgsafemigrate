package slicesort

import (
	"sort"
)

func StringsAscendingOrder(unsorted []string) []string {
	sorted := make([]string, len(unsorted))
	copy(sorted, unsorted)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	return sorted
}
