// Spec: specs/observer/scheduler.md
package scheduler_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/config"
	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/scheduler"
)

// --- Fakes ---

// fakeSyncRunner records calls and returns configurable results.
type fakeSyncRunner struct {
	runCalls     atomic.Int32
	deltaCalls   atomic.Int32
	detailCalls  atomic.Int32
	runResult    observer.SyncResult
	runErr       error // if set, Run returns this error
	deltaResult  observer.SyncResult
	detailResult observer.SyncResult
	runDelay     time.Duration // simulate slow sync
	deltaDelay   time.Duration
	deltaResults []observer.SyncResult // per-call delta results (if set, overrides deltaResult)
	deltaIdx     atomic.Int32
}

func (f *fakeSyncRunner) Run(_ context.Context) (observer.SyncResult, error) {
	f.runCalls.Add(1)
	if f.runDelay > 0 {
		time.Sleep(f.runDelay)
	}
	return f.runResult, f.runErr
}

func (f *fakeSyncRunner) RunDeltaSync(_ context.Context) (observer.SyncResult, error) {
	f.deltaCalls.Add(1)
	if f.deltaDelay > 0 {
		time.Sleep(f.deltaDelay)
	}
	if len(f.deltaResults) > 0 {
		idx := int(f.deltaIdx.Add(1)) - 1
		if idx < len(f.deltaResults) {
			return f.deltaResults[idx], nil
		}
	}
	return f.deltaResult, nil
}

func (f *fakeSyncRunner) RunDetailSync(_ context.Context) (observer.SyncResult, error) {
	f.detailCalls.Add(1)
	return f.detailResult, nil
}

// Compile-time check: *observer.Observer must satisfy SyncRunner.
var _ scheduler.SyncRunner = (*observer.Observer)(nil)

func shortCfg() config.SyncConfig {
	return config.SyncConfig{
		Interval:             20 * time.Millisecond,
		JitterMax:            0,
		LowActivityThreshold: 3,
	}
}

// --- Config tests (SC-1, SC-2) ---

func TestSyncConfigDefaults(t *testing.T) {
	cfg := config.LoadSyncConfig()

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"InitialDelay", cfg.InitialDelay, 20 * time.Second},
		{"DeltaDelay", cfg.DeltaDelay, 2 * time.Second},
		{"Interval", cfg.Interval, 15 * time.Minute},
		{"LowActivityThreshold", cfg.LowActivityThreshold, 3},
		{"MaxRetries", cfg.MaxRetries, 3},
		{"JitterMax", cfg.JitterMax, 60 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestSyncConfigFromEnv(t *testing.T) {
	t.Setenv("SYNC_INITIAL_DELAY_SECONDS", "10")
	t.Setenv("SYNC_DELTA_DELAY_SECONDS", "5")
	t.Setenv("SYNC_INTERVAL", "300")
	t.Setenv("SYNC_LOW_ACTIVITY_THRESHOLD", "5")
	t.Setenv("SYNC_MAX_RETRIES", "7")
	t.Setenv("SYNC_JITTER_SECONDS", "30")

	cfg := config.LoadSyncConfig()

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"InitialDelay", cfg.InitialDelay, 10 * time.Second},
		{"DeltaDelay", cfg.DeltaDelay, 5 * time.Second},
		{"Interval", cfg.Interval, 300 * time.Second},
		{"LowActivityThreshold", cfg.LowActivityThreshold, 5},
		{"MaxRetries", cfg.MaxRetries, 7},
		{"JitterMax", cfg.JitterMax, 30 * time.Second},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// --- Scheduler behavior tests ---

func TestSchedulerRunsImmediately(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult: observer.SyncResult{Mode: "initial", TopicsStored: 10},
	}
	cfg := shortCfg()
	cfg.Interval = 1 * time.Hour // long interval so only the first sync fires

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	if got := fake.runCalls.Load(); got != 1 {
		t.Errorf("Run calls = %d, want 1", got)
	}
}

