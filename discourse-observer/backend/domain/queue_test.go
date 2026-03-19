package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestComputeQueueSummary(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{ID: 1, ReplyCount: 0, Outcome: "", Tags: []string{"api"},
			CreatedAt: now.Add(-48 * time.Hour)},
		{ID: 2, ReplyCount: 0, Outcome: "", Tags: []string{},
			CreatedAt: now.Add(-168 * time.Hour)},
		{ID: 3, ReplyCount: 1, Outcome: "solved", Tags: []string{"api"}},
	}
	result := ComputeQueueSummary(topics, now)
	if result.UnrepliedCount != 2 {
		t.Errorf("unreplied: got %d, want 2", result.UnrepliedCount)
	}
	if result.UntaggedCount != 1 {
		t.Errorf("untagged: got %d, want 1", result.UntaggedCount)
	}
	if result.OldestUnrepliedDays == nil || *result.OldestUnrepliedDays != 7 {
		t.Errorf("oldest: got %v, want 7", result.OldestUnrepliedDays)
	}
}

func TestComputeQueueSummaryEmpty(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	result := ComputeQueueSummary(nil, now)
	if result.UnrepliedCount != 0 || result.UntaggedCount != 0 {
		t.Errorf("expected zeros for empty input")
	}
	if result.OldestUnrepliedDays != nil {
		t.Errorf("expected nil oldest for empty input")
	}
}

func TestFindStalledTopics(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	resolved := map[string]model.ResolvedTag{
		"api": {StalledDays: 7, StalledDaysIsDefault: false},
	}

	topics := []model.Topic{
		// Stalled: 18 days since last activity, threshold 7
		{ID: 1, ReplyCount: 3, Tags: []string{"api"},
			CreatedAt:      now.Add(-20 * 24 * time.Hour),
			LastActivityAt: timePtr(now.Add(-18 * 24 * time.Hour))},
		// Not stalled: 3 days since activity, threshold 7
		{ID: 2, ReplyCount: 1, Tags: []string{"api"},
			CreatedAt:      now.Add(-30 * 24 * time.Hour),
			LastActivityAt: timePtr(now.Add(-3 * 24 * time.Hour))},
		// Not stalled: resolved
		{ID: 3, ReplyCount: 2, Outcome: "solved", Tags: []string{"api"}},
		// Not stalled: unreplied (replyCount 0)
		{ID: 4, ReplyCount: 0, Tags: []string{"api"},
			CreatedAt: now.Add(-20 * 24 * time.Hour)},
	}

	stalled := FindStalledTopics(topics, resolved, 7, now)
	if len(stalled) != 1 {
		t.Fatalf("got %d stalled, want 1", len(stalled))
	}
	if stalled[0].Topic.ID != 1 {
		t.Errorf("stalled topic ID: got %d, want 1", stalled[0].Topic.ID)
	}
	if stalled[0].DaysSinceActivity != 18 {
		t.Errorf("days since activity: got %d, want 18", stalled[0].DaysSinceActivity)
	}
	if stalled[0].StrictestTag == nil || *stalled[0].StrictestTag != "api" {
		t.Errorf("strictest tag: got %v, want api", stalled[0].StrictestTag)
	}
}

func TestStalledBoundary(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	resolved := map[string]model.ResolvedTag{
		"api": {StalledDays: 14},
	}
	topics := []model.Topic{
		// Exactly 14 days: NOT stalled
		{ID: 1, ReplyCount: 1, Tags: []string{"api"},
			CreatedAt:      now.Add(-20 * 24 * time.Hour),
			LastActivityAt: timePtr(now.Add(-14 * 24 * time.Hour))},
		// 15 days: stalled
		{ID: 2, ReplyCount: 1, Tags: []string{"api"},
			CreatedAt:      now.Add(-20 * 24 * time.Hour),
			LastActivityAt: timePtr(now.Add(-15 * 24 * time.Hour))},
	}
	stalled := FindStalledTopics(topics, resolved, 14, now)
	if len(stalled) != 1 || stalled[0].Topic.ID != 2 {
		t.Errorf("boundary: exactly 14d should NOT be stalled")
	}
}

func timePtr(t time.Time) *time.Time { return &t }
