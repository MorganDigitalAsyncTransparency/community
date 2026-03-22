// Spec: specs/api/escalations.md (EP-1, EP-2, EP-3, EP-4, EP-5, EP-6, EP-7)
package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestEscalationDetection(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	replied := now.Add(-20 * time.Hour)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 2, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 3, CreatedAt: now.Add(-24 * time.Hour)}, // no reply
	}
	events := map[int][]model.TopicEvent{
		// Topic 1: tag change AFTER reply → escalation
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
			Detail: tagChangeDetail([]string{"api"}, []string{"api", "auth"})}},
		// Topic 2: tag change BEFORE reply → not escalation
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-22 * time.Hour),
			Detail: tagChangeDetail(nil, []string{"api"})}},
	}

	result := ComputeEscalations(topics, events)

	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
}

func TestEscalationRate(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	replied := now.Add(-20 * time.Hour)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 2, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 3, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 4, CreatedAt: now.Add(-24 * time.Hour)}, // no reply, excluded
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
			Detail: tagChangeDetail([]string{"api"}, []string{"auth"})}},
	}

	result := ComputeEscalations(topics, events)

	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
	if result.Rate == nil {
		t.Fatal("rate is nil")
	}
	// 1 escalation out of 3 replied topics
	expected := 1.0 / 3.0
	if *result.Rate < expected-0.01 || *result.Rate > expected+0.01 {
		t.Errorf("rate = %f, want ~%f", *result.Rate, expected)
	}
}

func TestEscalationByPeriod(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	replied := now.Add(-200 * time.Hour)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-240 * time.Hour), FirstReplyAt: &replied},
	}
	events := map[int][]model.TopicEvent{
		1: {
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-100 * time.Hour),
				Detail: tagChangeDetail([]string{"api"}, []string{"auth"})},
			{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
				Detail: tagChangeDetail([]string{"auth"}, []string{"auth", "sso"})},
		},
	}

	result := ComputeEscalations(topics, events)

	if len(result.ByPeriod) == 0 {
		t.Fatal("byPeriod should not be empty")
	}
	// Both events should have ISO week labels
	for _, p := range result.ByPeriod {
		if p.Period == "" {
			t.Error("period label should not be empty")
		}
		if p.Count == 0 {
			t.Error("count should not be 0")
		}
	}
}

func TestEscalationPatterns(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	replied := now.Add(-20 * time.Hour)

	topics := []model.Topic{
		{ID: 1, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
		{ID: 2, CreatedAt: now.Add(-24 * time.Hour), FirstReplyAt: &replied},
	}
	events := map[int][]model.TopicEvent{
		1: {{TopicID: 1, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
			Detail: tagChangeDetail([]string{"api"}, []string{"api", "auth"})}},
		2: {{TopicID: 2, EventType: "tag_change", HappenedAt: now.Add(-10 * time.Hour),
			Detail: tagChangeDetail([]string{"api"}, []string{"api", "auth"})}},
	}

	result := ComputeEscalations(topics, events)

	if len(result.CommonPatterns) != 1 {
		t.Fatalf("commonPatterns = %d, want 1", len(result.CommonPatterns))
	}
	p := result.CommonPatterns[0]
	if p.Count != 2 {
		t.Errorf("count = %d, want 2", p.Count)
	}
	if len(p.OriginalTags) != 1 || p.OriginalTags[0] != "api" {
		t.Errorf("originalTags = %v, want [api]", p.OriginalTags)
	}
	if len(p.AddedAfterReply) != 1 || p.AddedAfterReply[0] != "auth" {
		t.Errorf("addedAfterReply = %v, want [auth]", p.AddedAfterReply)
	}
}

func TestNoEscalationsWhenNoRevisions(t *testing.T) {
	result := ComputeEscalations(nil, nil)

	if result.Total != 0 {
		t.Errorf("total = %d, want 0", result.Total)
	}
	if result.Rate != nil {
		t.Error("rate should be nil with no topics")
	}
	if len(result.ByPeriod) != 0 {
		t.Errorf("byPeriod should be empty, got %d", len(result.ByPeriod))
	}
	if len(result.CommonPatterns) != 0 {
		t.Errorf("commonPatterns should be empty, got %d", len(result.CommonPatterns))
	}
}
