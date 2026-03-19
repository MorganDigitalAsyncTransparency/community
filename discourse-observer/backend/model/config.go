// Spec: specs/api/api-contract.md (AC-27)
// Tests: backend/api/contract_test.go
package model

// TagConfig holds tag configuration loaded from config/tagConfig.json.
type TagConfig struct {
	Defaults TagDefaults        `json:"defaults"`
	Areas    []Area             `json:"areas"`
	Tags     map[string]TagSpec `json:"tags"`
}

// TagDefaults holds global fallback values.
type TagDefaults struct {
	StalledDays int           `json:"stalledDays"`
	Area        string        `json:"area"`
	SLO         SLOThresholds `json:"slo"`
}

// Area represents a named group of related tags.
type Area struct {
	Name       string `json:"name"`
	PrimaryTag string `json:"primaryTag"`
}

// TagSpec holds per-tag overrides (all optional except area).
type TagSpec struct {
	Area        string         `json:"area,omitempty"`
	ClosedTag   string         `json:"closedTag,omitempty"`
	StalledDays *int           `json:"stalledDays,omitempty"`
	SLO         *SLOThresholds `json:"slo,omitempty"`
}

// SLOThresholds holds SLO threshold values in hours.
type SLOThresholds struct {
	FirstReplyHours int `json:"firstReplyHours"`
	ResolutionHours int `json:"resolutionHours"`
	InactivityHours int `json:"inactivityHours"`
}

// ResolvedTag holds the fully resolved configuration for a single tag.
type ResolvedTag struct {
	Area                 string        `json:"area"`
	AreaIsDefault        bool          `json:"areaIsDefault"`
	StalledDays          int           `json:"stalledDays"`
	StalledDaysIsDefault bool          `json:"stalledDaysIsDefault"`
	SLO                  SLOThresholds `json:"slo"`
	SLOIsDefault         bool          `json:"sloIsDefault"`
	ClosedTag            *string       `json:"closedTag"`
}

// DistributionBuckets holds bucket ceilings from config/distributionBuckets.json.
type DistributionBuckets struct {
	BucketCeilingsHours []int `json:"bucketCeilingsHours"`
}
