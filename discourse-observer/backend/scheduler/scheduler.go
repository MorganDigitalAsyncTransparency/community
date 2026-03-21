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
	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/observer"
)

// SyncRunner abstracts the observer's sync methods for testability.
type SyncRunner interface {
	Run(ctx context.Context) (observer.SyncResult, error)
	RunDeltaSync(ctx context.Context) (observer.SyncResult, error)
	RunDetailSync(ctx context.Context) (observer.SyncResult, error)
}

// ActivityDataProvider supplies historical activity counts by hour and day.
// The scheduler uses this to identify low-activity windows for detail sync.
// Implemented by storage (querying topic creation times).
type ActivityDataProvider interface {
	// ActivityByHour returns the number of topics created during each
	// day-of-week (0=Monday..6=Sunday) and hour (0..23) combination.
	ActivityByHour(ctx context.Context) ([7][24]int, error)
}

// SyncLogStore persists sync log entries. Implemented by SQLiteStore.
type SyncLogStore interface {
	SaveSyncLogEntry(ctx context.Context, e *model.SyncLogEntry) error
	LoadSyncLog(ctx context.Context) ([]model.SyncLogEntry, error)
}

// SyncStatus holds thread-safe operational state for the API to read.
type SyncStatus struct {
	mu           sync.RWMutex
	State        string
	LastDuration time.Duration
	LastTopics   int
	LastSyncedAt *time.Time
	log          []model.SyncLogEntry
	progress     *model.SyncProgress
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

// GetProgress returns the current in-progress sync, or nil if idle.
func (s *SyncStatus) GetProgress() *model.SyncProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.progress == nil {
		return nil
	}
	cp := *s.progress
	return &cp
}

// UpdateProgress records per-page progress during a sync cycle.
func (s *SyncStatus) UpdateProgress(mode string, pages, topics, totalTopics int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.progress == nil {
		return
	}
	s.progress.Mode = mode
	s.progress.Pages = pages
	s.progress.Topics = topics
	if totalTopics > 0 {
		s.progress.TotalTopics = totalTopics
	}
}

// GetLog returns the most recent sync log entries (newest first).
func (s *SyncStatus) GetLog() []model.SyncLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.SyncLogEntry, len(s.log))
	copy(out, s.log)
	return out
}

