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

// resolveQueryOpts converts parsed filters into QueryOpts with concrete
// time bounds suitable for database queries. Period is resolved to an
// absolute from/to range using now. Tag is passed through.
func resolveQueryOpts(f ParsedFilters, now time.Time) model.QueryOpts {
	opts := model.QueryOpts{Tag: f.Tag}

	if f.From != nil && f.To != nil {
		from := *f.From
		to := f.To.Add(24*time.Hour - time.Millisecond) // end of day inclusive
		opts.From = &from
		opts.To = &to
		return opts
	}

	days := periodToDays(f.Period)
	if days > 0 {
		from := now.AddDate(0, 0, -days)
		opts.From = &from
	}
	return opts
}

// periodToDays returns the number of lookback days for a period.
// Returns 0 for "all" (no time bound).
func periodToDays(period string) int {
	switch period {
	case "7d":
		return 7
	case "30d":
		return 30
	case "1y":
		return 365
	default:
		return 0
	}
}

// applyTagFilter applies in-memory tag filtering to topics already loaded
// from the store. When no tag is specified in the query opts, only topics
// with at least one monitored tag are included (AC-10).
func applyTagFilter(topics []model.Topic, f ParsedFilters, monitored map[string]bool) []model.Topic {
	if f.Tag != "" {
		return topics // already filtered by tag in the SQL query
	}
	return domain.FilterByMonitoredTags(topics, monitored)
}
