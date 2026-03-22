// Spec: specs/api/triage-time.md (TT-8, TT-9, TT-10)
// Tests: backend/api/triage-time_contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleTriageTime(w http.ResponseWriter, r *http.Request) {
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

	events, err := s.Events.LoadAllTopicEvents(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "event query failed")
		return
	}

	result := domain.ComputeTriageTime(topics, events)

	type overallEntry struct {
		MedianHours *float64 `json:"medianHours"`
		Count       int      `json:"count"`
	}
	type tagEntry struct {
		Tag         string   `json:"tag"`
		MedianHours *float64 `json:"medianHours"`
		Count       int      `json:"count"`
	}

	byTag := make([]tagEntry, len(result.ByTag))
	for i, e := range result.ByTag {
		byTag[i] = tagEntry{Tag: e.Tag, MedianHours: e.MedianHours, Count: e.Count}
	}
	respondJSON(w, map[string]any{
		"overall": overallEntry{MedianHours: result.MedianHours, Count: result.Count},
		"byTag":   byTag,
	})
}
