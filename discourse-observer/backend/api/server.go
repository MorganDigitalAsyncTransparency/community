// Spec: specs/api/api-contract.md (AC-1)
// Tests: backend/api/contract_test.go
package api

import (
	"net/http"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// SyncStateProvider exposes sync operational state for the status endpoint.
// Implemented by the scheduler's SyncStatus; nil when sync is disabled.
type SyncStateProvider interface {
	GetState() string
	GetLastDuration() time.Duration
	GetLastTopics() int
	GetLastSyncedAt() *time.Time
	GetLog() []model.SyncLogEntry
	GetProgress() *model.SyncProgress
}

// Server holds shared state for all API handlers.
type Server struct {
	Store          TopicReader
	Events         EventReader
	TagConfig      model.TagConfig
	ResolvedTags   map[string]model.ResolvedTag
	BucketCeilings []int
	Version        string
	SyncStatus     SyncStateProvider
	Now            func() time.Time
}

// MonitoredTags returns a set of tag names from the tag config.
func (s *Server) MonitoredTags() map[string]bool {
	m := make(map[string]bool, len(s.TagConfig.Tags))
	for tag := range s.TagConfig.Tags {
		m[tag] = true
	}
	return m
}

// RegisterRoutes adds all /api/v1/ routes to the mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/queue/summary", s.handleQueueSummary)
	mux.HandleFunc("GET /api/v1/queue/unreplied", s.handleQueueUnreplied)
	mux.HandleFunc("GET /api/v1/queue/untagged", s.handleQueueUntagged)
	mux.HandleFunc("GET /api/v1/queue/stalled", s.handleQueueStalled)

	mux.HandleFunc("GET /api/v1/metrics/summary", s.handleMetricsSummary)
	mux.HandleFunc("GET /api/v1/metrics/volume", s.handleMetricsVolume)
	mux.HandleFunc("GET /api/v1/metrics/median-trends", s.handleMetricsMedianTrends)
	mux.HandleFunc("GET /api/v1/metrics/distribution", s.handleMetricsDistribution)
	mux.HandleFunc("GET /api/v1/metrics/triage-time", s.handleTriageTime)

	mux.HandleFunc("GET /api/v1/distribution/volume", s.handleDistributionVolume)
	mux.HandleFunc("GET /api/v1/distribution/resolution", s.handleDistributionResolution)
	mux.HandleFunc("GET /api/v1/distribution/backlog", s.handleDistributionBacklog)
	mux.HandleFunc("GET /api/v1/distribution/backlog-trend", s.handleDistributionBacklogTrend)

	mux.HandleFunc("GET /api/v1/slo/violations", s.handleSLOViolations)
	mux.HandleFunc("GET /api/v1/slo/compliance", s.handleSLOCompliance)

	mux.HandleFunc("GET /api/v1/activity/heatmap", s.handleActivityHeatmap)

	mux.HandleFunc("GET /api/v1/config", s.handleConfig)
	mux.HandleFunc("GET /api/v1/status", s.handleStatus)
	mux.HandleFunc("GET /api/v1/sync-log", s.handleSyncLog)
}
