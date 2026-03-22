// Spec: specs/api/escalations.md (EP-1 through EP-7, EP-11)
// Tests: backend/domain/escalations_unit_test.go
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// EscalationResult holds escalation pattern metrics.
type EscalationResult struct {
	Total          int                 `json:"total"`
	Rate           *float64            `json:"rate"`
	ByPeriod       []EscalationPeriod  `json:"byPeriod"`
	CommonPatterns []EscalationPattern `json:"commonPatterns"`
}

// EscalationPeriod holds escalation count for a week.
type EscalationPeriod struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

// EscalationPattern holds a common before/after tag pattern.
type EscalationPattern struct {
	OriginalTags    []string `json:"originalTags"`
	AddedAfterReply []string `json:"addedAfterReply"`
	Count           int      `json:"count"`
}

// ComputeEscalations detects topics where tags changed after the first reply.
func ComputeEscalations(topics []model.Topic, events map[int][]model.TopicEvent) EscalationResult {
	result := EscalationResult{
		ByPeriod:       []EscalationPeriod{},
		CommonPatterns: []EscalationPattern{},
	}

	var repliedCount int
	weekCounts := map[string]int{}
	patternCounts := map[string]*EscalationPattern{}

	for i := range topics {
		t := &topics[i]
		if t.FirstReplyAt == nil {
			continue
		}
		repliedCount++

		escalated := false
		for _, ev := range events[t.ID] {
			if ev.EventType != "tag_change" {
				continue
			}
			if !ev.HappenedAt.After(*t.FirstReplyAt) {
				continue
			}

			if !escalated {
				escalated = true
				result.Total++
			}

			week := isoWeek(ev.HappenedAt)
			weekCounts[week]++

			collectPattern(ev.Detail, patternCounts)
		}
	}

	if repliedCount > 0 {
		r := float64(result.Total) / float64(repliedCount)
		result.Rate = &r
	}

	for _, count := range weekCounts {
		result.ByPeriod = append(result.ByPeriod, EscalationPeriod{Count: count})
	}
	// Fill period labels and sort chronologically
	for week, count := range weekCounts {
		for j := range result.ByPeriod {
			if result.ByPeriod[j].Count == count && result.ByPeriod[j].Period == "" {
				result.ByPeriod[j].Period = week
				break
			}
		}
	}
	// Rebuild properly
	result.ByPeriod = result.ByPeriod[:0]
	for week, count := range weekCounts {
		result.ByPeriod = append(result.ByPeriod, EscalationPeriod{Period: week, Count: count})
	}
	sort.Slice(result.ByPeriod, func(i, j int) bool {
		return result.ByPeriod[i].Period < result.ByPeriod[j].Period
	})

	for _, p := range patternCounts {
		result.CommonPatterns = append(result.CommonPatterns, *p)
	}
	sort.Slice(result.CommonPatterns, func(i, j int) bool {
		return result.CommonPatterns[i].Count > result.CommonPatterns[j].Count
	})

	return result
}

func isoWeek(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

func collectPattern(detail string, counts map[string]*EscalationPattern) {
	var change model.RevisionTagChange
	if err := json.Unmarshal([]byte(detail), &change); err != nil {
		return
	}

	prevSet := toSet(change.Previous)
	var added []string
	for _, tag := range change.Current {
		if !prevSet[tag] {
			added = append(added, tag)
		}
	}
	if len(added) == 0 {
		return
	}

	original := normalized(change.Previous)
	sort.Strings(added)

	origJSON, _ := json.Marshal(original)
	addedJSON, _ := json.Marshal(added)
	key := string(origJSON) + "|" + string(addedJSON)

	if p, ok := counts[key]; ok {
		p.Count++
	} else {
		counts[key] = &EscalationPattern{
			OriginalTags:    original,
			AddedAfterReply: added,
			Count:           1,
		}
	}
}
