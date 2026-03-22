// Spec: specs/api/tag-flows.md (TF-14, TF-15, TF-16)
// Tests: backend/api/tag-flows_contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleTagFlows(w http.ResponseWriter, r *http.Request) {
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

	result := domain.ComputeTagFlows(topics, events)
	respondJSON(w, result)
}
