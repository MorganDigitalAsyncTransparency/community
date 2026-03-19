// Contract tests verify API response shapes match specs/api/api-contract.md.
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/domain"
	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/model"
)

func testServer() (ts *httptest.Server, srv *Server) {
	cfg := model.TagConfig{
		Defaults: model.TagDefaults{
			StalledDays: 7,
			Area:        "Other",
			SLO:         model.SLOThresholds{FirstReplyHours: 24, ResolutionHours: 336, InactivityHours: 48},
		},
		Areas: []model.Area{
			{Name: "Integration", PrimaryTag: "api"},
		},
		Tags: map[string]model.TagSpec{
			"api":            {Area: "Integration", ClosedTag: "closed", StalledDays: intP(7), SLO: &model.SLOThresholds{FirstReplyHours: 4, ResolutionHours: 48, InactivityHours: 24}},
			"webhooks":       {Area: "Integration"},
			"authentication": {Area: "Access"},
			"sso":            {Area: "Access"},
			"editor":         {Area: "Content"},
			"search":         {Area: "Content"},
			"ssl":            {Area: "Access"},
			"installation":   {Area: "Infrastructure"},
			"data-import":    {Area: "Infrastructure"},
			"migration":      {},
			"plugin":         {SLO: &model.SLOThresholds{FirstReplyHours: 8, ResolutionHours: 72, InactivityHours: 48}},
		},
	}
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	synced := now.Add(-2 * time.Hour)

	srv = &Server{
		Topics:         mock.Topics(),
		TagConfig:      cfg,
		ResolvedTags:   domain.ResolveAllTags(&cfg),
		BucketCeilings: []int{1, 4, 12, 24, 48, 96, 168},
		Version:        "0.1.0",
		LastSyncedAt:   &synced,
		Now:            func() time.Time { return now },
	}

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	return httptest.NewServer(mux), srv
}

func intP(v int) *int { return &v }

func get(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	_ = resp.Body.Close()
	return data
}

func decodeArray(t *testing.T, resp *http.Response) []any {
	t.Helper()
	var data []any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("decode array: %v", err)
	}
	_ = resp.Body.Close()
	return data
}

// AC-6: Content-Type
func TestContentType(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	resp := get(t, ts, "/api/v1/status")
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type: got %q", ct)
	}
}

// AC-7: Error response structure
func TestErrorResponse(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	resp := get(t, ts, "/api/v1/queue/summary?period=invalid")
	if resp.StatusCode != 400 {
		t.Errorf("status: got %d, want 400", resp.StatusCode)
	}
	data := decodeJSON(t, resp)
	if _, ok := data["error"]; !ok {
		t.Error("error response missing 'error' field")
	}
}

