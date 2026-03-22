// Spec: specs/api/tag-flows.md (TF-1 through TF-13)
package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestTagTransitions(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-20 * time.Hour),
				Detail: tagChangeDetail(nil, []string{"auth"})},
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
				Detail: tagChangeDetail([]string{"auth"}, []string{"auth", "sso"})},
		},
	}

	result := ComputeTagFlows(topics, events)

	if len(result.Transitions) != 2 {
		t.Fatalf("transitions = %d, want 2", len(result.Transitions))
	}
	// First transition: [] → [auth]
	tr0 := result.Transitions[0]
	if tr0.Count != 1 {
		t.Errorf("transition[0].count = %d, want 1", tr0.Count)
	}
}

func TestTagTransitionMedianDuration(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour)},
		{ID: 2, CreatedAt: now.Add(-24 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-20 * time.Hour),
			Detail: tagChangeDetail(nil, []string{"api"})}},
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-18 * time.Hour),
			Detail: tagChangeDetail(nil, []string{"api"})}},
	}

	result := ComputeTagFlows(topics, events)

	if len(result.Transitions) != 1 {
		t.Fatalf("transitions = %d, want 1", len(result.Transitions))
	}
	tr := result.Transitions[0]
	if tr.Count != 2 {
		t.Errorf("count = %d, want 2", tr.Count)
	}
	// Durations: topic1=4h, topic2=6h → median=5h
	if tr.MedianDurationHours == nil {
		t.Fatal("medianDurationHours is nil")
	}
	if *tr.MedianDurationHours != 5.0 {
		t.Errorf("medianDurationHours = %f, want 5.0", *tr.MedianDurationHours)
	}
}

func TestTagPairs(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-20 * time.Hour),
			Detail: tagChangeDetail(nil, []string{"auth", "sso"})}},
	}

	result := ComputeTagFlows(topics, events)

	if len(result.TagPairs) != 1 {
		t.Fatalf("tagPairs = %d, want 1", len(result.TagPairs))
	}
	pair := result.TagPairs[0]
	if pair.Tags[0] != "auth" || pair.Tags[1] != "sso" {
		t.Errorf("tags = %v, want [auth, sso]", pair.Tags)
	}
	if pair.Count != 1 {
		t.Errorf("count = %d, want 1", pair.Count)
	}
}

func TestTagFlowsSummary(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour)},
		{ID: 2, CreatedAt: now.Add(-24 * time.Hour)},
		{ID: 3, CreatedAt: now.Add(-24 * time.Hour)},
	}
	events := map[int][]model.TopicEvent{
		1: {
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-20 * time.Hour),
				Detail: tagChangeDetail(nil, []string{"api"})},
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
				Detail: tagChangeDetail([]string{"api"}, []string{"auth"})},
		},
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-18 * time.Hour),
			Detail: tagChangeDetail(nil, []string{"api"})}},
		// topic 3 has no events
	}

	result := ComputeTagFlows(topics, events)

	s := result.Summary
	if s.TotalTopics != 3 {
		t.Errorf("totalTopics = %d, want 3", s.TotalTopics)
	}
	if s.TopicsWithTagChanges != 2 {
		t.Errorf("topicsWithTagChanges = %d, want 2", s.TopicsWithTagChanges)
	}
	// Change counts: topic1=2, topic2=1 → median=1 (truncated avg of 1 and 2 = 1)
	if s.MedianChangesPerTopic == nil {
		t.Fatal("medianChangesPerTopic is nil")
	}
	if s.MostCommonFirstTag == nil || *s.MostCommonFirstTag != "api" {
		t.Errorf("mostCommonFirstTag = %v, want api", s.MostCommonFirstTag)
	}
}

func TestTagFlowsEmptyDataset(t *testing.T) {
	result := ComputeTagFlows(nil, nil)

	if len(result.Transitions) != 0 {
		t.Errorf("transitions should be empty, got %d", len(result.Transitions))
	}
	if len(result.TagPairs) != 0 {
		t.Errorf("tagPairs should be empty, got %d", len(result.TagPairs))
	}
	if result.Summary.TopicsWithTagChanges != 0 {
		t.Errorf("topicsWithTagChanges should be 0")
	}
	if result.Summary.MedianChangesPerTopic != nil {
		t.Error("medianChangesPerTopic should be nil")
	}
	if result.Summary.MostCommonFirstTag != nil {
		t.Error("mostCommonFirstTag should be nil")
	}
	if result.Summary.MostUnstableTag != nil {
		t.Error("mostUnstableTag should be nil")
	}
}
