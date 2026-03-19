package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestTagVolumeRanking(t *testing.T) {
	topics := []model.Topic{
		{Tags: []string{"api", "webhooks"}},
		{Tags: []string{"api"}},
		{Tags: []string{"editor"}},
		{Tags: []string{}},
	}
	ranking := TagVolumeRanking(topics)
	if len(ranking) != 3 {
		t.Fatalf("got %d tags, want 3", len(ranking))
	}
	if ranking[0].Tag != "api" || ranking[0].TopicCount != 2 {
		t.Errorf("rank 0: got %s=%d", ranking[0].Tag, ranking[0].TopicCount)
	}
}

func TestTagVolumeRankingEmpty(t *testing.T) {
	ranking := TagVolumeRanking(nil)
	if len(ranking) != 0 {
		t.Errorf("got %d, want 0", len(ranking))
	}
}

func TestTagVolumeRankingTiesAlphabetical(t *testing.T) {
	topics := []model.Topic{
		{Tags: []string{"beta"}},
		{Tags: []string{"alpha"}},
	}
	ranking := TagVolumeRanking(topics)
	if ranking[0].Tag != "alpha" {
		t.Errorf("equal counts: got %s first, want alpha", ranking[0].Tag)
	}
}

func TestTagResolutionRanking(t *testing.T) {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{Tags: []string{"api"}, CreatedAt: base,
			ResolvedAt: timePtr(base.Add(48 * time.Hour))},
		{Tags: []string{"api"}, CreatedAt: base,
			ResolvedAt: timePtr(base.Add(96 * time.Hour))},
		{Tags: []string{"editor"}, CreatedAt: base},
	}
	ranking := TagResolutionRanking(topics)
	if len(ranking) != 2 {
		t.Fatalf("got %d, want 2", len(ranking))
	}
	// api has median 72h, editor has nil → editor sorts last
	if ranking[0].Tag != "api" {
		t.Errorf("api should be first (has data), got %s", ranking[0].Tag)
	}
	if ranking[0].ResolvedCount != 2 {
		t.Errorf("api resolved count: got %d, want 2", ranking[0].ResolvedCount)
	}
	if ranking[1].MedianResolutionMs != nil {
		t.Errorf("editor should have nil median")
	}
}

func TestTagBacklogRanking(t *testing.T) {
	topics := []model.Topic{
		{Tags: []string{"api"}, ReplyCount: 0, Outcome: ""},
		{Tags: []string{"api"}, ReplyCount: 0, Outcome: ""},
		{Tags: []string{"editor"}, ReplyCount: 0, Outcome: ""},
		{Tags: []string{"editor"}, ReplyCount: 1, Outcome: ""},
	}
	ranking := TagBacklogRanking(topics)
	if len(ranking) != 2 {
		t.Fatalf("got %d, want 2", len(ranking))
	}
	if ranking[0].Tag != "api" || ranking[0].OpenCount != 2 {
		t.Errorf("api should have 2 open, got %s=%d", ranking[0].Tag, ranking[0].OpenCount)
	}
}

func TestComputeWeeklyBacklog(t *testing.T) {
	mon := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	wed := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	allTopics := []model.Topic{
		{ID: 1, CreatedAt: mon},
		{ID: 2, CreatedAt: mon},
		{ID: 3, CreatedAt: wed},
	}
	openTopics := []model.Topic{
		{ID: 1, CreatedAt: mon},
	}
	trend := ComputeWeeklyBacklog(allTopics, openTopics)
	if len(trend) != 1 {
		t.Fatalf("same week: got %d rows, want 1", len(trend))
	}
	if trend[0].Created != 3 || trend[0].Resolved != 2 || trend[0].StillOpen != 1 {
		t.Errorf("week: created=%d resolved=%d open=%d",
			trend[0].Created, trend[0].Resolved, trend[0].StillOpen)
	}
}

func TestComputeWeeklyBacklogEmpty(t *testing.T) {
	trend := ComputeWeeklyBacklog(nil, nil)
	if trend != nil {
		t.Errorf("expected nil for empty input")
	}
}

func TestComputeWeeklyBacklogNewestFirst(t *testing.T) {
	w1 := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	w2 := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	allTopics := []model.Topic{
		{ID: 1, CreatedAt: w1},
		{ID: 2, CreatedAt: w2},
	}
	trend := ComputeWeeklyBacklog(allTopics, nil)
	if len(trend) != 2 {
		t.Fatalf("got %d, want 2", len(trend))
	}
	if trend[0].WeekStart < trend[1].WeekStart {
		t.Errorf("should be newest first")
	}
}
