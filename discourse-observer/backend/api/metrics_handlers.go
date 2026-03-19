// Spec: specs/api/api-contract.md (AC-16, AC-17, AC-18, AC-19)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleMetricsSummary(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())
	result := domain.ComputeMetricsSummary(topics)

	respondJSON(w, map[string]any{
		"medianFirstReplyMs": result.MedianFirstReplyMs,
		"medianResolutionMs": result.MedianResolutionMs,
		"solvedCount":        result.SolvedCount,
		"selfClosedCount":    result.SelfClosedCount,
		"answerRatePercent":  result.AnswerRatePercent,
	})
}

func (s *Server) handleMetricsVolume(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())
	granularity := domain.Granularity(f.Period, f.From, f.To)
	start, end := domain.ComputeTimeRange(topics, granularity)
	buckets := domain.ComputeVolumeBuckets(topics, granularity, start, end)

	if buckets == nil {
		buckets = []domain.VolumeBucket{}
	}

	type item struct {
		Label     string `json:"label"`
		BucketKey string `json:"bucketKey"`
		Created   int    `json:"created"`
		Accepted  int    `json:"accepted"`
		Closed    int    `json:"closed"`
		Open      int    `json:"open"`
	}
	items := make([]item, len(buckets))
	for i, b := range buckets {
		items[i] = item{
			Label: b.Label, BucketKey: b.BucketKey,
			Created: b.Created, Accepted: b.Accepted,
			Closed: b.Closed, Open: b.Open,
		}
	}
	respondJSON(w, items)
}

func (s *Server) handleMetricsMedianTrends(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())
	granularity := domain.Granularity(f.Period, f.From, f.To)
	start, end := domain.ComputeTimeRange(topics, granularity)

	firstReply := domain.ComputeMedianTrend(topics, granularity, start, end, domain.FirstReplyExtractor)
	resolution := domain.ComputeMedianTrend(topics, granularity, start, end, domain.ResolutionExtractor)

	if firstReply == nil {
		firstReply = []domain.MedianBucket{}
	}
	if resolution == nil {
		resolution = []domain.MedianBucket{}
	}

	type bucket struct {
		Label     string `json:"label"`
		BucketKey string `json:"bucketKey"`
		MedianMs  *int64 `json:"medianMs"`
	}
	toBuckets := func(src []domain.MedianBucket) []bucket {
		out := make([]bucket, len(src))
		for i, b := range src {
			out[i] = bucket{Label: b.Label, BucketKey: b.BucketKey, MedianMs: b.MedianMs}
		}
		return out
	}
	respondJSON(w, map[string]any{
		"firstReply": toBuckets(firstReply),
		"resolution": toBuckets(resolution),
	})
}

func (s *Server) handleMetricsDistribution(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())

	frDurations := domain.FirstReplyDurations(topics)
	resDurations := domain.ResolutionDurations(topics)
	frBuckets := domain.BucketDurations(frDurations, s.BucketCeilings)
	resBuckets := domain.BucketDurations(resDurations, s.BucketCeilings)

	type bucket struct {
		Label string `json:"label"`
		Count int    `json:"count"`
	}
	toBuckets := func(src []domain.DistributionBucket) []bucket {
		out := make([]bucket, len(src))
		for i, b := range src {
			out[i] = bucket{Label: b.Label, Count: b.Count}
		}
		return out
	}
	respondJSON(w, map[string]any{
		"firstReply": toBuckets(frBuckets),
		"resolution": toBuckets(resBuckets),
	})
}
