// Spec: specs/api/api-contract.md (AC-8, AC-9, AC-10, AC-11)
// Tests: backend/api/filters_test.go
package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/code-community/discourse-observer/backend/domain"
	"github.com/code-community/discourse-observer/backend/model"
)

// ParsedFilters holds validated filter parameters.
type ParsedFilters struct {
	Period string
	From   *time.Time
	To     *time.Time
	Tag    string
}

var validPeriods = map[string]bool{
	"7d": true, "30d": true, "1y": true, "all": true,
}

// parseFilters extracts and validates filter query parameters.
func parseFilters(r *http.Request, cfg *model.TagConfig) (ParsedFilters, error) {
	q := r.URL.Query()
	f := ParsedFilters{Period: "all"}

	if p := q.Get("period"); p != "" {
		if !validPeriods[p] {
			return f, fmt.Errorf("invalid period: %q (valid: 7d, 30d, 1y, all)", p)
		}
		f.Period = p
	}

	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr != "" || toStr != "" {
		if fromStr == "" || toStr == "" {
			return f, fmt.Errorf("both 'from' and 'to' must be provided together")
		}
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return f, fmt.Errorf("invalid 'from' date: %q (expected YYYY-MM-DD)", fromStr)
		}
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return f, fmt.Errorf("invalid 'to' date: %q (expected YYYY-MM-DD)", toStr)
		}
		if from.After(to) {
			return f, fmt.Errorf("'from' (%s) must not be after 'to' (%s)", fromStr, toStr)
		}
		f.From = &from
		f.To = &to
	}

	if tag := q.Get("tag"); tag != "" {
		if _, ok := cfg.Tags[tag]; !ok {
			return f, fmt.Errorf("unknown tag: %q", tag)
		}
		f.Tag = tag
	}

	return f, nil
}

// applyTimeFilters applies period or date range filters to topics.
func applyTimeFilters(topics []model.Topic, f ParsedFilters, now time.Time) []model.Topic {
	if f.From != nil && f.To != nil {
		return domain.FilterByDateRange(topics, *f.From, *f.To)
	}
	return domain.FilterByPeriod(topics, f.Period, now)
}

// applyTagFilter applies the tag filter to topics.
// When a specific tag is requested, only that tag's topics are returned.
// When no tag is specified, only topics with at least one monitored tag
// are returned (AC-10).
func applyTagFilter(topics []model.Topic, f ParsedFilters, monitored map[string]bool) []model.Topic {
	if f.Tag != "" {
		return domain.FilterByTag(topics, f.Tag)
	}
	return domain.FilterByMonitoredTags(topics, monitored)
}

// applyAllFilters applies both time and tag filters.
func applyAllFilters(topics []model.Topic, f ParsedFilters, now time.Time, monitored map[string]bool) []model.Topic {
	topics = applyTimeFilters(topics, f, now)
	return applyTagFilter(topics, f, monitored)
}
