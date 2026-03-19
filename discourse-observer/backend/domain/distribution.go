// Spec: specs/api/api-contract.md (AC-19)
// Tests: backend/domain/distribution_test.go
package domain

import (
	"fmt"

	"github.com/code-community/discourse-observer/backend/model"
)

// DistributionBucket holds one histogram bucket.
type DistributionBucket struct {
	Label string
	Count int
}

// BucketDurations distributes durations (in ms) into histogram buckets defined by ceilingsHours.
// A duration is placed in the first bucket where duration < ceiling.
// Equal-to-ceiling goes into the next bucket.
func BucketDurations(durationsMs []int64, ceilingsHours []int) []DistributionBucket {
	labels := distributionLabels(ceilingsHours)
	counts := make([]int, len(labels))

	for _, dur := range durationsMs {
		idx := len(ceilingsHours)
		for i, ceil := range ceilingsHours {
			ceilMs := int64(ceil) * 3_600_000
			if dur < ceilMs {
				idx = i
				break
			}
		}
		counts[idx]++
	}

	buckets := make([]DistributionBucket, len(labels))
	for i, label := range labels {
		buckets[i] = DistributionBucket{Label: label, Count: counts[i]}
	}
	return buckets
}

// FirstReplyDurations extracts first-reply durations in ms from topics that have firstReplyAt.
func FirstReplyDurations(topics []model.Topic) []int64 {
	var durations []int64
	for i := range topics {
		if topics[i].FirstReplyAt != nil {
			durations = append(durations, topics[i].FirstReplyAt.Sub(topics[i].CreatedAt).Milliseconds())
		}
	}
	return durations
}

// ResolutionDurations extracts resolution durations in ms from topics that have resolvedAt.
func ResolutionDurations(topics []model.Topic) []int64 {
	var durations []int64
	for i := range topics {
		if topics[i].ResolvedAt != nil {
			durations = append(durations, topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds())
		}
	}
	return durations
}

func distributionLabels(ceilingsHours []int) []string {
	n := len(ceilingsHours)
	labels := make([]string, n+1)
	labels[0] = fmt.Sprintf("< %s", formatCeiling(ceilingsHours[0]))
	for i := 1; i < n; i++ {
		labels[i] = fmt.Sprintf("%s\u2013%s", formatCeiling(ceilingsHours[i-1]), formatCeiling(ceilingsHours[i]))
	}
	labels[n] = fmt.Sprintf("> %s", formatCeiling(ceilingsHours[n-1]))
	return labels
}

func formatCeiling(hours int) string {
	if hours >= 24 && hours%24 == 0 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dh", hours)
}
