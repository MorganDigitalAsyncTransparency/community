// Spec: specs/api/api-contract.md (AC-17, AC-18)
// Tests: backend/domain/trend_test.go
package domain

import (
	"github.com/code-community/discourse-observer/backend/model"
)

// VolumeBucket holds topic counts for one time bucket.
type VolumeBucket struct {
	Label     string
	BucketKey string
	Created   int
	Accepted  int
	Closed    int
	Open      int
}

// ComputeVolumeBuckets groups topics by time bucket and counts by outcome.
func ComputeVolumeBuckets(topics []model.Topic, granularity, start, end string) []VolumeBucket {
	if start == "" || end == "" {
		return nil
	}

	created := bucketCount(topics, granularity)
	solved := bucketCount(filterByOutcome(topics, "solved"), granularity)
	selfClosed := bucketCount(filterByOutcome(topics, "self-closed"), granularity)
	open := bucketCount(filterByOutcome(topics, ""), granularity)

	keys := GenerateBucketKeys(start, end, granularity)
	buckets := make([]VolumeBucket, len(keys))
	for i, k := range keys {
		buckets[i] = VolumeBucket{
			Label:     FormatBucketLabel(k, granularity),
			BucketKey: k,
			Created:   created[k],
			Accepted:  solved[k],
			Closed:    selfClosed[k],
			Open:      open[k],
		}
	}
	return buckets
}

// MedianBucket holds median duration for one time bucket.
type MedianBucket struct {
	Label     string
	BucketKey string
	MedianMs  *int64
}

// ComputeMedianTrend groups topics by time bucket and computes median duration per bucket.
// extractor returns the duration in ms for a topic, or nil if not applicable.
func ComputeMedianTrend(
	topics []model.Topic,
	granularity string,
	start, end string,
	extractor func(*model.Topic) *int64,
) []MedianBucket {
	if start == "" || end == "" {
		return nil
	}

	grouped := make(map[string][]int64)
	for i := range topics {
		dur := extractor(&topics[i])
		if dur == nil {
			continue
		}
		key := BucketKey(topics[i].CreatedAt, granularity)
		grouped[key] = append(grouped[key], *dur)
	}

	keys := GenerateBucketKeys(start, end, granularity)
	buckets := make([]MedianBucket, len(keys))
	for i, k := range keys {
		buckets[i] = MedianBucket{
			Label:     FormatBucketLabel(k, granularity),
			BucketKey: k,
			MedianMs:  Median(grouped[k]),
		}
	}
	return buckets
}

// FirstReplyExtractor returns the first-reply duration in ms, or nil.
func FirstReplyExtractor(t *model.Topic) *int64 {
	if t.FirstReplyAt == nil {
		return nil
	}
	d := t.FirstReplyAt.Sub(t.CreatedAt).Milliseconds()
	return &d
}

// ResolutionExtractor returns the resolution duration in ms, or nil.
func ResolutionExtractor(t *model.Topic) *int64 {
	if t.ResolvedAt == nil {
		return nil
	}
	d := t.ResolvedAt.Sub(t.CreatedAt).Milliseconds()
	return &d
}

// ComputeTimeRange returns the first and last bucket keys for a set of topics.
func ComputeTimeRange(topics []model.Topic, granularity string) (start, end string) {
	if len(topics) == 0 {
		return "", ""
	}
	first := topics[0].CreatedAt
	last := topics[0].CreatedAt
	for i := range topics[1:] {
		if topics[i+1].CreatedAt.Before(first) {
			first = topics[i+1].CreatedAt
		}
		if topics[i+1].CreatedAt.After(last) {
			last = topics[i+1].CreatedAt
		}
	}
	return BucketKey(first, granularity), BucketKey(last, granularity)
}

func bucketCount(topics []model.Topic, granularity string) map[string]int {
	counts := make(map[string]int)
	for i := range topics {
		key := BucketKey(topics[i].CreatedAt, granularity)
		counts[key]++
	}
	return counts
}

func filterByOutcome(topics []model.Topic, outcome string) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if topics[i].Outcome == outcome {
			result = append(result, topics[i])
		}
	}
	return result
}
