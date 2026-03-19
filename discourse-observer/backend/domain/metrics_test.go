package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func TestComputeMetricsSummary(t *testing.T) {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{ID: 1, Outcome: "solved",
			CreatedAt:    base,
			FirstReplyAt: timePtr(base.Add(2 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(48 * time.Hour))},
		{ID: 2, Outcome: "solved",
			CreatedAt:    base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(96 * time.Hour))},
		{ID: 3, Outcome: "self-closed",
			CreatedAt:    base,
			FirstReplyAt: timePtr(base.Add(48 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(72 * time.Hour))},
	}

	result := ComputeMetricsSummary(topics)

	// Median first reply: [2h, 6h, 48h] → 6h = 21600000ms
	if result.MedianFirstReplyMs == nil || *result.MedianFirstReplyMs != 6*3_600_000 {
		t.Errorf("median first reply: got %v, want %d", result.MedianFirstReplyMs, 6*3_600_000)
	}
	// Median resolution: [48h, 72h, 96h] → 72h
	if result.MedianResolutionMs == nil || *result.MedianResolutionMs != 72*3_600_000 {
		t.Errorf("median resolution: got %v, want %d", result.MedianResolutionMs, 72*3_600_000)
	}
	if result.SolvedCount != 2 {
		t.Errorf("solved: got %d, want 2", result.SolvedCount)
	}
	if result.SelfClosedCount != 1 {
		t.Errorf("self-closed: got %d, want 1", result.SelfClosedCount)
	}
	// Answer rate: 2/3 = 67%
	if result.AnswerRatePercent == nil || *result.AnswerRatePercent != 67 {
		t.Errorf("answer rate: got %v, want 67", result.AnswerRatePercent)
	}
}

func TestComputeMetricsSummaryEmpty(t *testing.T) {
	result := ComputeMetricsSummary(nil)
	if result.MedianFirstReplyMs != nil {
		t.Errorf("expected nil median first reply for empty input")
	}
	if result.MedianResolutionMs != nil {
		t.Errorf("expected nil median resolution for empty input")
	}
	if result.AnswerRatePercent != nil {
		t.Errorf("expected nil answer rate for empty input")
	}
}

func TestAnswerRate100Percent(t *testing.T) {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{Outcome: "solved", CreatedAt: base, ResolvedAt: timePtr(base.Add(time.Hour))},
	}
	result := ComputeMetricsSummary(topics)
	if result.AnswerRatePercent == nil || *result.AnswerRatePercent != 100 {
		t.Errorf("answer rate: got %v, want 100", result.AnswerRatePercent)
	}
}

func TestAnswerRate0Percent(t *testing.T) {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{Outcome: "self-closed", CreatedAt: base, ResolvedAt: timePtr(base.Add(time.Hour))},
	}
	result := ComputeMetricsSummary(topics)
	if result.AnswerRatePercent == nil || *result.AnswerRatePercent != 0 {
		t.Errorf("answer rate: got %v, want 0", result.AnswerRatePercent)
	}
}