func TestSchedulerRunsOnInterval(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult:   observer.SyncResult{Mode: "delta", TopicsStored: 5},
		deltaResult: observer.SyncResult{Mode: "delta", TopicsStored: 3},
	}
	cfg := shortCfg()
	cfg.Interval = 10 * time.Millisecond

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	// After ~80ms with 10ms interval, expect multiple delta calls.
	if got := fake.deltaCalls.Load(); got < 2 {
		t.Errorf("delta calls = %d, want >= 2", got)
	}
}

func TestSchedulerJitter(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult:   observer.SyncResult{Mode: "delta", TopicsStored: 1},
		deltaResult: observer.SyncResult{Mode: "delta", TopicsStored: 1},
	}
	cfg := shortCfg()
	cfg.Interval = 5 * time.Millisecond
	cfg.JitterMax = 5 * time.Millisecond

	// Run several times and verify syncs still happen (jitter doesn't break timing).
	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	if got := fake.deltaCalls.Load(); got < 1 {
		t.Errorf("delta calls = %d, want >= 1 (jitter should not prevent syncs)", got)
	}
}

func TestSchedulerConcurrencyGuard(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult:   observer.SyncResult{Mode: "initial", TopicsStored: 1},
		deltaDelay:  30 * time.Millisecond, // slow delta sync
		deltaResult: observer.SyncResult{Mode: "delta", TopicsStored: 1},
	}
	cfg := shortCfg()
	cfg.Interval = 5 * time.Millisecond // fires faster than delta completes

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	// With 30ms delta and 5ms interval over 100ms, without guard we'd see many calls.
	// With guard, calls should be limited (each takes 30ms so max ~3).
	if got := fake.deltaCalls.Load(); got > 4 {
		t.Errorf("delta calls = %d, want <= 4 (concurrency guard should prevent overlapping)", got)
	}
}

