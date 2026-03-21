// Spec: specs/observer/scheduler.md (SC-12)
// Tests: backend/scheduler/scheduler_acceptance_test.go
package model

import "time"

// SyncLogEntry records one completed sync cycle for the sync log.
type SyncLogEntry struct {
	Timestamp  time.Time
	Mode       string
	Pages      int
	Topics     int
	Duration   time.Duration
	HasChanges bool
	Error      string // empty on success, error message on failure
}

// SyncProgress tracks a sync cycle in progress.
type SyncProgress struct {
	Mode        string
	Pages       int
	Topics      int
	TotalTopics int
	StartedAt   time.Time
}

// ProgressFunc is called after each page during a sync cycle.
// totalTopics is the estimated total (0 if unknown).
type ProgressFunc func(mode string, pages, topics, totalTopics int)
