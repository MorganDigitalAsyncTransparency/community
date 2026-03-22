// Spec: specs/api/triage-time.md (TT-1, TT-2, TT-3, TT-4, TT-5, TT-6, TT-7, TT-11, TT-12)
// Tests: backend/domain/triage_unit_test.go
package domain

import (
	"encoding/json"
	"sort"

	"github.com/code-community/discourse-observer/backend/model"
)

// TriageResult holds the computed triage time metrics.
type TriageResult struct {
	MedianHours *float64         `json:"medianHours"`
	Count       int              `json:"count"`
	ByTag       []TriageTagEntry `json:"byTag"`
}

// TriageTagEntry holds triage metrics for a single tag.
type TriageTagEntry struct {
	Tag         string   `json:"tag"`
	MedianHours *float64 `json:"medianHours"`
	Count       int      `json:"count"`
}

// ComputeTriageTime calculates triage time metrics from topics and their events.
// Triage time is the duration from topic creation to the first tag_change event.
// Topics with no tag_change events are excluded (TT-2).
func ComputeTriageTime(topics []model.Topic, events map[int][]model.TopicEvent) TriageResult {
	var durations []int64
	tagDurations := map[string][]int64{}

	for i := range topics {
		t := &topics[i]
		topicEvents := events[t.ID]

		firstTag, durationMs := extractTriageInfo(t, topicEvents)
		if firstTag == "" {
			continue
		}

		durations = append(durations, durationMs)
		tagDurations[firstTag] = append(tagDurations[firstTag], durationMs)
	}

	result := TriageResult{
		Count: len(durations),
		ByTag: []TriageTagEntry{},
	}
	result.MedianHours = msToHours(Median(durations))

	for tag, durs := range tagDurations {
		result.ByTag = append(result.ByTag, TriageTagEntry{
			Tag:         tag,
			MedianHours: msToHours(Median(durs)),
			Count:       len(durs),
		})
	}
	sort.Slice(result.ByTag, func(i, j int) bool {
		if result.ByTag[i].Count != result.ByTag[j].Count {
			return result.ByTag[i].Count > result.ByTag[j].Count
		}
		return result.ByTag[i].Tag < result.ByTag[j].Tag
	})

	return result
}

// extractTriageInfo finds the first tag_change event for a topic and returns
// the first tag added and the triage duration in milliseconds.
// Returns empty string if no qualifying event exists.
func extractTriageInfo(topic *model.Topic, events []model.TopicEvent) (firstTag string, durationMs int64) {
	for _, e := range events {
		if e.EventType != "tag_change" {
			continue
		}
		tag := parseFirstTag(e.Detail)
		if tag == "" {
			continue
		}
		durationMs := e.HappenedAt.Sub(topic.CreatedAt).Milliseconds()
		return tag, durationMs
	}
	return "", 0
}

// parseFirstTag extracts the first tag from a tag_change event's Detail JSON.
// The Detail field is a JSON-encoded RevisionTagChange with previous/current arrays.
// The first tag is the first element in the current array that was not in previous.
func parseFirstTag(detail string) string {
	var change model.RevisionTagChange
	if err := json.Unmarshal([]byte(detail), &change); err != nil {
		return ""
	}
	prev := toSet(change.Previous)
	for _, tag := range change.Current {
		if !prev[tag] {
			return tag
		}
	}
	if len(change.Current) > 0 {
		return change.Current[0]
	}
	return ""
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func msToHours(ms *int64) *float64 {
	if ms == nil {
		return nil
	}
	h := float64(*ms) / 3_600_000.0
	return &h
}
