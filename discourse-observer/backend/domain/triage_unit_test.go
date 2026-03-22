// Spec: specs/api/triage-time.md (TT-1, TT-2, TT-3, TT-4, TT-5, TT-6, TT-12)
package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func tagChangeDetail(prev, curr []string) string {
	d, _ := json.Marshal(model.RevisionTagChange{Previous: prev, Current: curr})
	return string(d)
}

func TestTriageTimeMedian(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-6 * time.Hour)},
		{ID: 2, CreatedAt: now.Add(-10 * time.Hour)},
		{ID: 3, CreatedAt: now.Add(-8 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-4 * time.Hour), Detail: tagChangeDetail(nil, []string{"auth"})}},
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-6 * time.Hour), Detail: tagChangeDetail(nil, []string{"api"})}},
		3: {{TopicID: 3, EventType: "tag_change", HappenedAt: now.Add(-2 * time.Hour), Detail: tagChangeDetail(nil, []string{"auth"})}},
	}

	result := ComputeTriageTime(topics, events)

	if result.Count != 3 {
		t.Errorf("count = %d, want 3", result.Count)
	}
	// Durations: topic1=2h, topic2=4h, topic3=6h → median=4h
	if result.MedianHours == nil {
		t.Fatal("medianHours is nil, want 4.0")
	}
	if *result.MedianHours != 4.0 {
		t.Errorf("medianHours = %f, want 4.0", *result.MedianHours)
	}
}

func TestTriageTimeByTag(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-6 * time.Hour)},
		{ID: 2, CreatedAt: now.Add(-10 * time.Hour)},
		{ID: 3, CreatedAt: now.Add(-8 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-4 * time.Hour), Detail: tagChangeDetail(nil, []string{"auth"})}},
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-6 * time.Hour), Detail: tagChangeDetail(nil, []string{"api"})}},
		3: {{TopicID: 3, EventType: "tag_change", HappenedAt: now.Add(-2 * time.Hour), Detail: tagChangeDetail(nil, []string{"auth"})}},
	}

	result := ComputeTriageTime(topics, events)

	if len(result.ByTag) != 2 {
		t.Fatalf("byTag length = %d, want 2", len(result.ByTag))
	}
	// auth has 2 topics (count-descending sort), api has 1
	if result.ByTag[0].Tag != "auth" {
		t.Errorf("byTag[0].tag = %q, want auth", result.ByTag[0].Tag)
	}
	if result.ByTag[0].Count != 2 {
		t.Errorf("byTag[0].count = %d, want 2", result.ByTag[0].Count)
	}
	if result.ByTag[1].Tag != "api" {
		t.Errorf("byTag[1].tag = %q, want api", result.ByTag[1].Tag)
	}
}

func TestTriageTimeNoEvents(t *testing.T) {
	topics := []model.Topic{
		{ID: 1, CreatedAt: time.Now()},
	}
	events := map[int][]model.TopicEvent{}

	result := ComputeTriageTime(topics, events)

	if result.Count != 0 {
		t.Errorf("count = %d, want 0", result.Count)
	}
	if result.MedianHours != nil {
		t.Error("medianHours should be nil for empty results")
	}
	if len(result.ByTag) != 0 {
		t.Errorf("byTag should be empty, got %d", len(result.ByTag))
	}
}

func TestTriageTimeDetailParsing(t *testing.T) {
	tests := []struct {
		name    string
		detail  string
		wantTag string
	}{
		{
			name:    "new tag added from empty",
			detail:  tagChangeDetail(nil, []string{"authentication"}),
			wantTag: "authentication",
		},
		{
			name:    "tag added alongside existing",
			detail:  tagChangeDetail([]string{"api"}, []string{"api", "sso"}),
			wantTag: "sso",
		},
		{
			name:    "all tags already existed",
			detail:  tagChangeDetail([]string{"api"}, []string{"api"}),
			wantTag: "api",
		},
		{
			name:    "invalid JSON",
			detail:  "not json",
			wantTag: "",
		},
		{
			name:    "empty current",
			detail:  tagChangeDetail([]string{"api"}, nil),
			wantTag: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseFirstTag(tc.detail)
			if got != tc.wantTag {
				t.Errorf("parseFirstTag = %q, want %q", got, tc.wantTag)
			}
		})
	}
}
