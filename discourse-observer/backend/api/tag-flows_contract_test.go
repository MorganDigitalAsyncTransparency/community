// Spec: specs/api/tag-flows.md (TF-14, TF-15, TF-16)
package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestTagFlowsEndpoint(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/metrics/tag-flows")
	if err != nil {
		t.Fatalf("GET tag-flows: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Transitions []struct {
			From                []string `json:"from"`
			To                  []string `json:"to"`
			Count               int      `json:"count"`
			MedianDurationHours *float64 `json:"medianDurationHours"`
		} `json:"transitions"`
		TagPairs []struct {
			Tags  [2]string `json:"tags"`
			Count int       `json:"count"`
		} `json:"tagPairs"`
		Summary struct {
			TopicsWithTagChanges  int      `json:"topicsWithTagChanges"`
			TotalTopics           int      `json:"totalTopics"`
			MedianChangesPerTopic *float64 `json:"medianChangesPerTopic"`
			MostCommonFirstTag    *string  `json:"mostCommonFirstTag"`
			MostUnstableTag       *string  `json:"mostUnstableTag"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// With no events seeded, should have zero changes but topics present
	if body.Summary.TotalTopics == 0 {
		t.Error("totalTopics should be > 0 with seeded mock data")
	}
	if body.Summary.TopicsWithTagChanges != 0 {
		t.Errorf("topicsWithTagChanges = %d, want 0 (no events seeded)", body.Summary.TopicsWithTagChanges)
	}
	if body.Transitions == nil {
		t.Error("transitions should be empty array, not null")
	}
	if body.TagPairs == nil {
		t.Error("tagPairs should be empty array, not null")
	}
}

func TestTagFlowsEndpointFilters(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/metrics/tag-flows?period=7d")
	if err != nil {
		t.Fatalf("GET with period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("period filter: status = %d, want 200", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/api/v1/metrics/tag-flows?period=invalid")
	if err != nil {
		t.Fatalf("GET with invalid period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("invalid filter: status = %d, want 400", resp.StatusCode)
	}
}
