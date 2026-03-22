// Spec: specs/api/escalations.md (EP-8, EP-9, EP-10)
package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestEscalationEndpoint(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/metrics/escalations")
	if err != nil {
		t.Fatalf("GET escalations: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Total    int      `json:"total"`
		Rate     *float64 `json:"rate"`
		ByPeriod []struct {
			Period string `json:"period"`
			Count  int    `json:"count"`
		} `json:"byPeriod"`
		CommonPatterns []struct {
			OriginalTags    []string `json:"originalTags"`
			AddedAfterReply []string `json:"addedAfterReply"`
			Count           int      `json:"count"`
		} `json:"commonPatterns"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if body.Total != 0 {
		t.Errorf("total = %d, want 0 (no events seeded)", body.Total)
	}
	// Rate should not be nil — there are replied topics in mock data
	if body.Rate == nil {
		t.Error("rate should not be nil with replied topics in mock data")
	}
	if body.ByPeriod == nil {
		t.Error("byPeriod should be empty array, not null")
	}
	if body.CommonPatterns == nil {
		t.Error("commonPatterns should be empty array, not null")
	}
}

func TestEscalationEndpointFilters(t *testing.T) {
	ts, _ := testServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/metrics/escalations?period=7d")
	if err != nil {
		t.Fatalf("GET with period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("period filter: status = %d, want 200", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/api/v1/metrics/escalations?period=invalid")
	if err != nil {
		t.Fatalf("GET with invalid period: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("invalid filter: status = %d, want 400", resp.StatusCode)
	}
}