func TestSchedulerLowActivityDetection(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	fake := &fakeSyncRunner{
		runResult: observer.SyncResult{Mode: "delta", TopicsStored: 0},
		deltaResults: []observer.SyncResult{
			{Mode: "delta", TopicsStored: 0},
			{Mode: "delta", TopicsStored: 0},
			{Mode: "delta", TopicsStored: 0},
		},
	}
	cfg := shortCfg()
	cfg.Interval = 5 * time.Millisecond
	cfg.LowActivityThreshold = 3

	sched := scheduler.NewWithLogger(fake, cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	if !strings.Contains(buf.String(), "low activity") {
		t.Errorf("expected low-activity log message, got: %s", buf.String())
	}
}

func TestSchedulerGracefulShutdown(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult: observer.SyncResult{Mode: "initial", TopicsStored: 1},
		runDelay:  50 * time.Millisecond, // slow initial sync
	}
	cfg := shortCfg()
	cfg.Interval = 1 * time.Hour

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.Start(ctx)
		close(done)
	}()

	// Cancel while first sync is in progress.
	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Start returned — verify the in-progress sync completed.
		if got := fake.runCalls.Load(); got != 1 {
			t.Errorf("Run calls = %d, want 1 (sync should complete before shutdown)", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestStatusReflectsSyncState(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult: observer.SyncResult{
			Mode:         "delta",
			TopicsStored: 7,
			Duration:     42 * time.Millisecond,
		},
	}
	cfg := shortCfg()
	cfg.Interval = 1 * time.Hour

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	status := sched.Status()
	if got := status.GetState(); got != "idle" {
		t.Errorf("state = %q, want idle", got)
	}
	if got := status.GetLastTopics(); got != 7 {
		t.Errorf("LastTopics = %d, want 7", got)
	}
	if status.GetLastSyncedAt() == nil {
		t.Error("LastSyncedAt should be set after sync")
	}
}

// --- Detail sync tests (DS-17, DS-18, DS-19, DS-20) ---

// fakeActivityData returns a fixed activity grid for testing.
type fakeActivityData struct {
	grid [7][24]int
	err  error
}

func (f *fakeActivityData) ActivityByHour(_ context.Context) ([7][24]int, error) {
	return f.grid, f.err
}

func TestSchedulerTriggersDetailSync(t *testing.T) {
	// Set up: delta sync returns zero topics → triggers low activity detection.
	fake := &fakeSyncRunner{
		runResult:    observer.SyncResult{Mode: "initial", TopicsStored: 10},
		deltaResult:  observer.SyncResult{Mode: "delta", TopicsStored: 0},
		detailResult: observer.SyncResult{Mode: "detail", TopicsStored: 3},
	}
	cfg := shortCfg()
	cfg.LowActivityThreshold = 1 // trigger after 1 zero-change sync

	sched := scheduler.New(fake, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	// After initial + at least one delta (zero topics) → detail sync should run.
	if got := fake.detailCalls.Load(); got == 0 {
		t.Error("expected at least one RunDetailSync call after low activity")
	}
}

func TestSchedulerLowActivityWindow(t *testing.T) {
	// Build a grid where the current hour is low activity.
	now := time.Now().UTC()
	day := (int(now.Weekday()) + 6) % 7
	hour := now.Hour()

	var grid [7][24]int
	// Set peak at a different hour.
	peakHour := (hour + 12) % 24
	grid[day][peakHour] = 100
	// Current hour has 0 activity → should be detected as low.
	grid[day][hour] = 0

	activity := &fakeActivityData{grid: grid}

	fake := &fakeSyncRunner{
		runResult:    observer.SyncResult{Mode: "initial", TopicsStored: 10},
		deltaResult:  observer.SyncResult{Mode: "delta", TopicsStored: 0},
		detailResult: observer.SyncResult{Mode: "detail", TopicsStored: 1},
	}
	cfg := shortCfg()
	cfg.LowActivityThreshold = 100 // high threshold so heuristic alone wouldn't trigger

	sched := scheduler.New(fake, cfg)
	sched.SetActivityData(activity)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	if got := fake.detailCalls.Load(); got == 0 {
		t.Error("expected detail sync when current hour is low-activity per heatmap")
	}
}

// --- Sync error logging tests (SE-1, SE-2) ---

// fakeLogStore captures saved sync log entries for testing.
type fakeLogStore struct {
	entries []model.SyncLogEntry
}

func (f *fakeLogStore) SaveSyncLogEntry(_ context.Context, e *model.SyncLogEntry) error {
	f.entries = append(f.entries, *e)
	return nil
}

func (f *fakeLogStore) LoadSyncLog(_ context.Context) ([]model.SyncLogEntry, error) {
	return f.entries, nil
}

func TestSyncErrorSavedToLog(t *testing.T) {
	fake := &fakeSyncRunner{
		runResult: observer.SyncResult{Mode: "initial", PagesFetched: 2, TopicsStored: 15},
		runErr:    fmt.Errorf("fetch page 3: unexpected status 500 from /latest.json"),
	}
	cfg := shortCfg()
	cfg.Interval = 1 * time.Hour

	logStore := &fakeLogStore{}
	sched := scheduler.New(fake, cfg)
	sched.SetLogStore(logStore)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sched.Start(ctx)

	// The failed sync should produce an error entry in the log.
	if len(logStore.entries) == 0 {
		t.Fatal("expected at least one log entry after failed sync")
	}
	entry := logStore.entries[0]
	if entry.Error == "" {
		t.Error("expected non-empty error field on failed sync entry")
	}
	if entry.Mode != "initial" {
		t.Errorf("mode = %q, want initial", entry.Mode)
	}
	if !strings.Contains(entry.Error, "500") {
		t.Errorf("error = %q, expected to contain status code", entry.Error)
	}

	// Error entry should also appear in the in-memory status log.
	statusLog := sched.Status().GetLog()
	if len(statusLog) == 0 {
		t.Fatal("expected status log to contain error entry")
	}
	if statusLog[0].Error == "" {
		t.Error("expected status log entry to have non-empty error")
	}
}
