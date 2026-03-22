// Spec: specs/api/triage-time.md (TT-8, TT-9, TT-10)
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// fakeEventReader implements EventReader for tests.
type fakeEventReader struct {
	events map[int][]model.TopicEvent
}

func (f *fakeEventReader) LoadAllTopicEvents(_ context.Context) (map[int][]model.TopicEvent, error) {
	if f.events == nil {
		return map[int][]model.TopicEvent{}, nil
	}
	return f.events, nil
}

func TestTriageTimeEndpoint(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/metrics/triage-time")
	if err != nil {
		t.Fatalf("GET triage-time: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Overall struct {
			MedianHours *float64 `json:"medianHours"`
			Count       int      `json:"count"`
		} `json:"overall"`
		ByTag []struct {
			Tag         string   `json:"tag"`
			MedianHours *float64 `json:"medianHours"`
			Count       int      `json:"count"`
		} `json:"byTag"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// With no events seeded, count should be 0 and medianHours null
	if body.Overall.Count != 0 {
		t.Errorf("overall.count = %d, want 0", body.Overall.Count)
	}
	if body.Overall.MedianHours != nil {
		t.Error("overall.medianHours should be null with no events")
	}
	if body.ByTag == nil {
		t.Error("byTag should be empty array, not null")
	}
}

func TestTriageTimeEndpointWithEvents(t *testing.T) {
	ts, srv := testServer(t)
	defer ts.Close()

	now := srv.Now()
	tagDetail := `{"previous":[],"current":["api"]}`
	srv.Events = &fakeEventReader{
		events: map[int][]model.TopicEvent{
			1041: {{TopicID: 1041, EventType: "tag_change", HappenedAt: now.Add(-48 * time.Hour), Detail: tagDetail}},
		},
	}

	resp, err := http.Get(ts.URL + "/api/v1/metrics/triage-time")
	if err != nil {
		t.Fatalf("GET triage-time: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Overall struct {
			Count int `json:"count"`
		} `json:"overall"`
		ByTag []struct {
			Tag   string `json:"tag"`
			Count int    `json:"count"`
		} `json:"byTag"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Overall.Count < 1 {
		t.Errorf("expected at least 1 qualifying topic, got %d", body.Overall.Count)
	}
}

func TestTriageTimeEndpointFilters(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	// Period filter
	resp, err := http.Get(ts.URL + "/api/v1/metrics/triage-time?period=7d")
	if err != nil {
		t.Fatalf("GET with period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("period filter: status = %d, want 200", resp.StatusCode)
	}

	// Tag filter
	resp, err = http.Get(ts.URL + "/api/v1/metrics/triage-time?tag=api")
	if err != nil {
		t.Fatalf("GET with tag: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("tag filter: status = %d, want 200", resp.StatusCode)
	}

	// Invalid filter
	resp, err = http.Get(ts.URL + "/api/v1/metrics/triage-time?period=invalid")
	if err != nil {
		t.Fatalf("GET with invalid period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("invalid filter: status = %d, want 400", resp.StatusCode)
	}
}
