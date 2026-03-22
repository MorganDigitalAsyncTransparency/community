// Spec: specs/api/tag-flows.md (TF-1 through TF-13, TF-17)
// Tests: backend/domain/tagflows_unit_test.go
package domain

import (
	"encoding/json"
	"sort"

	"github.com/code-community/discourse-observer/backend/model"
)

// TagFlowResult holds the computed tag flow metrics.
type TagFlowResult struct {
	Transitions []Transition `json:"transitions"`
	TagPairs    []TagPair    `json:"tagPairs"`
	Summary     FlowSummary  `json:"summary"`
}

// Transition represents an aggregated from→to tag set change.
type Transition struct {
	From                []string `json:"from"`
	To                  []string `json:"to"`
	Count               int      `json:"count"`
	MedianDurationHours *float64 `json:"medianDurationHours"`
}

// TagPair represents two tags that co-occur via revision.
type TagPair struct {
	Tags  [2]string `json:"tags"`
	Count int       `json:"count"`
}

// FlowSummary holds aggregate tag flow statistics.
type FlowSummary struct {
	TopicsWithTagChanges  int      `json:"topicsWithTagChanges"`
	TotalTopics           int      `json:"totalTopics"`
	MedianChangesPerTopic *float64 `json:"medianChangesPerTopic"`
	MostCommonFirstTag    *string  `json:"mostCommonFirstTag"`
	MostUnstableTag       *string  `json:"mostUnstableTag"`
}

// ComputeTagFlows calculates tag flow metrics from topics and their events.
func ComputeTagFlows(topics []model.Topic, events map[int][]model.TopicEvent) TagFlowResult {
	result := TagFlowResult{
		Transitions: []Transition{},
		TagPairs:    []TagPair{},
		Summary:     FlowSummary{TotalTopics: len(topics)},
	}

	transKey := func(from, to []string) string {
		f, _ := json.Marshal(from)
		t, _ := json.Marshal(to)
		return string(f) + "|" + string(t)
	}

	type transAccum struct {
		from, to    []string
		durationsMs []int64
	}

	transMap := map[string]*transAccum{}
	pairCounts := map[[2]string]int{}
	changeCounts := []int64{}
	firstTagCounts := map[string]int{}
	tagOps := map[string]int{}    // tag → total add+remove ops
	tagTopics := map[string]int{} // tag → number of topics it appears in

	for i := range topics {
		t := &topics[i]
		topicEvents := filterTagChanges(events[t.ID])
		if len(topicEvents) == 0 {
			continue
		}

		result.Summary.TopicsWithTagChanges++
		changeCounts = append(changeCounts, int64(len(topicEvents)))

		seenTags := map[string]bool{}

		for j, ev := range topicEvents {
			change := parseTagChange(ev.Detail)
			if change == nil {
				continue
			}

			from := normalized(change.Previous)
			to := normalized(change.Current)

			var durationMs int64
			if j == 0 {
				durationMs = ev.HappenedAt.Sub(t.CreatedAt).Milliseconds()
			} else {
				durationMs = ev.HappenedAt.Sub(topicEvents[j-1].HappenedAt).Milliseconds()
			}

			key := transKey(from, to)
			acc, ok := transMap[key]
			if !ok {
				acc = &transAccum{from: from, to: to}
				transMap[key] = acc
			}
			acc.durationsMs = append(acc.durationsMs, durationMs)

			collectPairs(change, pairCounts)

			if j == 0 {
				firstTag := firstNewTag(change)
				if firstTag != "" {
					firstTagCounts[firstTag]++
				}
			}

			collectTagOps(change, seenTags, tagOps, tagTopics)
		}
	}

	for _, acc := range transMap {
		result.Transitions = append(result.Transitions, Transition{
			From:                acc.from,
			To:                  acc.to,
			Count:               len(acc.durationsMs),
			MedianDurationHours: msToHours(Median(acc.durationsMs)),
		})
	}
	sort.Slice(result.Transitions, func(i, j int) bool {
		return result.Transitions[i].Count > result.Transitions[j].Count
	})

	for pair, count := range pairCounts {
		result.TagPairs = append(result.TagPairs, TagPair{Tags: pair, Count: count})
	}
	sort.Slice(result.TagPairs, func(i, j int) bool {
		if result.TagPairs[i].Count != result.TagPairs[j].Count {
			return result.TagPairs[i].Count > result.TagPairs[j].Count
		}
		a := result.TagPairs[i].Tags
		b := result.TagPairs[j].Tags
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		return a[1] < b[1]
	})

	result.Summary.MedianChangesPerTopic = medianFloat(changeCounts)
	result.Summary.MostCommonFirstTag = maxKey(firstTagCounts)
	result.Summary.MostUnstableTag = mostUnstable(tagOps, tagTopics)

	return result
}

func filterTagChanges(events []model.TopicEvent) []model.TopicEvent {
	var result []model.TopicEvent
	for _, e := range events {
		if e.EventType == "tag_change" {
			result = append(result, e)
		}
	}
	return result
}

func parseTagChange(detail string) *model.RevisionTagChange {
	var change model.RevisionTagChange
	if err := json.Unmarshal([]byte(detail), &change); err != nil {
		return nil
	}
	return &change
}

func normalized(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	out := make([]string, len(tags))
	copy(out, tags)
	sort.Strings(out)
	return out
}

func firstNewTag(change *model.RevisionTagChange) string {
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

func collectPairs(change *model.RevisionTagChange, counts map[[2]string]int) {
	prevSet := toSet(change.Previous)
	curr := normalized(change.Current)

	for i := 0; i < len(curr); i++ {
		for j := i + 1; j < len(curr); j++ {
			bothExisted := prevSet[curr[i]] && prevSet[curr[j]]
			if !bothExisted {
				pair := [2]string{curr[i], curr[j]}
				counts[pair]++
			}
		}
	}
}

func collectTagOps(change *model.RevisionTagChange, seen map[string]bool, ops, topics map[string]int) {
	prevSet := toSet(change.Previous)
	currSet := toSet(change.Current)

	for _, tag := range change.Current {
		if !prevSet[tag] {
			ops[tag]++
			if !seen[tag] {
				seen[tag] = true
				topics[tag]++
			}
		}
	}
	for _, tag := range change.Previous {
		if !currSet[tag] {
			ops[tag]++
			if !seen[tag] {
				seen[tag] = true
				topics[tag]++
			}
		}
	}
}

func mostUnstable(ops, topics map[string]int) *string {
	if len(ops) == 0 {
		return nil
	}
	var best string
	var bestAvg float64
	for tag, count := range ops {
		t := topics[tag]
		if t == 0 {
			continue
		}
		avg := float64(count) / float64(t)
		if avg > bestAvg || (avg == bestAvg && tag < best) {
			bestAvg = avg
			best = tag
		}
	}
	return &best
}

func maxKey(counts map[string]int) *string {
	if len(counts) == 0 {
		return nil
	}
	var best string
	var bestCount int
	for k, v := range counts {
		if v > bestCount || (v == bestCount && k < best) {
			bestCount = v
			best = k
		}
	}
	return &best
}

func medianFloat(values []int64) *float64 {
	m := Median(values)
	if m == nil {
		return nil
	}
	v := float64(*m)
	return &v
}
