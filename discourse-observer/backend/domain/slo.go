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
func FindViolations(
	topics []model.Topic,
	resolved map[string]model.ResolvedTag,
	now time.Time,
) ViolationGroups {
	var groups ViolationGroups

	for i := range topics {
		tag, slo, isDefault := strictestSLO(topics[i].Tags, resolved)
		if tag == "" {
			continue
		}
		firstReplyThresholdMs := int64(slo.FirstReplyHours) * 3_600_000
		resolutionThresholdMs := int64(slo.ResolutionHours) * 3_600_000
		inactivityThresholdMs := int64(slo.InactivityHours) * 3_600_000

		if topics[i].Outcome != "" || topics[i].FirstReplyAt != nil {
			// Resolved or replied: check first-reply SLO
			var actualFirstReply int64
			if topics[i].FirstReplyAt != nil {
				actualFirstReply = topics[i].FirstReplyAt.Sub(topics[i].CreatedAt).Milliseconds()
			} else if topics[i].Outcome == "" {
				actualFirstReply = now.Sub(topics[i].CreatedAt).Milliseconds()
			}
			if topics[i].FirstReplyAt != nil && actualFirstReply > firstReplyThresholdMs {
				groups.FirstReply = append(groups.FirstReply, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: firstReplyThresholdMs,
					ActualMs: actualFirstReply, ExcessMs: actualFirstReply - firstReplyThresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
		}

		if topics[i].Outcome == "" && topics[i].ReplyCount == 0 {
			// Unreplied: check first-reply and inactivity
			elapsed := now.Sub(topics[i].CreatedAt).Milliseconds()
			if elapsed > firstReplyThresholdMs {
				groups.FirstReply = append(groups.FirstReply, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: firstReplyThresholdMs,
					ActualMs: elapsed, ExcessMs: elapsed - firstReplyThresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
			if elapsed > inactivityThresholdMs {
				groups.Inactivity = append(groups.Inactivity, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: inactivityThresholdMs,
					ActualMs: elapsed, ExcessMs: elapsed - inactivityThresholdMs,
					ThresholdIsDefault: isDefault,
				})
			}
		}

		if topics[i].ResolvedAt != nil {
			// Check resolution SLO
			actualResolution := topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds()
			if actualResolution > resolutionThresholdMs {
				groups.Resolution = append(groups.Resolution, Violation{
					TopicID: topics[i].ID, TopicTitle: topics[i].Title, TopicURL: topics[i].TopicURL,
					Tag: tag, ThresholdMs: resolutionThresholdMs,
					ActualMs: actualResolution, ExcessMs: actualResolution - resolutionThresholdMs,
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

func strictestSLO(
	tags []string,
	resolved map[string]model.ResolvedTag,
) (string, model.SLOThresholds, bool) {
	bestTag := ""
	var bestSLO model.SLOThresholds
	bestIsDefault := true
	bestTotal := -1

	for _, tag := range tags {
		rt, ok := resolved[tag]
		if !ok {
			continue
		}
		total := rt.SLO.FirstReplyHours + rt.SLO.ResolutionHours + rt.SLO.InactivityHours
		if bestTotal < 0 || total < bestTotal {
			bestTag = tag
			bestSLO = rt.SLO
			bestIsDefault = rt.SLOIsDefault
			bestTotal = total
		}
	}
	return bestTag, bestSLO, bestIsDefault
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
