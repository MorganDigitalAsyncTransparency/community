// Spec: specs/api/api-contract.md (AC-16, AC-18, AC-21)
// Tests: backend/domain/median_test.go
package domain

import "sort"

// Median returns the median of a sorted copy of values.
// For even-length slices, it returns the truncated average of the two middle values.
// Returns nil for empty input.
func Median(values []int64) *int64 {
	n := len(values)
	if n == 0 {
		return nil
	}
	sorted := make([]int64, n)
	copy(sorted, values)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var result int64
	if n%2 == 1 {
		result = sorted[n/2]
	} else {
		result = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return &result
}
