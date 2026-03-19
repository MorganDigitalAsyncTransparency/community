// Spec: specs/api/api-contract.md (AC-12, AC-13, AC-14, AC-15)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"
	"sort"

	"github.com/code-community/discourse-observer/backend/domain"
)

func (s *Server) handleQueueSummary(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	opts := resolveQueryOpts(f, s.Now())

	// Unreplied count uses monitored-tag-filtered topics
	monitoredTopics, err := s.Store.QueryTopics(r.Context(), opts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	monitoredTopics = applyTagFilter(monitoredTopics, f, s.MonitoredTags())

	// Untagged count uses time-only filtered topics (untagged topics have no
	// tags, so tag filtering would exclude them — AC-30)
	noTagOpts := opts
	noTagOpts.Tag = ""
	timeOnlyTopics, err := s.Store.QueryTopics(r.Context(), noTagOpts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}

	unreplied := domain.FilterUnreplied(monitoredTopics)
	untagged := domain.FilterUntagged(timeOnlyTopics)

	var oldest *int
	if len(unreplied) > 0 {
		maxAge := 0
		now := s.Now()
		for i := range unreplied {
			days := int(now.Sub(unreplied[i].CreatedAt).Hours() / 24)
			if days > maxAge {
				maxAge = days
			}
		}
		oldest = &maxAge
	}

	respondJSON(w, map[string]any{
		"unrepliedCount":         len(unreplied),
		"untaggedCount":          len(untagged),
		"oldestUnrepliedAgeDays": oldest,
	})
}

func (s *Server) handleQueueUnreplied(w http.ResponseWriter, r *http.Request) {
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
	unreplied := domain.FilterUnreplied(topics)

	sort.Slice(unreplied, func(i, j int) bool {
		return unreplied[i].CreatedAt.Before(unreplied[j].CreatedAt)
	})

	type item struct {
		ID        int      `json:"id"`
		Title     string   `json:"title"`
		CreatedAt string   `json:"createdAt"`
		Tags      []string `json:"tags"`
		TopicURL  string   `json:"topicUrl"`
	}
	items := make([]item, len(unreplied))
	for i := range unreplied {
		tags := unreplied[i].Tags
		if tags == nil {
			tags = []string{}
		}
		items[i] = item{
			ID: unreplied[i].ID, Title: unreplied[i].Title,
			CreatedAt: unreplied[i].CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Tags:      tags, TopicURL: unreplied[i].TopicURL,
		}
	}
	respondJSON(w, items)
}

func (s *Server) handleQueueUntagged(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilters(r, &s.TagConfig)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Tag filter does not apply to untagged endpoint (AC-30)
	f.Tag = ""
	topics, err := s.Store.QueryTopics(r.Context(), resolveQueryOpts(f, s.Now()))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	untagged := domain.FilterUntagged(topics)

	sort.Slice(untagged, func(i, j int) bool {
		return untagged[i].CreatedAt.Before(untagged[j].CreatedAt)
	})

	type item struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		CreatedAt    string `json:"createdAt"`
		CategoryName string `json:"categoryName"`
		TopicURL     string `json:"topicUrl"`
	}
	items := make([]item, len(untagged))
	for i := range untagged {
		items[i] = item{
			ID: untagged[i].ID, Title: untagged[i].Title,
			CreatedAt:    untagged[i].CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			CategoryName: untagged[i].CategoryName, TopicURL: untagged[i].TopicURL,
		}
	}
	respondJSON(w, items)
}

func (s *Server) handleQueueStalled(w http.ResponseWriter, r *http.Request) {
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
	stalled := domain.FindStalledTopics(topics, s.ResolvedTags, s.TagConfig.Defaults.StalledDays, s.Now())

	type item struct {
		ID                    int      `json:"id"`
		Title                 string   `json:"title"`
		CreatedAt             string   `json:"createdAt"`
		Tags                  []string `json:"tags"`
		TopicURL              string   `json:"topicUrl"`
		StrictestTag          *string  `json:"strictestTag"`
		ThresholdDays         int      `json:"thresholdDays"`
		ThresholdIsDefault    bool     `json:"thresholdIsDefault"`
		DaysSinceLastActivity int      `json:"daysSinceLastActivity"`
	}
	items := make([]item, len(stalled))
	for i := range stalled {
		tags := stalled[i].Topic.Tags
		if tags == nil {
			tags = []string{}
		}
		items[i] = item{
			ID: stalled[i].Topic.ID, Title: stalled[i].Topic.Title,
			CreatedAt: stalled[i].Topic.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Tags:      tags, TopicURL: stalled[i].Topic.TopicURL,
			StrictestTag:          stalled[i].StrictestTag,
			ThresholdDays:         stalled[i].ThresholdDays,
			ThresholdIsDefault:    stalled[i].ThresholdIsDefault,
			DaysSinceLastActivity: stalled[i].DaysSinceActivity,
		}
	}
	respondJSON(w, items)
}