// AC-12: Queue summary shape
func TestQueueSummaryShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/queue/summary"))

	for _, field := range []string{"unrepliedCount", "untaggedCount", "oldestUnrepliedAgeDays"} {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-13: Unreplied topics shape and sort
func TestUnrepliedShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/queue/unreplied"))
	if len(items) == 0 {
		t.Fatal("expected unreplied topics")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"id", "title", "createdAt", "tags", "topicUrl"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-14: Untagged topics shape
func TestUntaggedShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/queue/untagged"))
	if len(items) == 0 {
		t.Fatal("expected untagged topics")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"id", "title", "createdAt", "categoryName", "topicUrl"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-15: Stalled topics shape
func TestStalledShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/queue/stalled"))
	if len(items) == 0 {
		t.Fatal("expected stalled topics")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"id", "title", "createdAt", "tags", "topicUrl", "strictestTag", "thresholdDays", "thresholdIsDefault", "daysSinceLastActivity"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-16: Metrics summary shape
func TestMetricsSummaryShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/metrics/summary"))
	for _, field := range []string{"medianFirstReplyMs", "medianResolutionMs", "solvedCount", "selfClosedCount", "answerRatePercent"} {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-17: Volume trend shape
func TestVolumeShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/metrics/volume"))
	if len(items) == 0 {
		t.Fatal("expected volume buckets")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"label", "bucketKey", "created", "accepted", "closed", "open"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-18: Median trends shape
func TestMedianTrendsShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/metrics/median-trends"))
	for _, key := range []string{"firstReply", "resolution"} {
		arr, ok := data[key].([]any)
		if !ok {
			t.Errorf("missing or wrong type: %s", key)
			continue
		}
		if len(arr) > 0 {
			bucket := arr[0].(map[string]any)
			for _, field := range []string{"label", "bucketKey"} {
				if _, ok := bucket[field]; !ok {
					t.Errorf("%s bucket missing field: %s", key, field)
				}
			}
			// medianMs can be null, just check key exists
			if _, ok := bucket["medianMs"]; !ok {
				t.Errorf("%s bucket missing field: medianMs", key)
			}
		}
	}
}

// AC-19: Distribution shape
func TestDistributionShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/metrics/distribution"))
	for _, key := range []string{"firstReply", "resolution"} {
		arr, ok := data[key].([]any)
		if !ok || len(arr) == 0 {
			t.Errorf("missing or empty: %s", key)
			continue
		}
		bucket := arr[0].(map[string]any)
		for _, field := range []string{"label", "count"} {
			if _, ok := bucket[field]; !ok {
				t.Errorf("%s bucket missing: %s", key, field)
			}
		}
	}
}

// AC-20: Tag volume shape
func TestTagVolumeShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/distribution/volume"))
	if len(items) == 0 {
		t.Fatal("expected tag volume data")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"tag", "topicCount"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-21: Tag resolution shape
func TestTagResolutionShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/distribution/resolution"))
	if len(items) == 0 {
		t.Fatal("expected tag resolution data")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"tag", "resolvedCount"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
	// medianResolutionMs can be null
	if _, ok := first["medianResolutionMs"]; !ok {
		t.Error("missing field: medianResolutionMs")
	}
}

// AC-22: Tag backlog shape
func TestTagBacklogShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/distribution/backlog"))
	if len(items) == 0 {
		t.Fatal("expected tag backlog data")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"tag", "openCount"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-23: Backlog trend shape
func TestBacklogTrendShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/distribution/backlog-trend"))
	if len(items) == 0 {
		t.Fatal("expected backlog trend data")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"weekStart", "created", "resolved", "stillOpen"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-24: SLO violations shape
func TestSLOViolationsShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/slo/violations"))
	for _, key := range []string{"firstReply", "resolution", "inactivity"} {
		arr, ok := data[key].([]any)
		if !ok {
			t.Errorf("missing or wrong type: %s", key)
			continue
		}
		if len(arr) > 0 {
			v := arr[0].(map[string]any)
			for _, field := range []string{"topicId", "topicTitle", "topicUrl", "tag", "thresholdMs", "actualMs", "excessMs", "thresholdIsDefault"} {
				if _, ok := v[field]; !ok {
					t.Errorf("%s violation missing: %s", key, field)
				}
			}
		}
	}
}

// AC-25: SLO compliance shape
func TestSLOComplianceShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	items := decodeArray(t, get(t, ts, "/api/v1/slo/compliance"))
	if len(items) == 0 {
		t.Fatal("expected compliance data")
	}
	first := items[0].(map[string]any)
	for _, field := range []string{"tag", "thresholdIsDefault"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
	// Percent fields can be null
	for _, field := range []string{"firstReplyPercent", "resolutionPercent", "inactivityPercent"} {
		if _, ok := first[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-26: Heatmap shape
func TestHeatmapShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/activity/heatmap"))
	if _, ok := data["maxCount"]; !ok {
		t.Error("missing maxCount")
	}
	cells, ok := data["cells"].([]any)
	if !ok || len(cells) != 7 {
		t.Fatalf("cells: got %d rows, want 7", len(cells))
	}
	row := cells[0].([]any)
	if len(row) != 24 {
		t.Fatalf("row: got %d cols, want 24", len(row))
	}
	cell := row[0].(map[string]any)
	for _, field := range []string{"day", "hour", "count"} {
		if _, ok := cell[field]; !ok {
			t.Errorf("cell missing: %s", field)
		}
	}
}

// AC-27: Config shape
func TestConfigShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/config"))
	for _, field := range []string{"areas", "tags", "defaults", "distributionBucketCeilings"} {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}

	// Verify tag structure
	tags := data["tags"].(map[string]any)
	if len(tags) == 0 {
		t.Fatal("expected tags")
	}
	for name, val := range tags {
		tag := val.(map[string]any)
		for _, field := range []string{"area", "areaIsDefault", "stalledDays", "stalledDaysIsDefault", "slo", "sloIsDefault", "closedTag"} {
			if _, ok := tag[field]; !ok {
				t.Errorf("tag %s missing: %s", name, field)
			}
		}
		break
	}
}

// AC-28: Status shape
func TestStatusShape(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	data := decodeJSON(t, get(t, ts, "/api/v1/status"))
	for _, field := range []string{"lastSyncedAt", "version"} {
		if _, ok := data[field]; !ok {
			t.Errorf("missing field: %s", field)
		}
	}
}

// AC-5: Empty dataset returns normal structure
func TestEmptyDatasetNotFoundNotReturned(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	// Filter to a tag with no unreplied topics in a very short window
	resp := get(t, ts, "/api/v1/queue/summary?tag=search&period=7d")
	if resp.StatusCode == 404 {
		t.Error("empty dataset should not return 404")
	}
	if resp.StatusCode != 200 {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}
}

// AC-11: Invalid filter values
func TestInvalidFilters(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()

	tests := []struct {
		query string
		desc  string
	}{
		{"period=3d", "unknown period"},
		{"from=bad&to=2026-01-01", "malformed from date"},
		{"from=2026-01-01", "from without to"},
		{"tag=nonexistent", "unknown tag"},
		{"from=2026-03-01&to=2026-01-01", "from after to"},
	}
	for _, tt := range tests {
		resp := get(t, ts, "/api/v1/queue/summary?"+tt.query)
		if resp.StatusCode != 400 {
			t.Errorf("%s: got %d, want 400", tt.desc, resp.StatusCode)
		}
		_ = resp.Body.Close()
	}
}

// AC-2: Only GET allowed
func TestOnlyGetAllowed(t *testing.T) {
	ts, _ := testServer()
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/api/v1/status", "application/json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	// Go 1.22+ returns 405 for method mismatch
	if resp.StatusCode == 200 {
		t.Error("POST should not return 200")
	}
}
