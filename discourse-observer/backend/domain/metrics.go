// Spec: specs/api/api-contract.md (AC-16)
// Tests: backend/domain/metrics_test.go
package domain

import (
	"math"

	"github.com/code-community/discourse-observer/backend/model"
)

// MetricsSummary holds computed response metrics.
type MetricsSummary struct {
	MedianFirstReplyMs *int64
	MedianResolutionMs *int64
	SolvedCount        int
	SelfClosedCount    int
	AnswerRatePercent  *int
}

// ComputeMetricsSummary calculates aggregate response metrics from topics.
func ComputeMetricsSummary(topics []model.Topic) MetricsSummary {
	var firstReplyDurations []int64
	var resolutionDurations []int64
	solved := 0
	selfClosed := 0

	for i := range topics {
		if topics[i].FirstReplyAt != nil {
			dur := topics[i].FirstReplyAt.Sub(topics[i].CreatedAt).Milliseconds()
			firstReplyDurations = append(firstReplyDurations, dur)
		}
		if topics[i].ResolvedAt != nil {
			dur := topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds()
			resolutionDurations = append(resolutionDurations, dur)
		}
		switch topics[i].Outcome {
		case "solved":
			solved++
		case "self-closed":
			selfClosed++
		}
	}

	result := MetricsSummary{
		MedianFirstReplyMs: Median(firstReplyDurations),
		MedianResolutionMs: Median(resolutionDurations),
		SolvedCount:        solved,
		SelfClosedCount:    selfClosed,
	}

	total := solved + selfClosed
	if total > 0 {
		rate := int(math.Round(float64(solved) / float64(total) * 100))
		result.AnswerRatePercent = &rate
	}

	return result
}
