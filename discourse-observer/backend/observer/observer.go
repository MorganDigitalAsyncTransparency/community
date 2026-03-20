// Spec: specs/observer/observer-behavior.md, specs/observer/initial-delta-sync.md
// Tests: backend/pipeline_test.go, backend/sync_test.go
package observer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// FetchClient fetches raw data from a Discourse-compatible API.
type FetchClient interface {
	FetchTopics(ctx context.Context) ([]model.RawTopic, error)
	FetchTopicsPages(ctx context.Context, startPage int, fn func(topics []model.RawTopic, page int) error) error
	FetchCategories(ctx context.Context) ([]model.RawCategory, error)
}

// StorageBackend persists normalized topics and sync metadata.
type StorageBackend interface {
	StoreTopics(ctx context.Context, topics []model.Topic) error
	SaveWatermark(ctx context.Context, t time.Time) error
	LoadWatermark(ctx context.Context) (*time.Time, error)
	SaveLastPage(ctx context.Context, page int) error
	LoadLastPage(ctx context.Context) (int, error)
	ClearLastPage(ctx context.Context) error
	Close() error
}

// SyncResult reports what a sync cycle did.
type SyncResult struct {
	Mode         string
	PagesFetched int
	TopicsStored int
	NewWatermark *time.Time
	Duration     time.Duration
}

// errStopPagination is a sentinel error used by delta sync to signal that
// all topics on a page are at or below the watermark.
var errStopPagination = errors.New("stop pagination")

// Observer coordinates fetch, normalization, and storage.
type Observer struct {
	fetch   FetchClient
	store   StorageBackend
	baseURL string
}

// New creates an Observer. baseURL is the forum URL used to construct topic links.
func New(fetch FetchClient, store StorageBackend, baseURL string) *Observer {
	return &Observer{fetch: fetch, store: store, baseURL: baseURL}
}

// Run executes one sync cycle, auto-detecting mode:
// no watermark → initial sync, watermark exists → delta sync.
func (o *Observer) Run(ctx context.Context) (SyncResult, error) {
	wm, err := o.store.LoadWatermark(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("load watermark: %w", err)
	}
	if wm == nil {
		return o.RunInitialSync(ctx)
	}
	return o.RunDeltaSync(ctx)
}

// RunInitialSync performs a full crawl of all pages.
func (o *Observer) RunInitialSync(ctx context.Context) (SyncResult, error) {
	start := time.Now()

	catMap, err := o.fetchCategoryMap(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	startPage, err := o.resumePage(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	var maxBump time.Time
	result := SyncResult{Mode: "initial"}

	err = o.fetch.FetchTopicsPages(ctx, startPage, func(raws []model.RawTopic, page int) error {
		if err := o.storeAndTrack(ctx, raws, catMap, &result, &maxBump); err != nil {
			return fmt.Errorf("page %d: %w", page, err)
		}
		if err := o.store.SaveLastPage(ctx, page); err != nil {
			return fmt.Errorf("save last page %d: %w", page, err)
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	if err := o.finalizeInitialSync(ctx, &result, maxBump); err != nil {
		return result, err
	}
	result.Duration = time.Since(start)
	return result, nil
}

// RunDeltaSync fetches pages until all topics on a page are at or below
// the stored watermark.
func (o *Observer) RunDeltaSync(ctx context.Context) (SyncResult, error) {
	start := time.Now()

	wm, err := o.store.LoadWatermark(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("load watermark: %w", err)
	}
	if wm == nil {
		return SyncResult{}, fmt.Errorf("no watermark found: run initial sync first")
	}

	catMap, err := o.fetchCategoryMap(ctx)
	if err != nil {
		return SyncResult{}, err
	}

	var maxBump time.Time
	result := SyncResult{Mode: "delta"}

	err = o.fetch.FetchTopicsPages(ctx, 0, func(raws []model.RawTopic, page int) error {
		if err := o.storeAndTrack(ctx, raws, catMap, &result, &maxBump); err != nil {
			return fmt.Errorf("page %d: %w", page, err)
		}
		if allAtOrBelowWatermark(raws, *wm) {
			return errStopPagination
		}
		return nil
	})
	if err != nil && !errors.Is(err, errStopPagination) {
		return result, err
	}

	newWM := *wm
	if maxBump.After(newWM) {
		newWM = maxBump
	}
	if err := o.store.SaveWatermark(ctx, newWM); err != nil {
		return result, fmt.Errorf("save watermark: %w", err)
	}
	result.NewWatermark = &newWM
	result.Duration = time.Since(start)
	return result, nil
}

// fetchCategoryMap fetches categories and builds an ID→name lookup.
func (o *Observer) fetchCategoryMap(ctx context.Context) (map[int]string, error) {
	cats, err := o.fetch.FetchCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch categories: %w", err)
	}
	return buildCategoryMap(cats), nil
}

// resumePage returns the page to start from based on stored progress.
// Returns 0 if no progress is stored.
func (o *Observer) resumePage(ctx context.Context) (int, error) {
	lastPage, err := o.store.LoadLastPage(ctx)
	if err != nil {
		return 0, fmt.Errorf("load last page: %w", err)
	}
	if lastPage >= 0 {
		return lastPage + 1, nil
	}
	return 0, nil
}

// storeAndTrack normalizes a page of topics, stores them, and updates counters.
func (o *Observer) storeAndTrack(ctx context.Context, raws []model.RawTopic, catMap map[int]string, result *SyncResult, maxBump *time.Time) error {
	topics := normalizeAll(raws, catMap, o.baseURL)
	if err := o.store.StoreTopics(ctx, topics); err != nil {
		return fmt.Errorf("store topics: %w", err)
	}
	result.PagesFetched++
	result.TopicsStored += len(topics)
	trackMaxBump(maxBump, raws)
	return nil
}

// finalizeInitialSync saves the watermark and clears page progress.
func (o *Observer) finalizeInitialSync(ctx context.Context, result *SyncResult, maxBump time.Time) error {
	if !maxBump.IsZero() {
		if err := o.store.SaveWatermark(ctx, maxBump); err != nil {
			return fmt.Errorf("save watermark: %w", err)
		}
		result.NewWatermark = &maxBump
	}
	if err := o.store.ClearLastPage(ctx); err != nil {
		return fmt.Errorf("clear last page: %w", err)
	}
	return nil
}

// trackMaxBump updates maxBump to the latest BumpedAt in the slice.
func trackMaxBump(maxBump *time.Time, raws []model.RawTopic) {
	for i := range raws {
		if b := raws[i].BumpedAt; b != nil && b.After(*maxBump) {
			*maxBump = *b
		}
	}
}

// allAtOrBelowWatermark returns true if every topic's BumpedAt is ≤ wm.
func allAtOrBelowWatermark(raws []model.RawTopic, wm time.Time) bool {
	for i := range raws {
		b := raws[i].BumpedAt
		if b == nil {
			continue
		}
		if b.After(wm) {
			return false
		}
	}
	return true
}
