// Spec: specs/api/api-contract.md (AC-12, AC-15)
// Tests: backend/domain/queue_test.go
package domain

import (
	"sort"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// QueueSummary holds computed queue summary values.
type QueueSummary struct {
	UnrepliedCount      int
	UntaggedCount       int
	OldestUnrepliedDays *int
}

// ComputeQueueSummary computes queue summary from pre-filtered topics.
func ComputeQueueSummary(topics []model.Topic, now time.Time) QueueSummary {
	unreplied := FilterUnreplied(topics)
	untagged := FilterUntagged(topics)

	var oldest *int
	if len(unreplied) > 0 {
		maxAge := 0
		for i := range unreplied {
			days := int(now.Sub(unreplied[i].CreatedAt).Hours() / 24)
			if days > maxAge {
				maxAge = days
			}
		}
		oldest = &maxAge
	}

	return QueueSummary{
		UnrepliedCount:      len(unreplied),
		UntaggedCount:       len(untagged),
		OldestUnrepliedDays: oldest,
	}
}

// StalledTopic holds computed stalled topic data for one topic.
type StalledTopic struct {
	Topic              model.Topic
	StrictestTag       *string
	ThresholdDays      int
	ThresholdIsDefault bool
	DaysSinceActivity  int
}

// FindStalledTopics returns replied-open topics that exceed their stalled threshold.
func FindStalledTopics(
	topics []model.Topic,
	resolved map[string]model.ResolvedTag,
	defaultStalledDays int,
	now time.Time,
) []StalledTopic {
	repliedOpen := FilterRepliedOpen(topics, resolved)

	var stalled []StalledTopic
	for i := range repliedOpen {
		tag, days, isDefault := strictestStalledThreshold(repliedOpen[i].Tags, resolved, defaultStalledDays)
		daysSince := daysSinceLastActivity(&repliedOpen[i], now)
		if daysSince <= days {
			continue
		}
		stalled = append(stalled, StalledTopic{
			Topic:              repliedOpen[i],
			StrictestTag:       tag,
			ThresholdDays:      days,
			ThresholdIsDefault: isDefault,
			DaysSinceActivity:  daysSince,
		})
	}

	sort.Slice(stalled, func(i, j int) bool {
		return stalled[i].DaysSinceActivity > stalled[j].DaysSinceActivity
	})
	return stalled
}

func strictestStalledThreshold(
	tags []string,
	resolved map[string]model.ResolvedTag,
	defaultDays int,
) (tag *string, days int, isDefault bool) {
	var bestTag *string
	bestDays := -1
	bestIsDefault := true

	for _, tag := range tags {
		rt, ok := resolved[tag]
		if !ok {
			continue
		}
		if bestDays < 0 || rt.StalledDays < bestDays {
			t := tag
			bestTag = &t
			bestDays = rt.StalledDays
			bestIsDefault = rt.StalledDaysIsDefault
		}
	}

	if bestTag == nil {
		return nil, defaultDays, true
	}
	return bestTag, bestDays, bestIsDefault
}

func daysSinceLastActivity(t *model.Topic, now time.Time) int {
	ref := t.CreatedAt
	if t.LastActivityAt != nil {
		ref = *t.LastActivityAt
	}
	return int(now.Sub(ref).Hours() / 24)
}
