// Spec: specs/api/api-contract.md (AC-8, AC-9, AC-10)
// Tests: backend/domain/filter_test.go
package domain

import (
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// FilterByPeriod keeps topics created within the rolling window ending at now.
func FilterByPeriod(topics []model.Topic, period string, now time.Time) []model.Topic {
	days := periodToDays(period)
	if days <= 0 {
		return topics
	}
	cutoff := now.AddDate(0, 0, -days)
	return filterByTime(topics, cutoff, now)
}

// FilterByDateRange keeps topics created between from (00:00:00Z) and to (23:59:59.999Z).
func FilterByDateRange(topics []model.Topic, from, to time.Time) []model.Topic {
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999_000_000, time.UTC)
	return filterByTime(topics, start, end)
}

// FilterByTag keeps topics that carry the given tag. Empty tag returns all.
func FilterByTag(topics []model.Topic, tag string) []model.Topic {
	if tag == "" {
		return topics
	}
	var result []model.Topic
	for i := range topics {
		if hasTag(topics[i].Tags, tag) {
			result = append(result, topics[i])
		}
	}
	return result
}

// FilterByMonitoredTags keeps topics that carry at least one monitored tag.
// Untagged topics are excluded. AC-10: when no tag filter is specified,
// only monitored topics are included.
func FilterByMonitoredTags(topics []model.Topic, monitored map[string]bool) []model.Topic {
	var result []model.Topic
	for i := range topics {
		for _, tag := range topics[i].Tags {
			if monitored[tag] {
				result = append(result, topics[i])
				break
			}
		}
	}
	return result
}

// FilterUnreplied returns topics with no replies and no outcome.
func FilterUnreplied(topics []model.Topic) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if topics[i].ReplyCount == 0 && topics[i].Outcome == "" {
			result = append(result, topics[i])
		}
	}
	return result
}

// FilterUntagged returns topics with empty tag lists.
func FilterUntagged(topics []model.Topic) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if len(topics[i].Tags) == 0 {
			result = append(result, topics[i])
		}
	}
	return result
}

// FilterResolved returns topics with an outcome.
func FilterResolved(topics []model.Topic) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if topics[i].Outcome != "" {
			result = append(result, topics[i])
		}
	}
	return result
}

// FilterRepliedOpen returns topics with replies but no outcome,
// excluding those resolved via closedTag.
func FilterRepliedOpen(topics []model.Topic, resolved map[string]model.ResolvedTag) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if topics[i].ReplyCount > 0 && topics[i].Outcome == "" && !isClosedByTag(&topics[i], resolved) {
			result = append(result, topics[i])
		}
	}
	return result
}

func periodToDays(period string) int {
	switch period {
	case "7d":
		return 7
	case "30d":
		return 30
	case "1y":
		return 365
	default:
		return 0
	}
}

func filterByTime(topics []model.Topic, from, to time.Time) []model.Topic {
	var result []model.Topic
	for i := range topics {
		if !topics[i].CreatedAt.Before(from) && !topics[i].CreatedAt.After(to) {
			result = append(result, topics[i])
		}
	}
	return result
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func isClosedByTag(topic *model.Topic, resolved map[string]model.ResolvedTag) bool {
	for _, tag := range topic.Tags {
		rt, ok := resolved[tag]
		if !ok {
			continue
		}
		if rt.ClosedTag == nil {
			continue
		}
		if hasTag(topic.Tags, *rt.ClosedTag) {
			return true
		}
	}
	return false
}
