// Spec: specs/observer/initial-delta-sync.md
// Tests initial sync, delta sync, auto-detect, resume, and watermark stop.
package main_test

import (
	"context"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/storage"
)

// Compile-time interface checks (R1, R2, R3).
var _ observer.FetchClient = (*discourse.Client)(nil)
var _ observer.StorageBackend = (*storage.SQLiteStore)(nil)

const testPageSize = 5

func newTestEnv(t *testing.T) (*observer.Observer, *storage.SQLiteStore, *discourse.Client) {
	t.Helper()
	srv := mockserver.NewWithPageSize(testPageSize)
	t.Cleanup(srv.Close)

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	client := discourse.NewClient(srv.URL, "", "")
	obs := observer.New(client, store, srv.URL)
	return obs, store, client
}

func TestMockServerSortOrder(t *testing.T) {
	srv := mockserver.NewWithPageSize(100)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	topics, err := client.FetchTopics(context.Background())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	for i := 1; i < len(topics); i++ {
		prev := topics[i-1].BumpedAt
		curr := topics[i].BumpedAt
		if prev == nil || curr == nil {
			t.Fatalf("topic %d or %d has nil BumpedAt", topics[i-1].ID, topics[i].ID)
		}
		if curr.After(*prev) {
			t.Errorf("topics not sorted by bumped_at desc: topic[%d].BumpedAt=%v > topic[%d].BumpedAt=%v",
				i, *curr, i-1, *prev)
		}
	}
}

func TestInitialSyncEndToEnd(t *testing.T) {
	obs, store, _ := newTestEnv(t)
	ctx := context.Background()

	result, err := obs.RunInitialSync(ctx)
	if err != nil {
		t.Fatalf("RunInitialSync: %v", err)
	}

	if result.Mode != "initial" {
		t.Errorf("mode = %q, want initial", result.Mode)
	}
	if result.TopicsStored != 44 {
		t.Errorf("topics stored = %d, want 44", result.TopicsStored)
	}
	// With pageSize=5 and 44 topics: 9 pages (5*8 + 4)
	if result.PagesFetched != 9 {
		t.Errorf("pages fetched = %d, want 9", result.PagesFetched)
	}
	if result.NewWatermark == nil {
		t.Fatal("watermark is nil after initial sync")
	}

	// Verify watermark is stored.
	wm, err := store.LoadWatermark(ctx)
	if err != nil {
		t.Fatalf("load watermark: %v", err)
	}
	if wm == nil || !wm.Equal(*result.NewWatermark) {
		t.Errorf("stored watermark = %v, want %v", wm, result.NewWatermark)
	}

	// Verify last page is cleared.
	lp, err := store.LoadLastPage(ctx)
	if err != nil {
		t.Fatalf("load last page: %v", err)
	}
	if lp != -1 {
		t.Errorf("last page = %d, want -1 (cleared)", lp)
	}

	// Verify all topics stored.
	topics, err := store.LoadTopics(ctx)
	if err != nil {
		t.Fatalf("load topics: %v", err)
	}
	if len(topics) != 44 {
		t.Errorf("stored topics = %d, want 44", len(topics))
	}
}

func TestInitialSyncResume(t *testing.T) {
	obs, store, _ := newTestEnv(t)
	ctx := context.Background()

	// Simulate interrupted sync at page 5.
	if err := store.SaveLastPage(ctx, 5); err != nil {
		t.Fatalf("save last page: %v", err)
	}

	result, err := obs.RunInitialSync(ctx)
	if err != nil {
		t.Fatalf("RunInitialSync: %v", err)
	}

	// With 44 topics, pageSize=5: pages 0–8. Resuming from page 6 (5+1)
	// means pages 6,7,8 = 3 pages.
	if result.PagesFetched != 3 {
		t.Errorf("pages fetched = %d, want 3 (resumed from page 6)", result.PagesFetched)
	}

	// Only topics from pages 6–8 are stored (not all 44).
	// Pages 6,7 have 5 topics each, page 8 has 4 = 14 topics.
	if result.TopicsStored != 14 {
		t.Errorf("topics stored = %d, want 14", result.TopicsStored)
	}
}

