// Spec: specs/observer/observer-behavior.md
// Tests: backend/pipeline_test.go
package observer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

	cats, err := o.fetch.FetchCategories(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("fetch categories: %w", err)
	}
	catMap := buildCategoryMap(cats)

	lastPage, err := o.store.LoadLastPage(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("load last page: %w", err)
	}
	startPage := 0
	if lastPage >= 0 {
		startPage = lastPage + 1
	}

	var maxBump time.Time
	var result SyncResult
	result.Mode = "initial"

	err = o.fetch.FetchTopicsPages(ctx, startPage, func(raws []model.RawTopic, page int) error {
		topics := normalizeAll(raws, catMap, o.baseURL)
		if err := o.store.StoreTopics(ctx, topics); err != nil {
			return fmt.Errorf("store page %d: %w", page, err)
		}
		if err := o.store.SaveLastPage(ctx, page); err != nil {
			return fmt.Errorf("save last page %d: %w", page, err)
		}
		result.PagesFetched++
		result.TopicsStored += len(topics)
		trackMaxBump(&maxBump, raws)
		return nil
	})
	if err != nil {
		return result, err
	}

	if !maxBump.IsZero() {
		if err := o.store.SaveWatermark(ctx, maxBump); err != nil {
			return result, fmt.Errorf("save watermark: %w", err)
		}
		result.NewWatermark = &maxBump
	}
	if err := o.store.ClearLastPage(ctx); err != nil {
		return result, fmt.Errorf("clear last page: %w", err)
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

	cats, err := o.fetch.FetchCategories(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("fetch categories: %w", err)
	}
	catMap := buildCategoryMap(cats)

	var maxBump time.Time
	var result SyncResult
	result.Mode = "delta"

	err = o.fetch.FetchTopicsPages(ctx, 0, func(raws []model.RawTopic, page int) error {
		topics := normalizeAll(raws, catMap, o.baseURL)
		if err := o.store.StoreTopics(ctx, topics); err != nil {
			return fmt.Errorf("store page %d: %w", page, err)
		}
		result.PagesFetched++
		result.TopicsStored += len(topics)
		trackMaxBump(&maxBump, raws)

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

// normalizeAll converts a slice of raw topics to domain topics.
func normalizeAll(raws []model.RawTopic, catMap map[int]string, baseURL string) []model.Topic {
	topics := make([]model.Topic, len(raws))
	for i := range raws {
		topics[i] = Normalize(&raws[i], catMap, baseURL)
	}
	return topics
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

// Normalize transforms a raw Discourse topic into a domain Topic.
func Normalize(raw *model.RawTopic, categories map[int]string, baseURL string) model.Topic {
	outcome := deriveOutcome(raw)

	t := model.Topic{
		ID:           raw.ID,
		Title:        raw.Title,
		CreatedAt:    raw.CreatedAt,
		Tags:         raw.Tags,
		CategoryName: categories[raw.CategoryID],
		ReplyCount:   raw.ReplyCount,
		Outcome:      outcome,
		FirstReplyAt: raw.FirstReplyAt,
		TopicURL:     baseURL + "/t/" + strconv.Itoa(raw.ID),
	}

	if t.Tags == nil {
		t.Tags = []string{}
	}

	switch outcome {
	case "solved":
		t.ResolvedAt = raw.AcceptedAnswerAt
	case "self-closed":
		t.ResolvedAt = raw.ClosedAt
	}

	// Use LastPostedAt as last activity; fall back to BumpedAt.
	if raw.LastPostedAt != nil {
		t.LastActivityAt = raw.LastPostedAt
	} else if raw.BumpedAt != nil {
		t.LastActivityAt = raw.BumpedAt
	}

	return t
}

func deriveOutcome(raw *model.RawTopic) string {
	if raw.HasAcceptedAnswer {
		return "solved"
	}
	if raw.Closed {
		return "self-closed"
	}
	return ""
}

func buildCategoryMap(cats []model.RawCategory) map[int]string {
	m := make(map[int]string, len(cats))
	for _, c := range cats {
		m[c.ID] = c.Name
	}
	return m
}
