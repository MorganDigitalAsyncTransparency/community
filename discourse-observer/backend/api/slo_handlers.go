// Spec: specs/api/api-contract.md (AC-24, AC-25)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleSLOViolations(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())
	groups := domain.FindViolations(topics, s.ResolvedTags, s.Now())

	type violation struct {
		TopicID            int    `json:"topicId"`
		TopicTitle         string `json:"topicTitle"`
		TopicURL           string `json:"topicUrl"`
		Tag                string `json:"tag"`
		ThresholdMs        int64  `json:"thresholdMs"`
		ActualMs           int64  `json:"actualMs"`
		ExcessMs           int64  `json:"excessMs"`
		ThresholdIsDefault bool   `json:"thresholdIsDefault"`
	}
	toViolations := func(src []domain.Violation) []violation {
		out := make([]violation, len(src))
		for i, v := range src {
			out[i] = violation{
				TopicID: v.TopicID, TopicTitle: v.TopicTitle, TopicURL: v.TopicURL,
				Tag: v.Tag, ThresholdMs: v.ThresholdMs, ActualMs: v.ActualMs,
				ExcessMs: v.ExcessMs, ThresholdIsDefault: v.ThresholdIsDefault,
			}
		}
		return out
	}
	respondJSON(w, map[string]any{
		"firstReply": toViolations(groups.FirstReply),
		"resolution": toViolations(groups.Resolution),
		"inactivity": toViolations(groups.Inactivity),
	})
}

func (s *Server) handleSLOCompliance(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	topics := applyAllFilters(s.Topics, f, s.Now())
	compliance := domain.ComputeCompliance(topics, s.ResolvedTags, s.Now())

	type item struct {
		Tag                string `json:"tag"`
		FirstReplyPercent  *int   `json:"firstReplyPercent"`
		ResolutionPercent  *int   `json:"resolutionPercent"`
		InactivityPercent  *int   `json:"inactivityPercent"`
		ThresholdIsDefault bool   `json:"thresholdIsDefault"`
	}
	items := make([]item, len(compliance))
	for i, c := range compliance {
		items[i] = item{
			Tag: c.Tag, FirstReplyPercent: c.FirstReplyPercent,
			ResolutionPercent: c.ResolutionPercent, InactivityPercent: c.InactivityPercent,
			ThresholdIsDefault: c.ThresholdIsDefault,
		}
	}
	respondJSON(w, items)
}
