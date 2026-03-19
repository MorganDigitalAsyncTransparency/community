// Spec: specs/api/api-contract.md (AC-24, AC-25)
// Tests: backend/domain/slo_test.go
package domain

import (
	"sort"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// Violation holds one SLO violation.
type Violation struct {
	TopicID            int
	TopicTitle         string
	TopicURL           string
	Tag                string
	ThresholdMs        int64
	ActualMs           int64
	ExcessMs           int64
	ThresholdIsDefault bool
}

// ViolationGroups holds violations grouped by type.
type ViolationGroups struct {
	FirstReply []Violation
	Resolution []Violation
	Inactivity []Violation
}

// FindViolations identifies topics exceeding SLO thresholds.
// Each violation type uses the strictest threshold for that specific metric.
func FindViolations(
	topics []model.Topic,
	resolved map[string]model.ResolvedTag,
	now time.Time,
) ViolationGroups {
	var groups ViolationGroups

	for i := range topics {
		if !hasMonitoredTag(topics[i].Tags, resolved) {
			continue
		}

		// First-reply violations (resolved topics with firstReplyAt)
		if topics[i].FirstReplyAt != nil {
			actual := topics[i].FirstReplyAt.Sub(topics[i].CreatedAt).Milliseconds()
			tag, thresholdMs, isDefault := strictestForMetric(topics[i].Tags, resolved, metricFirstReply)
			if tag != "" && actual > thresholdMs {
				groups.FirstReply = append(groups.FirstReply, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: thresholdMs,
					ActualMs: actual, ExcessMs: actual - thresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
		}

		// Unreplied topics: check first-reply and inactivity
		if topics[i].Outcome == "" && topics[i].ReplyCount == 0 {
			elapsed := now.Sub(topics[i].CreatedAt).Milliseconds()

			tag, thresholdMs, isDefault := strictestForMetric(topics[i].Tags, resolved, metricFirstReply)
			if tag != "" && elapsed > thresholdMs {
				groups.FirstReply = append(groups.FirstReply, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: thresholdMs,
					ActualMs: elapsed, ExcessMs: elapsed - thresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}

			tag, thresholdMs, isDefault = strictestForMetric(topics[i].Tags, resolved, metricInactivity)
			if tag != "" && elapsed > thresholdMs {
				groups.Inactivity = append(groups.Inactivity, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: thresholdMs,
					ActualMs: elapsed, ExcessMs: elapsed - thresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
		}

		// Resolution violations
		if topics[i].ResolvedAt != nil {
			actual := topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds()
			tag, thresholdMs, isDefault := strictestForMetric(topics[i].Tags, resolved, metricResolution)
			if tag != "" && actual > thresholdMs {
				groups.Resolution = append(groups.Resolution, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: thresholdMs,
					ActualMs: actual, ExcessMs: actual - thresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
		}
	}

	sortViolations(groups.FirstReply)
	sortViolations(groups.Resolution)
	sortViolations(groups.Inactivity)
	return groups
}

// TagCompliance holds per-tag SLO compliance rates.
type TagCompliance struct {
	Tag                string
	FirstReplyPercent  *int
	ResolutionPercent  *int
	InactivityPercent  *int
	ThresholdIsDefault bool
}

// ComputeCompliance calculates per-tag SLO compliance percentages.
func ComputeCompliance(
	topics []model.Topic,
	resolved map[string]model.ResolvedTag,
	now time.Time,
) []TagCompliance {
	type tagStats struct {
		firstReplyTotal   int
		firstReplyPassing int
		resolutionTotal   int
		resolutionPassing int
		inactivityTotal   int
		inactivityPassing int
		isDefault         bool
	}

	stats := make(map[string]*tagStats)

	for tagName, rt := range resolved {
		stats[tagName] = &tagStats{isDefault: rt.SLOIsDefault}
	}

	for i := range topics {
		for _, tag := range topics[i].Tags {
			rt, ok := resolved[tag]
			if !ok {
				continue
			}
			s := stats[tag]

			firstReplyMs := int64(rt.SLO.FirstReplyHours) * 3_600_000
			resolutionMs := int64(rt.SLO.ResolutionHours) * 3_600_000
			inactivityMs := int64(rt.SLO.InactivityHours) * 3_600_000

			// First reply: resolved+unreplied topics
			if topics[i].FirstReplyAt != nil {
				s.firstReplyTotal++
				if topics[i].FirstReplyAt.Sub(topics[i].CreatedAt).Milliseconds() <= firstReplyMs {
					s.firstReplyPassing++
				}
			} else if topics[i].Outcome == "" && topics[i].ReplyCount == 0 {
				s.firstReplyTotal++
				if now.Sub(topics[i].CreatedAt).Milliseconds() <= firstReplyMs {
					s.firstReplyPassing++
				}
			}

			// Resolution: only resolved topics
			if topics[i].ResolvedAt != nil {
				s.resolutionTotal++
				if topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds() <= resolutionMs {
					s.resolutionPassing++
				}
			}

			// Inactivity: only unreplied topics
			if topics[i].Outcome == "" && topics[i].ReplyCount == 0 {
				s.inactivityTotal++
				if now.Sub(topics[i].CreatedAt).Milliseconds() <= inactivityMs {
					s.inactivityPassing++
				}
			}
		}
	}

	result := make([]TagCompliance, 0, len(stats))
	for tag, s := range stats {
		tc := TagCompliance{Tag: tag, ThresholdIsDefault: s.isDefault}
		if s.firstReplyTotal > 0 {
			pct := percent(s.firstReplyPassing, s.firstReplyTotal)
			tc.FirstReplyPercent = &pct
		}
		if s.resolutionTotal > 0 {
			pct := percent(s.resolutionPassing, s.resolutionTotal)
			tc.ResolutionPercent = &pct
		}
		if s.inactivityTotal > 0 {
			pct := percent(s.inactivityPassing, s.inactivityTotal)
			tc.InactivityPercent = &pct
		}
		result = append(result, tc)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Tag < result[j].Tag
	})
	return result
}

type sloMetric int

const (
	metricFirstReply sloMetric = iota
	metricResolution
	metricInactivity
)

func hasMonitoredTag(tags []string, resolved map[string]model.ResolvedTag) bool {
	for _, tag := range tags {
		if _, ok := resolved[tag]; ok {
			return true
		}
	}
	return false
}

// strictestForMetric returns the tag with the lowest threshold for a specific metric.
func strictestForMetric(
	tags []string,
	resolved map[string]model.ResolvedTag,
	metric sloMetric,
) (tag string, thresholdMs int64, isDefault bool) {
	bestTag := ""
	var bestMs int64
	bestIsDefault := true

	for _, t := range tags {
		rt, ok := resolved[t]
		if !ok {
			continue
		}
		var hours int
		switch metric {
		case metricFirstReply:
			hours = rt.SLO.FirstReplyHours
		case metricResolution:
			hours = rt.SLO.ResolutionHours
		case metricInactivity:
			hours = rt.SLO.InactivityHours
		}
		ms := int64(hours) * 3_600_000
		if bestTag == "" || ms < bestMs {
			bestTag = t
			bestMs = ms
			bestIsDefault = rt.SLOIsDefault
		}
	}
	return bestTag, bestMs, bestIsDefault
}

func sortViolations(violations []Violation) {
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].ExcessMs > violations[j].ExcessMs
	})
}

func percent(passing, total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(passing)/float64(total)*100.0 + 0.5)
}
