// Spec: specs/observer/scheduler.md (SC-12)
// Tests: backend/scheduler/scheduler_acceptance_test.go
package model

import "time"

// SyncLogEntry records one completed sync cycle for the sync log.
type SyncLogEntry struct {
	Timestamp time.Time
	Mode      string
	Pages     int
	Topics    int
	Duration  time.Duration
}
