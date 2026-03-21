// Spec: specs/observer/sync-metadata.md
// Tests sync metadata storage: watermark, last page, and detail sync methods.
package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestWatermarkRoundTrip(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	ts := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	if err := store.SaveWatermark(ctx, ts); err != nil {
		t.Fatalf("SaveWatermark: %v", err)
	}

	got, err := store.LoadWatermark(ctx)
	if err != nil {
		t.Fatalf("LoadWatermark: %v", err)
	}
	if got == nil || !got.Equal(ts) {
		t.Errorf("LoadWatermark = %v, want %v", got, ts)
	}
}

func TestWatermarkAbsent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	got, err := store.LoadWatermark(ctx)
	if err != nil {
		t.Fatalf("LoadWatermark: %v", err)
	}
	if got != nil {
		t.Errorf("LoadWatermark = %v, want nil", got)
	}
}

func TestWatermarkOverwrite(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	ts1 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	if err := store.SaveWatermark(ctx, ts1); err != nil {
		t.Fatalf("SaveWatermark(1): %v", err)
	}
	if err := store.SaveWatermark(ctx, ts2); err != nil {
		t.Fatalf("SaveWatermark(2): %v", err)
	}

	got, err := store.LoadWatermark(ctx)
	if err != nil {
		t.Fatalf("LoadWatermark: %v", err)
	}
	if got == nil || !got.Equal(ts2) {
		t.Errorf("LoadWatermark = %v, want %v", got, ts2)
	}
}

func TestLastPageRoundTrip(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SaveLastPage(ctx, 5); err != nil {
		t.Fatalf("SaveLastPage: %v", err)
	}

	got, err := store.LoadLastPage(ctx)
	if err != nil {
		t.Fatalf("LoadLastPage: %v", err)
	}
	if got != 5 {
		t.Errorf("LoadLastPage = %d, want 5", got)
	}
}

func TestLastPageAbsent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	got, err := store.LoadLastPage(ctx)
	if err != nil {
		t.Fatalf("LoadLastPage: %v", err)
	}
	if got != -1 {
		t.Errorf("LoadLastPage = %d, want -1", got)
	}
}

func TestLastPageClear(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SaveLastPage(ctx, 3); err != nil {
		t.Fatalf("SaveLastPage: %v", err)
	}
	if err := store.ClearLastPage(ctx); err != nil {
		t.Fatalf("ClearLastPage: %v", err)
	}

	got, err := store.LoadLastPage(ctx)
	if err != nil {
		t.Fatalf("LoadLastPage: %v", err)
	}
	if got != -1 {
		t.Errorf("LoadLastPage after clear = %d, want -1", got)
	}
}

func TestDetailSyncSortsSyncedAfterUnsynced(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	seedTopics(t, store, ctx, 1001, 1002)

	ts := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)
	if err := store.SaveDetailSync(ctx, 1001, 3, ts); err != nil {
		t.Fatalf("SaveDetailSync: %v", err)
	}

	states, err := store.TopicsNeedingDetailSync(ctx, 10)
	if err != nil {
		t.Fatalf("TopicsNeedingDetailSync: %v", err)
	}
	// Unsynced topic 1002 should appear before synced topic 1001.
	want := []int{1002, 1001}
	if len(states) != len(want) {
		t.Fatalf("got %d states, want %d", len(states), len(want))
	}
	for i := range want {
		if states[i].TopicID != want[i] {
			t.Errorf("states[%d].TopicID = %d, want %d", i, states[i].TopicID, want[i])
			break
		}
	}
}

func TestTopicsNeedingDetailSync(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	seedTopics(t, store, ctx, 1, 2, 3, 4, 5)

	// Mark topics 2 and 4 as synced (4 older, 2 newer).
	old := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	if err := store.SaveDetailSync(ctx, 4, 2, old); err != nil {
		t.Fatalf("SaveDetailSync(4): %v", err)
	}
	if err := store.SaveDetailSync(ctx, 2, 5, recent); err != nil {
		t.Fatalf("SaveDetailSync(2): %v", err)
	}

	states, err := store.TopicsNeedingDetailSync(ctx, 10)
	if err != nil {
		t.Fatalf("TopicsNeedingDetailSync: %v", err)
	}

	// Expected order: unsynced (1, 3, 5) first, then stale→recent (4, 2).
	want := []int{1, 3, 5, 4, 2}
	if len(states) != len(want) {
		t.Fatalf("got %d states, want %d", len(states), len(want))
	}
	for i := range want {
		if states[i].TopicID != want[i] {
			t.Errorf("states[%d].TopicID = %d, want %d", i, states[i].TopicID, want[i])
			break
		}
	}
}

func TestTopicsNeedingDetailSyncLimit(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	seedTopics(t, store, ctx, 10, 20, 30)

	ids, err := store.TopicsNeedingDetailSync(ctx, 2)
	if err != nil {
		t.Fatalf("TopicsNeedingDetailSync: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("got %d results, want 2 (ids: %v)", len(ids), ids)
	}
}

// --- Sync log error tests (SE-3, SE-6, SE-7) ---

func TestSyncLogErrorColumn(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	now := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// Save a successful entry.
	success := model.SyncLogEntry{
		Timestamp: now, Mode: "delta", Pages: 1, Topics: 5,
		Duration: 2 * time.Second, HasChanges: true,
	}
	if err := store.SaveSyncLogEntry(ctx, success); err != nil {
		t.Fatalf("save success: %v", err)
	}

	// Save an error entry.
	errEntry := model.SyncLogEntry{
		Timestamp: now.Add(time.Minute), Mode: "delta",
		Pages: 0, Topics: 0, Duration: 500 * time.Millisecond,
		Error: "fetch page 0: unexpected status 500",
	}
	if err := store.SaveSyncLogEntry(ctx, errEntry); err != nil {
		t.Fatalf("save error: %v", err)
	}

	entries, err := store.LoadSyncLog(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	// Newest first — error entry is first.
	if entries[0].Error != "fetch page 0: unexpected status 500" {
		t.Errorf("error entry: Error = %q, want error message", entries[0].Error)
	}
	if entries[1].Error != "" {
		t.Errorf("success entry: Error = %q, want empty", entries[1].Error)
	}
}

func TestSyncLogErrorRetention(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	base := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)

	// Error entries should not be deduplicated like no-change entries.
	for i := 0; i < 3; i++ {
		e := model.SyncLogEntry{
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Mode:      "delta", Error: "connection refused",
		}
		if err := store.SaveSyncLogEntry(ctx, e); err != nil {
			t.Fatalf("save error %d: %v", i, err)
		}
	}

	entries, err := store.LoadSyncLog(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// All 3 error entries should be kept (not deduplicated).
	if len(entries) != 3 {
		t.Errorf("got %d entries, want 3 (error entries should not be deduplicated)", len(entries))
	}
}

// seedTopics inserts minimal topics into the database so that
// TopicsNeedingDetailSync can join against the topics table.
func seedTopics(t *testing.T, store *SQLiteStore, ctx context.Context, ids ...int) {
	t.Helper()
	topics := make([]model.Topic, len(ids))
	for i, id := range ids {
		topics[i] = model.Topic{
			ID:        id,
			Title:     "test",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Tags:      []string{},
		}
	}
	if err := store.StoreTopics(ctx, topics); err != nil {
		t.Fatalf("StoreTopics: %v", err)
	}
}
