// Spec: specs/observer/scheduler.md
// Tests: backend/scheduler/scheduler_acceptance_test.go
package scheduler

import (
	"context"
	"log"
	"math/rand/v2"
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

// GetState returns the current sync state.
func (s *SyncStatus) GetState() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.State
}

// GetLastDuration returns the duration of the last completed sync.
func (s *SyncStatus) GetLastDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastDuration
}

// GetLastTopics returns the topic count from the last completed sync.
func (s *SyncStatus) GetLastTopics() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastTopics
}

// GetLastSyncedAt returns the timestamp of the last completed sync.
func (s *SyncStatus) GetLastSyncedAt() *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastSyncedAt
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
	runner     SyncRunner
	cfg        config.SyncConfig
	status     *SyncStatus
	logger     *log.Logger
	running    sync.Mutex
	zeroStreak int
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
// It runs one immediate sync (auto-detect), then loops delta syncs on interval+jitter.
// When ctx is canceled, it waits for any in-progress sync before returning.
func (s *Scheduler) Start(ctx context.Context) {
	s.runSync(ctx, func(ctx context.Context) (observer.SyncResult, error) {
		return s.runner.Run(ctx)
	})

	for {
		wait := s.cfg.Interval + jitter(s.cfg.JitterMax)
		if !sleepCtx(ctx, wait) {
			return
		}
		s.runSync(ctx, func(ctx context.Context) (observer.SyncResult, error) {
			return s.runner.RunDeltaSync(ctx)
		})
	}
}

// runSync executes a single sync cycle with concurrency guard and status updates.
// The sync runs with a non-cancelable context so it completes even during shutdown.
func (s *Scheduler) runSync(ctx context.Context, fn func(context.Context) (observer.SyncResult, error)) {
	if !s.running.TryLock() {
		s.logger.Println("sync skipped: previous sync still running")
		return
	}
	defer s.running.Unlock()

	s.setState("running")
	defer s.setState("idle")

	s.logger.Printf("sync started")
	start := time.Now()

	// Use a detached context so in-progress syncs finish during shutdown.
	syncCtx := context.WithoutCancel(ctx)
	result, err := fn(syncCtx)
	duration := time.Since(start)

	if err != nil {
		s.logger.Printf("sync aborted: %v (duration=%s)", err, duration)
		return
	}

	s.logCompleted(result, duration)
	s.recordResult(result, duration)
	s.trackLowActivity(result)
}

func (s *Scheduler) logCompleted(r observer.SyncResult, d time.Duration) {
	s.logger.Printf("sync completed: type=%s pages=%d topics=%d duration=%s",
		r.Mode, r.PagesFetched, r.TopicsStored, d)
}

func (s *Scheduler) recordResult(r observer.SyncResult, d time.Duration) {
	now := time.Now().UTC()
	s.status.mu.Lock()
	defer s.status.mu.Unlock()
	s.status.LastDuration = d
	s.status.LastTopics = r.TopicsStored
	s.status.LastSyncedAt = &now
}

func (s *Scheduler) trackLowActivity(r observer.SyncResult) {
	if r.TopicsStored > 0 {
		s.zeroStreak = 0
		return
	}
	s.zeroStreak++
	if s.zeroStreak >= s.cfg.LowActivityThreshold {
		s.logger.Printf("low activity detected: %d consecutive zero-change syncs", s.zeroStreak)
	}
}

func (s *Scheduler) setState(state string) {
	s.status.mu.Lock()
	defer s.status.mu.Unlock()
	s.status.State = state
}

// jitter returns a random duration in [0, maxJitter).
func jitter(maxJitter time.Duration) time.Duration {
	if maxJitter <= 0 {
		return 0
	}
	return time.Duration(rand.Int64N(int64(maxJitter)))
}

// sleepCtx waits for d or until ctx is canceled. Returns false if ctx was canceled.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}
