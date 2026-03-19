package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func makeMetricsTopics() []model.Topic {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	return []model.Topic{
		{ID: 1, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(2 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(48 * time.Hour))},
		{ID: 2, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(96 * time.Hour))},
		{ID: 3, Outcome: "self-closed", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(48 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(72 * time.Hour))},
	}
}

func TestComputeVolumeBuckets(t *testing.T) {
	base := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{CreatedAt: base, Outcome: "solved"},
		{CreatedAt: base, Outcome: "self-closed"},
		{CreatedAt: base, Outcome: ""},
	}
	buckets := ComputeVolumeBuckets(topics, "daily", "2026-03-12", "2026-03-12")
	if len(buckets) != 1 {
		t.Fatalf("got %d buckets, want 1", len(buckets))
	}
	b := buckets[0]
	if b.Created != 3 || b.Accepted != 1 || b.Closed != 1 || b.Open != 1 {
		t.Errorf("counts: created=%d accepted=%d closed=%d open=%d",
			b.Created, b.Accepted, b.Closed, b.Open)
	}
}

func TestComputeVolumeBucketsGapFill(t *testing.T) {
	topics := []model.Topic{
		{CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)},
		{CreatedAt: time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)},
	}
	buckets := ComputeVolumeBuckets(topics, "daily", "2026-03-10", "2026-03-13")
	if len(buckets) != 4 {
		t.Fatalf("got %d buckets, want 4 (with gap fill)", len(buckets))
	}
	if buckets[1].Created != 0 || buckets[2].Created != 0 {
		t.Errorf("gap buckets should have 0 created")
	}
}

func TestComputeVolumeBucketsEmpty(t *testing.T) {
	buckets := ComputeVolumeBuckets(nil, "daily", "", "")
	if buckets != nil {
		t.Errorf("nil range should return nil")
	}
}

func TestComputeMedianTrend(t *testing.T) {
	base := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour))},
		{CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(10 * time.Hour))},
	}
	buckets := ComputeMedianTrend(topics, "daily", "2026-03-10", "2026-03-10", FirstReplyExtractor)
	if len(buckets) != 1 {
		t.Fatalf("got %d, want 1", len(buckets))
	}
	// Median of [6h, 10h] = 8h = 28800000ms
	if buckets[0].MedianMs == nil || *buckets[0].MedianMs != 8*3_600_000 {
		t.Errorf("median: got %v, want %d", buckets[0].MedianMs, 8*3_600_000)
	}
}

func TestComputeMedianTrendNoData(t *testing.T) {
	topics := []model.Topic{
		{CreatedAt: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)},
	}
	buckets := ComputeMedianTrend(topics, "daily", "2026-03-10", "2026-03-10", FirstReplyExtractor)
	if len(buckets) != 1 {
		t.Fatalf("got %d, want 1", len(buckets))
	}
	if buckets[0].MedianMs != nil {
		t.Errorf("topic without firstReplyAt: median should be nil")
	}
}

func TestMondayOf(t *testing.T) {
	// Wednesday 2026-03-11 → Monday 2026-03-09
	wed := time.Date(2026, 3, 11, 15, 0, 0, 0, time.UTC)
	mon := MondayOf(wed)
	if DayString(mon) != "2026-03-09" {
		t.Errorf("MondayOf Wednesday: got %s", DayString(mon))
	}

	// Sunday 2026-03-15 → Monday 2026-03-09
	sun := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	mon = MondayOf(sun)
	if DayString(mon) != "2026-03-09" {
		t.Errorf("MondayOf Sunday: got %s", DayString(mon))
	}

	// Monday 2026-03-09 → Monday 2026-03-09
	monDay := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	mon = MondayOf(monDay)
	if DayString(mon) != "2026-03-09" {
		t.Errorf("MondayOf Monday: got %s", DayString(mon))
	}
}

func TestGranularity(t *testing.T) {
	tests := []struct {
		period string
		from   *time.Time
		to     *time.Time
		want   string
	}{
		{"7d", nil, nil, "daily"},
		{"30d", nil, nil, "daily"},
		{"1y", nil, nil, "weekly"},
		{"all", nil, nil, "weekly"},
	}
	for _, tt := range tests {
		got := Granularity(tt.period, tt.from, tt.to)
		if got != tt.want {
			t.Errorf("period %s: got %s, want %s", tt.period, got, tt.want)
		}
	}

	// Custom range: 89 days → daily
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	got := Granularity("", &from, &to)
	if got != "daily" {
		t.Errorf("89 day range: got %s, want daily", got)
	}

	// Custom range: 91 days → weekly
	to = time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC)
	got = Granularity("", &from, &to)
	if got != "weekly" {
		t.Errorf("91 day range: got %s, want weekly", got)
	}
}

func TestWeeklyBucketKey(t *testing.T) {
	// Wednesday 2026-03-11 → bucket key "2026-03-09" (Monday)
	wed := time.Date(2026, 3, 11, 15, 0, 0, 0, time.UTC)
	key := BucketKey(wed, "weekly")
	if key != "2026-03-09" {
		t.Errorf("weekly bucket key: got %s, want 2026-03-09", key)
	}

	// Same day in daily mode
	key = BucketKey(wed, "daily")
	if key != "2026-03-11" {
		t.Errorf("daily bucket key: got %s, want 2026-03-11", key)
	}
}
