package domain

import (
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

var sloConfig = map[string]model.ResolvedTag{
	"api": {
		SLO:          model.SLOThresholds{FirstReplyHours: 4, ResolutionHours: 48, InactivityHours: 24},
		SLOIsDefault: false,
	},
	"plugin": {
		SLO:          model.SLOThresholds{FirstReplyHours: 8, ResolutionHours: 72, InactivityHours: 48},
		SLOIsDefault: false,
	},
}

func TestFindViolationsFirstReply(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := now.Add(-24 * time.Hour)
	topics := []model.Topic{
		// 6h first reply vs 4h threshold → violation
		{ID: 1, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(24 * time.Hour))},
		// 2h first reply vs 4h threshold → OK
		{ID: 2, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(2 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(12 * time.Hour))},
	}

	groups := FindViolations(topics, sloConfig, now)
	if len(groups.FirstReply) != 1 {
		t.Fatalf("first reply violations: got %d, want 1", len(groups.FirstReply))
	}
	v := groups.FirstReply[0]
	if v.TopicID != 1 {
		t.Errorf("violation topic: got %d, want 1", v.TopicID)
	}
	if v.ExcessMs <= 0 {
		t.Errorf("excess should be positive")
	}
}

func TestFindViolationsUnreplied(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		// Unreplied for ~48h vs 4h threshold → first-reply + inactivity violations
		{ID: 1, Tags: []string{"api"}, ReplyCount: 0,
			CreatedAt: now.Add(-48 * time.Hour)},
	}

	groups := FindViolations(topics, sloConfig, now)
	if len(groups.FirstReply) != 1 {
		t.Errorf("unreplied first reply: got %d, want 1", len(groups.FirstReply))
	}
	if len(groups.Inactivity) != 1 {
		t.Errorf("unreplied inactivity: got %d, want 1", len(groups.Inactivity))
	}
}

func TestFindViolationsResolution(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := now.Add(-5 * 24 * time.Hour)
	topics := []model.Topic{
		// 96h resolution vs 72h plugin threshold → violation
		{ID: 1, Tags: []string{"plugin"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(4 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(96 * time.Hour))},
	}

	groups := FindViolations(topics, sloConfig, now)
	if len(groups.Resolution) != 1 {
		t.Fatalf("resolution violations: got %d, want 1", len(groups.Resolution))
	}
	if groups.Resolution[0].ExcessMs != 24*3_600_000 {
		t.Errorf("excess: got %d, want %d", groups.Resolution[0].ExcessMs, 24*3_600_000)
	}
}

func TestFindViolationsWithinThreshold(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := now.Add(-24 * time.Hour)
	topics := []model.Topic{
		{ID: 1, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(2 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(24 * time.Hour))},
	}
	groups := FindViolations(topics, sloConfig, now)
	if len(groups.FirstReply) != 0 || len(groups.Resolution) != 0 || len(groups.Inactivity) != 0 {
		t.Errorf("within threshold should produce no violations")
	}
}

func TestFindViolationsSortedByExcess(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := now.Add(-5 * 24 * time.Hour)
	topics := []model.Topic{
		{ID: 1, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(24 * time.Hour))},
		{ID: 2, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(10 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(24 * time.Hour))},
	}
	groups := FindViolations(topics, sloConfig, now)
	if len(groups.FirstReply) < 2 {
		t.Fatalf("expected 2 violations")
	}
	if groups.FirstReply[0].ExcessMs < groups.FirstReply[1].ExcessMs {
		t.Errorf("violations should be sorted by excess descending")
	}
}

func TestFindViolationsUnknownTag(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	topics := []model.Topic{
		{ID: 1, Tags: []string{"unknown"}, ReplyCount: 0,
			CreatedAt: now.Add(-48 * time.Hour)},
	}
	groups := FindViolations(topics, sloConfig, now)
	if len(groups.FirstReply) != 0 {
		t.Errorf("unknown tags should produce no violations")
	}
}

func TestComputeCompliance(t *testing.T) {
	now := time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
	base := now.Add(-48 * time.Hour)
	topics := []model.Topic{
		// Compliant: 2h first reply, 24h resolution
		{ID: 1, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(2 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(24 * time.Hour))},
		// Violating: 6h first reply (>4h), 72h resolution (>48h)
		{ID: 2, Tags: []string{"api"}, Outcome: "solved", CreatedAt: base,
			FirstReplyAt: timePtr(base.Add(6 * time.Hour)),
			ResolvedAt:   timePtr(base.Add(72 * time.Hour))},
	}

	compliance := ComputeCompliance(topics, sloConfig, now)

	var apiCompliance *TagCompliance
	for i := range compliance {
		if compliance[i].Tag == "api" {
			apiCompliance = &compliance[i]
			break
		}
	}
	if apiCompliance == nil {
		t.Fatal("api compliance not found")
	}
	if apiCompliance.FirstReplyPercent == nil || *apiCompliance.FirstReplyPercent != 50 {
		t.Errorf("first reply: got %v, want 50", apiCompliance.FirstReplyPercent)
	}
	if apiCompliance.ResolutionPercent == nil || *apiCompliance.ResolutionPercent != 50 {
		t.Errorf("resolution: got %v, want 50", apiCompliance.ResolutionPercent)
	}
}

func TestComputeComplianceNullForNoData(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	compliance := ComputeCompliance(nil, sloConfig, now)

	for _, c := range compliance {
		if c.FirstReplyPercent != nil || c.ResolutionPercent != nil || c.InactivityPercent != nil {
			t.Errorf("tag %s: expected nil percentages for no data", c.Tag)
		}
	}
}

func TestComputeComplianceSortedByTag(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	compliance := ComputeCompliance(nil, sloConfig, now)
	for i := 1; i < len(compliance); i++ {
		if compliance[i].Tag < compliance[i-1].Tag {
			t.Errorf("not sorted: %s before %s", compliance[i-1].Tag, compliance[i].Tag)
		}
	}
}
