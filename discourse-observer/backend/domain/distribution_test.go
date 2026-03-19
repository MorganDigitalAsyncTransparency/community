package domain

import "testing"

func TestBucketDurations(t *testing.T) {
	ceilings := []int{1, 4, 12, 24, 48, 96, 168}
	durationsMs := []int64{
		30 * 60_000,     // 0.5h → "< 1h"
		2 * 3_600_000,   // 2h → "1h–4h"
		6 * 3_600_000,   // 6h → "4h–12h"
		20 * 3_600_000,  // 20h → "12h–1d"
		30 * 3_600_000,  // 30h → "1d–2d"
		72 * 3_600_000,  // 72h → "2d–4d"
		120 * 3_600_000, // 120h → "4d–7d"
		200 * 3_600_000, // 200h → "> 7d"
	}

	buckets := BucketDurations(durationsMs, ceilings)
	if len(buckets) != 8 {
		t.Fatalf("got %d buckets, want 8", len(buckets))
	}

	expected := []struct {
		label string
		count int
	}{
		{"< 1h", 1},
		{"1h\u20134h", 1},
		{"4h\u201312h", 1},
		{"12h\u20131d", 1},
		{"1d\u20132d", 1},
		{"2d\u20134d", 1},
		{"4d\u20137d", 1},
		{"> 7d", 1},
	}
	for i, e := range expected {
		if buckets[i].Label != e.label {
			t.Errorf("bucket %d label: got %q, want %q", i, buckets[i].Label, e.label)
		}
		if buckets[i].Count != e.count {
			t.Errorf("bucket %d count: got %d, want %d", i, buckets[i].Count, e.count)
		}
	}
}

func TestBucketDurationsBoundary(t *testing.T) {
	ceilings := []int{1, 4}
	// Exactly 1h = 3600000ms should go to "1h–4h", not "< 1h"
	buckets := BucketDurations([]int64{3_600_000}, ceilings)
	if buckets[0].Count != 0 {
		t.Errorf("exactly 1h should NOT be in '< 1h' bucket")
	}
	if buckets[1].Count != 1 {
		t.Errorf("exactly 1h should be in '1h–4h' bucket")
	}
}

func TestBucketDurationsEmpty(t *testing.T) {
	ceilings := []int{1, 4, 12, 24, 48, 96, 168}
	buckets := BucketDurations(nil, ceilings)
	if len(buckets) != 8 {
		t.Fatalf("got %d buckets, want 8 for empty input", len(buckets))
	}
	for _, b := range buckets {
		if b.Count != 0 {
			t.Errorf("bucket %q: got count %d, want 0", b.Label, b.Count)
		}
	}
}

func TestFirstReplyDurations(t *testing.T) {
	topics := makeMetricsTopics()
	durations := FirstReplyDurations(topics)
	if len(durations) != 3 {
		t.Errorf("got %d durations, want 3", len(durations))
	}
}

func TestResolutionDurations(t *testing.T) {
	topics := makeMetricsTopics()
	durations := ResolutionDurations(topics)
	if len(durations) != 3 {
		t.Errorf("got %d durations, want 3", len(durations))
	}
}
