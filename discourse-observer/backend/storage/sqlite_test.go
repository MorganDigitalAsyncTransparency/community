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

func TestDetailSyncRoundTrip(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	ts := time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC)

	if err := store.SaveDetailSync(ctx, 1001, ts); err != nil {
		t.Fatalf("SaveDetailSync: %v", err)
	}

	// Seed a topic so TopicsNeedingDetailSync can join against it.
	seedTopics(t, store, ctx, 1001)

	ids, err := store.TopicsNeedingDetailSync(ctx, 10)
	if err != nil {
		t.Fatalf("TopicsNeedingDetailSync: %v", err)
	}
	// Topic 1001 is the only topic and it has been synced, but it still
	// appears because there are no unsynced topics to prioritize over it.
	if len(ids) != 1 || ids[0] != 1001 {
		t.Errorf("TopicsNeedingDetailSync = %v, want [1001]", ids)
	}
}

func TestTopicsNeedingDetailSync(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	seedTopics(t, store, ctx, 1, 2, 3, 4, 5)

	// Mark topics 2 and 4 as synced (4 older, 2 newer).
	old := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	if err := store.SaveDetailSync(ctx, 4, old); err != nil {
		t.Fatalf("SaveDetailSync(4): %v", err)
	}
	if err := store.SaveDetailSync(ctx, 2, recent); err != nil {
		t.Fatalf("SaveDetailSync(2): %v", err)
	}

	ids, err := store.TopicsNeedingDetailSync(ctx, 10)
	if err != nil {
		t.Fatalf("TopicsNeedingDetailSync: %v", err)
	}

	// Expected order: unsynced (1, 3, 5) first, then stale→recent (4, 2).
	want := []int{1, 3, 5, 4, 2}
	if len(ids) != len(want) {
		t.Fatalf("got %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("ids[%d] = %d, want %d (full: %v)", i, ids[i], want[i], ids)
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
