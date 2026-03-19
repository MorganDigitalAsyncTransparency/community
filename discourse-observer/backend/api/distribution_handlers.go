// Spec: specs/api/api-contract.md (AC-20, AC-21, AC-22, AC-23)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
	"github.com/code-community/discourse-observer/backend/model"
)

func (s *Server) handleDistributionVolume(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics, err := s.Store.QueryTopics(r.Context(), resolveQueryOpts(f, s.Now()))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	topics = applyTagFilter(topics, f, s.MonitoredTags())
	ranking := domain.TagVolumeRanking(topics)

	type item struct {
		Tag        string `json:"tag"`
		TopicCount int    `json:"topicCount"`
	}
	items := make([]item, len(ranking))
	for i, r := range ranking {
		items[i] = item{Tag: r.Tag, TopicCount: r.TopicCount}
	}
	respondJSON(w, items)
}

func (s *Server) handleDistributionResolution(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics, err := s.Store.QueryTopics(r.Context(), resolveQueryOpts(f, s.Now()))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	topics = applyTagFilter(topics, f, s.MonitoredTags())
	ranking := domain.TagResolutionRanking(topics)

	type item struct {
		Tag                string `json:"tag"`
		ResolvedCount      int    `json:"resolvedCount"`
		MedianResolutionMs *int64 `json:"medianResolutionMs"`
	}
	items := make([]item, len(ranking))
	for i, r := range ranking {
		items[i] = item{Tag: r.Tag, ResolvedCount: r.ResolvedCount, MedianResolutionMs: r.MedianResolutionMs}
	}
	respondJSON(w, items)
}

func (s *Server) handleDistributionBacklog(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics, err := s.Store.QueryTopics(r.Context(), resolveQueryOpts(f, s.Now()))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	topics = applyTagFilter(topics, f, s.MonitoredTags())
	ranking := domain.TagBacklogRanking(topics)

	type item struct {
		Tag       string `json:"tag"`
		OpenCount int    `json:"openCount"`
	}
	items := make([]item, len(ranking))
	for i, r := range ranking {
		items[i] = item{Tag: r.Tag, OpenCount: r.OpenCount}
	}
	respondJSON(w, items)
}

func (s *Server) handleDistributionBacklogTrend(w http.ResponseWriter, r *http.Request) {
	// AC-23: period filter does not apply, only tag filter
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Only tag filter — no time bounds
	tagOnlyOpts := model.QueryOpts{Tag: f.Tag}
	topics, err := s.Store.QueryTopics(r.Context(), tagOnlyOpts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	topics = applyTagFilter(topics, f, s.MonitoredTags())

	// Separate open from resolved for weekly backlog computation
	unrepliedTopics := domain.FilterUnreplied(topics)
	repliedOpen := domain.FilterRepliedOpen(topics, s.ResolvedTags)
	openAll := make([]model.Topic, 0, len(unrepliedTopics)+len(repliedOpen))
	openAll = append(openAll, unrepliedTopics...)
	openAll = append(openAll, repliedOpen...)

	trend := domain.ComputeWeeklyBacklog(topics, openAll)
	if trend == nil {
		trend = []domain.WeeklyBacklog{}
	}

	type item struct {
		WeekStart string `json:"weekStart"`
		Created   int    `json:"created"`
		Resolved  int    `json:"resolved"`
		StillOpen int    `json:"stillOpen"`
	}
	items := make([]item, len(trend))
	for i, w := range trend {
		items[i] = item{
			WeekStart: w.WeekStart, Created: w.Created,
			Resolved: w.Resolved, StillOpen: w.StillOpen,
		}
	}
	respondJSON(w, items)
}
