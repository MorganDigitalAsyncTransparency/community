// Spec: specs/observer/scheduler.md
package scheduler_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/config"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/scheduler"
)

// --- Fakes ---

// fakeSyncRunner records calls and returns configurable results.
type fakeSyncRunner struct {
	runCalls      atomic.Int32
	deltaCalls    atomic.Int32
	runResult     observer.SyncResult
	deltaResult   observer.SyncResult
	runDelay      time.Duration // simulate slow sync
	deltaDelay    time.Duration
	deltaResults  []observer.SyncResult // per-call delta results (if set, overrides deltaResult)
	deltaIdx      atomic.Int32
}

func (f *fakeSyncRunner) Run(_ context.Context) (observer.SyncResult, error) {
	f.runCalls.Add(1)
	if f.runDelay > 0 {
		time.Sleep(f.runDelay)
	}
	return f.runResult, nil
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
		runResult:  observer.SyncResult{Mode: "initial", TopicsStored: 1},
		deltaDelay: 30 * time.Millisecond, // slow delta sync
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
		runResult:  observer.SyncResult{Mode: "initial", TopicsStored: 1},
		runDelay:   50 * time.Millisecond, // slow initial sync
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

	snap := sched.Status().Snapshot()
	if snap.State != "idle" {
		t.Errorf("state = %q, want idle", snap.State)
	}
	if snap.LastTopics != 7 {
		t.Errorf("LastTopics = %d, want 7", snap.LastTopics)
	}
	if snap.LastSyncedAt == nil {
		t.Error("LastSyncedAt should be set after sync")
	}
}
