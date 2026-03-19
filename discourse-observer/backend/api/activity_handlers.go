// Spec: specs/api/api-contract.md (AC-26)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleActivityHeatmap(w http.ResponseWriter, r *http.Request) {
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
	heatmap := domain.BuildHeatmap(topics)

	type cell struct {
		Day   int `json:"day"`
		Hour  int `json:"hour"`
		Count int `json:"count"`
	}
	cells := make([][]cell, 7)
	for day := 0; day < 7; day++ {
		cells[day] = make([]cell, 24)
		for hour := 0; hour < 24; hour++ {
			c := heatmap.Cells[day][hour]
			cells[day][hour] = cell{Day: c.Day, Hour: c.Hour, Count: c.Count}
		}
	}
	respondJSON(w, map[string]any{
		"cells":    cells,
		"maxCount": heatmap.MaxCount,
	})
}
