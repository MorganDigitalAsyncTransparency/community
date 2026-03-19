package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

var filterNow = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

func makeTopics(daysAgo ...int) []model.Topic {
	topics := make([]model.Topic, len(daysAgo))
	for i, d := range daysAgo {
		topics[i] = model.Topic{
			ID:        i + 1,
			CreatedAt: filterNow.AddDate(0, 0, -d),
			Tags:      []string{"api"},
		}
	}
	return topics
}

func TestFilterByPeriod(t *testing.T) {
	topics := makeTopics(1, 5, 7, 10, 30, 45, 365, 400)

	tests := []struct {
		period string
		want   int
	}{
		{"7d", 3},
		{"30d", 5},
		{"1y", 7},
		{"all", 8},
	}
	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			got := FilterByPeriod(topics, tt.period, filterNow)
			if len(got) != tt.want {
				t.Errorf("period %s: got %d topics, want %d", tt.period, len(got), tt.want)
			}
		})
	}
}

func TestFilterByPeriodBoundary(t *testing.T) {
	topics := makeTopics(7)
	got := FilterByPeriod(topics, "7d", filterNow)
	if len(got) != 1 {
		t.Errorf("exactly 7 days old should be included, got %d", len(got))
	}
}

func TestFilterByDateRange(t *testing.T) {
	topics := makeTopics(1, 5, 10, 20)
	from := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	got := FilterByDateRange(topics, from, to)
	if len(got) != 3 {
		t.Errorf("date range: got %d topics, want 3", len(got))
	}
}

func TestFilterByTag(t *testing.T) {
	topics := []model.Topic{
		{ID: 1, Tags: []string{"api", "webhooks"}},
		{ID: 2, Tags: []string{"editor"}},
		{ID: 3, Tags: []string{"api"}},
		{ID: 4, Tags: []string{}},
	}
	got := FilterByTag(topics, "api")
	if len(got) != 2 {
		t.Errorf("tag filter: got %d, want 2", len(got))
	}
	got = FilterByTag(topics, "")
	if len(got) != 4 {
		t.Errorf("empty tag: got %d, want 4", len(got))
	}
}

func TestFilterUnreplied(t *testing.T) {
	topics := []model.Topic{
		{ID: 1, ReplyCount: 0, Outcome: ""},
		{ID: 2, ReplyCount: 1, Outcome: ""},
		{ID: 3, ReplyCount: 0, Outcome: "solved"},
	}
	got := FilterUnreplied(topics)
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("unreplied filter: got %d topics", len(got))
	}
}

func TestFilterUntagged(t *testing.T) {
	topics := []model.Topic{
		{ID: 1, Tags: []string{}},
		{ID: 2, Tags: []string{"api"}},
		{ID: 3, Tags: nil},
	}
	got := FilterUntagged(topics)
	if len(got) != 2 {
		t.Errorf("untagged filter: got %d, want 2", len(got))
	}
}