// Scheduler drives the sync lifecycle.
type Scheduler struct {
	runner       SyncRunner
	cfg          config.SyncConfig
	status       *SyncStatus
	logStore     SyncLogStore
	activityData ActivityDataProvider
	logger       *log.Logger
	running      sync.Mutex
	zeroStreak   int
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

// SetLogStore sets the persistent store for sync log entries.
// If set, entries are persisted to SQLite and loaded on startup.
func (s *Scheduler) SetLogStore(store SyncLogStore) {
	s.logStore = store
}

// SetActivityData sets the provider for historical activity data
// used to detect low-activity windows for detail sync.
func (s *Scheduler) SetActivityData(provider ActivityDataProvider) {
	s.activityData = provider
}

// Status returns the shared status for the API to read.
func (s *Scheduler) Status() *SyncStatus {
	return s.status
}

// Start runs the sync lifecycle until ctx is canceled.
// It runs one immediate sync (auto-detect), then loops delta syncs on interval+jitter.
// When ctx is canceled, it waits for any in-progress sync before returning.
func (s *Scheduler) Start(ctx context.Context) {
	if s.logStore != nil {
		if entries, err := s.logStore.LoadSyncLog(ctx); err == nil {
			s.status.mu.Lock()
			s.status.log = entries
			s.status.mu.Unlock()
		}
	}

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
		if s.shouldRunDetailSync(ctx) {
			s.logger.Printf("low activity detected: triggering detail sync")
			s.runSync(ctx, func(ctx context.Context) (observer.SyncResult, error) {
				return s.runner.RunDetailSync(ctx)
			})
		}
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

	now := time.Now()
	s.status.mu.Lock()
	s.status.progress = &model.SyncProgress{StartedAt: now}
	s.status.mu.Unlock()
	defer func() {
		s.status.mu.Lock()
		s.status.progress = nil
		s.status.mu.Unlock()
	}()

	s.logger.Printf("sync started")
	start := now

	// Use a detached context so in-progress syncs finish during shutdown.
	syncCtx := context.WithoutCancel(ctx)
	result, err := fn(syncCtx)
	duration := time.Since(start)

	if err != nil {
		s.logger.Printf("sync aborted: %v (duration=%s)", err, duration)
		s.recordError(result, duration, err)
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

func (s *Scheduler) recordError(r observer.SyncResult, d time.Duration, syncErr error) {
	now := time.Now().UTC()
	entry := model.SyncLogEntry{
		Timestamp: now,
		Mode:      r.Mode,
		Pages:     r.PagesFetched,
		Topics:    r.TopicsStored,
		Duration:  d,
		Error:     syncErr.Error(),
	}

	if s.logStore != nil {
		ctx := context.Background()
		if err := s.logStore.SaveSyncLogEntry(ctx, &entry); err != nil {
			s.logger.Printf("failed to persist sync error log: %v", err)
		}
	}

	s.status.mu.Lock()
	defer s.status.mu.Unlock()
	s.status.log = append([]model.SyncLogEntry{entry}, s.status.log...)
}

func (s *Scheduler) recordResult(r observer.SyncResult, d time.Duration) {
	now := time.Now().UTC()
	entry := model.SyncLogEntry{
		Timestamp:  now,
		Mode:       r.Mode,
		Pages:      r.PagesFetched,
		Topics:     r.TopicsStored,
		Duration:   d,
		HasChanges: r.HasChanges,
	}

	if s.logStore != nil {
		ctx := context.Background()
		if err := s.logStore.SaveSyncLogEntry(ctx, &entry); err != nil {
			s.logger.Printf("failed to persist sync log: %v", err)
		}
	}

	s.status.mu.Lock()
	defer s.status.mu.Unlock()
	s.status.LastDuration = d
	s.status.LastTopics = r.TopicsStored
	s.status.LastSyncedAt = &now
	s.status.log = append([]model.SyncLogEntry{entry}, s.status.log...)
}

func (s *Scheduler) trackLowActivity(r observer.SyncResult) {
	if r.Mode == "detail" {
		return // detail sync results don't affect activity tracking
	}
	if r.TopicsStored > 0 {
		s.zeroStreak = 0
		return
	}
	s.zeroStreak++
}

// shouldRunDetailSync returns true if conditions indicate a low-activity
// window suitable for detail sync.
func (s *Scheduler) shouldRunDetailSync(ctx context.Context) bool {
	if s.activityData != nil {
		grid, err := s.activityData.ActivityByHour(ctx)
		if err == nil {
			return isLowActivityHour(&grid, time.Now().UTC())
		}
		// Fall through to heuristic on error.
	}
	return s.zeroStreak >= s.cfg.LowActivityThreshold
}

// lowActivityThresholdPct is the fraction of peak activity below which
// an hour is considered low-activity. 0.2 = hours with ≤20% of peak.
const lowActivityThresholdPct = 0.2

// isLowActivityHour checks if the given time falls in a low-activity hour
// based on historical activity data.
func isLowActivityHour(grid *[7][24]int, now time.Time) bool {
	var peak int
	for d := 0; d < 7; d++ {
		for h := 0; h < 24; h++ {
			if grid[d][h] > peak {
				peak = grid[d][h]
			}
		}
	}
	if peak == 0 {
		return false // no data yet
	}

	day := (int(now.Weekday()) + 6) % 7 // Monday=0
	hour := now.Hour()
	threshold := int(float64(peak) * lowActivityThresholdPct)
	return grid[day][hour] <= threshold
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