func TestDeltaSyncEndToEnd(t *testing.T) {
	obs, _, _ := newTestEnv(t)
	ctx := context.Background()

	// First do an initial sync.
	initResult, err := obs.RunInitialSync(ctx)
	if err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	// Now do a delta sync — with no changes, it should still succeed.
	deltaResult, err := obs.RunDeltaSync(ctx)
	if err != nil {
		t.Fatalf("delta sync: %v", err)
	}

	if deltaResult.Mode != "delta" {
		t.Errorf("mode = %q, want delta", deltaResult.Mode)
	}
	if deltaResult.NewWatermark == nil {
		t.Fatal("watermark is nil after delta sync")
	}
	if !deltaResult.NewWatermark.Equal(*initResult.NewWatermark) {
		t.Errorf("delta watermark = %v, want %v (unchanged)", deltaResult.NewWatermark, initResult.NewWatermark)
	}
}

func TestDeltaSyncStopsAtWatermark(t *testing.T) {
	obs, store, client := newTestEnv(t)
	ctx := context.Background()

	// Fetch all topics to find a watermark that lets some pages through.
	allTopics, err := client.FetchTopics(ctx)
	if err != nil {
		t.Fatalf("fetch all: %v", err)
	}

	// Sort by bumped_at desc (matching server order).
	sort.Slice(allTopics, func(i, j int) bool {
		bi, bj := allTopics[i].BumpedAt, allTopics[j].BumpedAt
		if bi == nil || bj == nil {
			return false
		}
		return bi.After(*bj)
	})

	// Set watermark to the bumped_at of topic at position 7 (start of page 1
	// with pageSize=5). This means page 0 has newer topics, page 1 should
	// have the watermark topic, and delta sync should stop at page 1.
	wmTopic := allTopics[7]
	if wmTopic.BumpedAt == nil {
		t.Fatal("watermark topic has nil BumpedAt")
	}
	if err := store.SaveWatermark(ctx, *wmTopic.BumpedAt); err != nil {
		t.Fatalf("save watermark: %v", err)
	}

	result, err := obs.RunDeltaSync(ctx)
	if err != nil {
		t.Fatalf("delta sync: %v", err)
	}

	// Should fetch only a few pages, not all 9.
	if result.PagesFetched >= 9 {
		t.Errorf("pages fetched = %d, want < 9 (should stop early)", result.PagesFetched)
	}
}

func TestDeltaSyncWithoutWatermarkFails(t *testing.T) {
	obs, _, _ := newTestEnv(t)
	ctx := context.Background()

	_, err := obs.RunDeltaSync(ctx)
	if err == nil {
		t.Fatal("expected error from RunDeltaSync without watermark")
	}
}

func TestRunAutoDetectsMode(t *testing.T) {
	obs, _, _ := newTestEnv(t)
	ctx := context.Background()

	// First Run — no watermark, should do initial sync.
	r1, err := obs.Run(ctx)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if r1.Mode != "initial" {
		t.Errorf("first run mode = %q, want initial", r1.Mode)
	}
	if r1.TopicsStored != 44 {
		t.Errorf("first run topics = %d, want 44", r1.TopicsStored)
	}

	// Second Run — watermark exists, should do delta sync.
	r2, err := obs.Run(ctx)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if r2.Mode != "delta" {
		t.Errorf("second run mode = %q, want delta", r2.Mode)
	}
}

func TestSyncResultFields(t *testing.T) {
	obs, _, _ := newTestEnv(t)
	ctx := context.Background()

	result, err := obs.RunInitialSync(ctx)
	if err != nil {
		t.Fatalf("RunInitialSync: %v", err)
	}

	if result.Mode != "initial" {
		t.Errorf("Mode = %q, want initial", result.Mode)
	}
	if result.PagesFetched == 0 {
		t.Error("PagesFetched = 0, want > 0")
	}
	if result.TopicsStored == 0 {
		t.Error("TopicsStored = 0, want > 0")
	}
	if result.NewWatermark == nil {
		t.Error("NewWatermark is nil")
	}
	if result.Duration <= 0 {
		t.Error("Duration <= 0")
	}
}

func TestWatermarkIsMaxBumpedAt(t *testing.T) {
	obs, store, client := newTestEnv(t)
	ctx := context.Background()

	// Find the actual max bumped_at across all topics.
	allTopics, err := client.FetchTopics(ctx)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	var maxBump time.Time
	for i := range allTopics {
		if b := allTopics[i].BumpedAt; b != nil && b.After(maxBump) {
			maxBump = *b
		}
	}

	result, err := obs.RunInitialSync(ctx)
	if err != nil {
		t.Fatalf("RunInitialSync: %v", err)
	}

	if !result.NewWatermark.Equal(maxBump) {
		t.Errorf("watermark = %v, want %v (max bumped_at)", result.NewWatermark, maxBump)
	}

	wm, _ := store.LoadWatermark(ctx)
	if !wm.Equal(maxBump) {
		t.Errorf("stored watermark = %v, want %v", wm, maxBump)
	}
}
