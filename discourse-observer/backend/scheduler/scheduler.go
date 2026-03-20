// Spec: specs/observer/scheduler.md
// Tests: backend/scheduler/scheduler_acceptance_test.go
package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/code-community/discourse-observer/backend/config"
	"github.com/code-community/discourse-observer/backend/observer"
)

// SyncRunner abstracts the observer's sync methods for testability.
type SyncRunner interface {
	Run(ctx context.Context) (observer.SyncResult, error)
	RunDeltaSync(ctx context.Context) (observer.SyncResult, error)
}

// SyncStatus holds thread-safe operational state for the API to read.
type SyncStatus struct {
	mu           sync.RWMutex
	State        string
	LastDuration time.Duration
	LastTopics   int
	LastSyncedAt *time.Time
}

// Snapshot returns a copy of the current status.
func (s *SyncStatus) Snapshot() StatusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return StatusSnapshot{
		State:        s.State,
		LastDuration: s.LastDuration,
		LastTopics:   s.LastTopics,
		LastSyncedAt: s.LastSyncedAt,
	}
}

// StatusSnapshot is a point-in-time copy of SyncStatus, safe to read without locks.
type StatusSnapshot struct {
	State        string
	LastDuration time.Duration
	LastTopics   int
	LastSyncedAt *time.Time
}

// Scheduler drives the sync lifecycle.
type Scheduler struct {
	runner SyncRunner
	cfg    config.SyncConfig
	status *SyncStatus
	logger *log.Logger
}

// New creates a Scheduler with default logging to stderr.
func New(runner SyncRunner, cfg config.SyncConfig) *Scheduler {
	return NewWithLogger(runner, cfg, nil)
}

// NewWithLogger creates a Scheduler with a custom logger.
// If logger is nil, a default logger writing to stderr is used.
func NewWithLogger(runner SyncRunner, cfg config.SyncConfig, logger *log.Logger) *Scheduler {
	if logger == nil {
		logger = log.Default()
	}
	return &Scheduler{
		runner: runner,
		cfg:    cfg,
		status: &SyncStatus{State: "idle"},
		logger: logger,
	}
}

// Status returns the shared status for the API to read.
func (s *Scheduler) Status() *SyncStatus {
	return s.status
}

// Start runs the sync lifecycle until ctx is canceled.
// Placeholder — implementation in Phase 4.
func (s *Scheduler) Start(ctx context.Context) {
	_ = ctx
}
