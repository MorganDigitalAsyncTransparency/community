package api

import (
	"net/http"
	"testing"

	"github.com/code-community/discourse-observer/backend/model"
)

var testConfig = model.TagConfig{
	Tags: map[string]model.TagSpec{
		"api":    {Area: "Integration"},
		"editor": {Area: "Content"},
	},
}

func makeRequest(query string) *http.Request {
	r, _ := http.NewRequest("GET", "/test?"+query, http.NoBody)
	return r
}

func TestParseFiltersDefaults(t *testing.T) {
	f, err := parseFilters(makeRequest(""), &testConfig)
	if err != nil {
		t.Fatal(err)
	}
	if f.Period != "all" {
		t.Errorf("default period: got %s, want all", f.Period)
	}
	if f.Tag != "" {
		t.Errorf("default tag: got %q, want empty", f.Tag)
	}
	if f.From != nil || f.To != nil {
		t.Errorf("default from/to: should be nil")
	}
}

func TestParseFiltersValidPeriod(t *testing.T) {
	for _, p := range []string{"7d", "30d", "1y", "all"} {
		f, err := parseFilters(makeRequest("period="+p), &testConfig)
		if err != nil {
			t.Errorf("period %s: unexpected error: %v", p, err)
		}
		if f.Period != p {
			t.Errorf("got %s, want %s", f.Period, p)
		}
	}
}

func TestParseFiltersInvalidPeriod(t *testing.T) {
	_, err := parseFilters(makeRequest("period=3d"), &testConfig)
	if err == nil {
		t.Error("expected error for invalid period")
	}
}

func TestParseFiltersValidDateRange(t *testing.T) {
	f, err := parseFilters(makeRequest("from=2026-01-01&to=2026-03-01"), &testConfig)
	if err != nil {
		t.Fatal(err)
	}
	if f.From == nil || f.To == nil {
		t.Fatal("from/to should not be nil")
	}
}

func TestParseFiltersFromOnly(t *testing.T) {
	_, err := parseFilters(makeRequest("from=2026-01-01"), &testConfig)
	if err == nil {
		t.Error("expected error for from without to")
	}
}

func TestParseFiltersFromAfterTo(t *testing.T) {
	_, err := parseFilters(makeRequest("from=2026-03-01&to=2026-01-01"), &testConfig)
	if err == nil {
		t.Error("expected error for from after to")
	}
}

func TestParseFiltersMalformedDate(t *testing.T) {
	_, err := parseFilters(makeRequest("from=not-a-date&to=2026-01-01"), &testConfig)
	if err == nil {
		t.Error("expected error for malformed date")
	}
}

func TestParseFiltersValidTag(t *testing.T) {
	f, err := parseFilters(makeRequest("tag=api"), &testConfig)
	if err != nil {
		t.Fatal(err)
	}
	if f.Tag != "api" {
		t.Errorf("got %s, want api", f.Tag)
	}
}

func TestParseFiltersUnknownTag(t *testing.T) {
	_, err := parseFilters(makeRequest("tag=nonexistent"), &testConfig)
	if err == nil {
		t.Error("expected error for unknown tag")
	}
}
