// Spec: specs/api/api-contract.md (AC-20, AC-21, AC-22, AC-23)
// Tests: backend/domain/tagdist_test.go
package domain

import (
	"sort"

	"github.com/code-community/discourse-observer/backend/model"
)

// TagVolume holds a tag's topic count for volume ranking.
type TagVolume struct {
	Tag        string
	TopicCount int
}

// TagVolumeRanking ranks tags by topic count (descending), then alphabetically.
func TagVolumeRanking(topics []model.Topic) []TagVolume {
	counts := make(map[string]int)
	for i := range topics {
		for _, tag := range topics[i].Tags {
			counts[tag]++
		}
	}
	ranking := make([]TagVolume, 0, len(counts))
	for tag, count := range counts {
		ranking = append(ranking, TagVolume{Tag: tag, TopicCount: count})
	}
	sort.Slice(ranking, func(i, j int) bool {
		if ranking[i].TopicCount != ranking[j].TopicCount {
			return ranking[i].TopicCount > ranking[j].TopicCount
		}
		return ranking[i].Tag < ranking[j].Tag
	})
	return ranking
}

// TagResolution holds a tag's median resolution time.
type TagResolution struct {
	Tag                string
	ResolvedCount      int
	MedianResolutionMs *int64
}

// TagResolutionRanking ranks tags by median resolution time (descending).
// Tags with no resolved topics sort last.
func TagResolutionRanking(topics []model.Topic) []TagResolution {
	grouped := make(map[string][]int64)
	counts := make(map[string]int)
	for i := range topics {
		if topics[i].ResolvedAt == nil {
			continue
		}
		dur := topics[i].ResolvedAt.Sub(topics[i].CreatedAt).Milliseconds()
		for _, tag := range topics[i].Tags {
			grouped[tag] = append(grouped[tag], dur)
			counts[tag]++
		}
	}

	// Also include tags from topics without resolution, for completeness
	for i := range topics {
		for _, tag := range topics[i].Tags {
			if _, exists := grouped[tag]; !exists {
				grouped[tag] = nil
			}
		}
	}

	ranking := make([]TagResolution, 0, len(grouped))
	for tag, durations := range grouped {
		ranking = append(ranking, TagResolution{
			Tag:                tag,
			ResolvedCount:      counts[tag],
			MedianResolutionMs: Median(durations),
		})
	}

	sort.Slice(ranking, func(i, j int) bool {
		mi, mj := ranking[i].MedianResolutionMs, ranking[j].MedianResolutionMs
		if mi == nil && mj == nil {
			return ranking[i].Tag < ranking[j].Tag
		}
		if mi == nil {
			return false
		}
		if mj == nil {
			return true
		}
		return *mi > *mj
	})
	return ranking
}

// TagBacklog holds a tag's open (unreplied) topic count.
type TagBacklog struct {
	Tag       string
	OpenCount int
}

// TagBacklogRanking ranks tags by open topic count (descending), then alphabetically.
func TagBacklogRanking(topics []model.Topic) []TagBacklog {
	unreplied := FilterUnreplied(topics)
	counts := make(map[string]int)
	for i := range unreplied {
		for _, tag := range unreplied[i].Tags {
			counts[tag]++
		}
	}
	ranking := make([]TagBacklog, 0, len(counts))
	for tag, count := range counts {
		ranking = append(ranking, TagBacklog{Tag: tag, OpenCount: count})
	}
	sort.Slice(ranking, func(i, j int) bool {
		if ranking[i].OpenCount != ranking[j].OpenCount {
			return ranking[i].OpenCount > ranking[j].OpenCount
		}
		return ranking[i].Tag < ranking[j].Tag
	})
	return ranking
}

// WeeklyBacklog holds one week's backlog trend data.
type WeeklyBacklog struct {
	WeekStart string
	Created   int
	Resolved  int
	StillOpen int
}

// ComputeWeeklyBacklog groups topics by ISO week and counts created/resolved/stillOpen.
// Returned newest-first.
func ComputeWeeklyBacklog(allTopics, openTopics []model.Topic) []WeeklyBacklog {
	if len(allTopics) == 0 {
		return nil
	}

	openSet := make(map[int]bool)
	for i := range openTopics {
		openSet[openTopics[i].ID] = true
	}

	type weekData struct {
		created   int
		resolved  int
		stillOpen int
	}
	weeks := make(map[string]*weekData)

	for i := range allTopics {
		key := DayString(MondayOf(allTopics[i].CreatedAt))
		w, ok := weeks[key]
		if !ok {
			w = &weekData{}
			weeks[key] = w
		}
		w.created++
		if openSet[allTopics[i].ID] {
			w.stillOpen++
		} else {
			w.resolved++
		}
	}

	result := make([]WeeklyBacklog, 0, len(weeks))
	for key, w := range weeks {
		result = append(result, WeeklyBacklog{
			WeekStart: key,
			Created:   w.created,
			Resolved:  w.resolved,
			StillOpen: w.stillOpen,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].WeekStart > result[j].WeekStart
	})
	return result
}
