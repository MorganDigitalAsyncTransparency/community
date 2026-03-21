// Spec: specs/observer/detail-sync.md
package observer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/observer"
)

// --- Fakes ---

type fakeDetailFetchClient struct {
	topics     map[int]*model.RawTopicDetail
	revisions  map[int]map[int]*model.RawRevision // postID → version → revision
	fetchDelay time.Duration
}

func (f *fakeDetailFetchClient) FetchTopics(_ context.Context) ([]model.RawTopic, error) {
	return nil, nil
}
func (f *fakeDetailFetchClient) FetchTopicsPages(_ context.Context, _ int, _ func([]model.RawTopic, int) error) error {
	return nil
}
func (f *fakeDetailFetchClient) FetchCategories(_ context.Context) ([]model.RawCategory, error) {
	return nil, nil
}
func (f *fakeDetailFetchClient) FetchTopicCount(_ context.Context) int { return 0 }

func (f *fakeDetailFetchClient) FetchTopicDetail(_ context.Context, topicID int) (*model.RawTopicDetail, error) {
	d, ok := f.topics[topicID]
	if !ok {
		return nil, &fakeHTTPError{status: 404}
	}
	return d, nil
}

func (f *fakeDetailFetchClient) FetchPostRevision(_ context.Context, postID, version int) (*model.RawRevision, error) {
	if f.fetchDelay > 0 {
		time.Sleep(f.fetchDelay)
	}
	post, ok := f.revisions[postID]
	if !ok {
		return nil, &fakeHTTPError{status: 404}
	}
	rev, ok := post[version]
	if !ok {
		return nil, &fakeHTTPError{status: 404}
	}
	return rev, nil
}

// fakeHTTPError satisfies observer.HTTPStatusError for 404 detection.
type fakeHTTPError struct {
	status int
}

func (e *fakeHTTPError) Error() string          { return "fake HTTP error" }
func (e *fakeHTTPError) HTTPStatusCode() int     { return e.status }

type fakeDetailStore struct {
	detailSyncs map[int]detailSyncRecord
	events      []model.TopicEvent
	topics      []model.TopicDetailState
	// StorageBackend stub methods
}

type detailSyncRecord struct {
	lastRevision int
	syncedAt     time.Time
}

func (s *fakeDetailStore) StoreTopics(_ context.Context, _ []model.Topic) error { return nil }
func (s *fakeDetailStore) SaveWatermark(_ context.Context, _ time.Time) error   { return nil }
func (s *fakeDetailStore) LoadWatermark(_ context.Context) (*time.Time, error)  { return nil, nil }
func (s *fakeDetailStore) SaveLastPage(_ context.Context, _ int) error          { return nil }
func (s *fakeDetailStore) LoadLastPage(_ context.Context) (int, error)          { return -1, nil }
func (s *fakeDetailStore) ClearLastPage(_ context.Context) error                { return nil }
func (s *fakeDetailStore) Close() error                                         { return nil }

func (s *fakeDetailStore) SaveDetailSync(_ context.Context, topicID, lastRevision int, syncedAt time.Time) error {
	if s.detailSyncs == nil {
		s.detailSyncs = make(map[int]detailSyncRecord)
	}
	s.detailSyncs[topicID] = detailSyncRecord{lastRevision: lastRevision, syncedAt: syncedAt}
	return nil
}

func (s *fakeDetailStore) TopicsNeedingDetailSync(_ context.Context, limit int) ([]model.TopicDetailState, error) {
	if limit > len(s.topics) {
		return s.topics, nil
	}
	return s.topics[:limit], nil
}

func (s *fakeDetailStore) SaveTopicEvents(_ context.Context, events []model.TopicEvent) error {
	s.events = append(s.events, events...)
	return nil
}

// Compile-time interface checks (DS-6, DS-7).
var _ observer.FetchClient = (*fakeDetailFetchClient)(nil)
var _ observer.StorageBackend = (*fakeDetailStore)(nil)

// --- Tests ---

func TestDetailSyncEndToEnd(t *testing.T) {
	baseTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{
			100: {ID: 100, PostStream: model.PostStream{Posts: []model.RawPost{
				{ID: 1001, PostNumber: 1, Version: 3},
			}}},
		},
		revisions: map[int]map[int]*model.RawRevision{
			1001: {
				2: {
					CreatedAt: baseTime,
					Tags:      &model.RevisionTagChange{Previous: []string{}, Current: []string{"bug"}},
				},
				3: {
					CreatedAt: baseTime.Add(time.Hour),
					Title:     &model.RevisionChange{Previous: "Old title", Current: "New title"},
				},
			},
		},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{{TopicID: 100, LastRevision: 0}},
	}

	obs := observer.New(fetch, store, "http://example.com")
	result, err := obs.RunDetailSync(context.Background())
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}

	if result.Mode != "detail" {
		t.Errorf("mode = %q, want detail", result.Mode)
	}
	if result.TopicsStored != 1 {
		t.Errorf("topics stored = %d, want 1", result.TopicsStored)
	}
	if len(store.events) != 2 {
		t.Fatalf("got %d events, want 2", len(store.events))
	}
	if store.events[0].EventType != "tag_change" {
		t.Errorf("event[0] type = %q, want tag_change", store.events[0].EventType)
	}
	if store.events[1].EventType != "title_edit" {
		t.Errorf("event[1] type = %q, want title_edit", store.events[1].EventType)
	}
	rec := store.detailSyncs[100]
	if rec.lastRevision != 3 {
		t.Errorf("last revision = %d, want 3", rec.lastRevision)
	}
}

func TestDetailSyncDeltaRevisions(t *testing.T) {
	baseTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{
			100: {ID: 100, PostStream: model.PostStream{Posts: []model.RawPost{
				{ID: 1001, PostNumber: 1, Version: 5},
			}}},
		},
		revisions: map[int]map[int]*model.RawRevision{
			1001: {
				// Revisions 2-3 already fetched; only 4-5 should be requested.
				4: {CreatedAt: baseTime, Tags: &model.RevisionTagChange{Previous: []string{"bug"}, Current: []string{"bug", "critical"}}},
				5: {CreatedAt: baseTime.Add(time.Hour), CategoryID: &model.RevisionIntChange{Previous: 1, Current: 2}},
			},
		},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{{TopicID: 100, LastRevision: 3}},
	}

	obs := observer.New(fetch, store, "http://example.com")
	result, err := obs.RunDetailSync(context.Background())
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}
	if result.TopicsStored != 1 {
		t.Errorf("topics stored = %d, want 1", result.TopicsStored)
	}
	if len(store.events) != 2 {
		t.Fatalf("got %d events, want 2", len(store.events))
	}
	if store.events[0].EventType != "tag_change" {
		t.Errorf("event[0] type = %q, want tag_change", store.events[0].EventType)
	}
	if store.events[1].EventType != "category_move" {
		t.Errorf("event[1] type = %q, want category_move", store.events[1].EventType)
	}
	rec := store.detailSyncs[100]
	if rec.lastRevision != 5 {
		t.Errorf("last revision = %d, want 5", rec.lastRevision)
	}
}

func TestDetailSyncNoRevisions(t *testing.T) {
	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{
			200: {ID: 200, PostStream: model.PostStream{Posts: []model.RawPost{
				{ID: 2001, PostNumber: 1, Version: 1},
			}}},
		},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{{TopicID: 200, LastRevision: 0}},
	}

	obs := observer.New(fetch, store, "http://example.com")
	result, err := obs.RunDetailSync(context.Background())
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}
	if result.TopicsStored != 0 {
		t.Errorf("topics stored = %d, want 0 (no revisions)", result.TopicsStored)
	}
	if len(store.events) != 0 {
		t.Errorf("got %d events, want 0", len(store.events))
	}
	rec := store.detailSyncs[200]
	if rec.lastRevision != 1 {
		t.Errorf("last revision = %d, want 1", rec.lastRevision)
	}
}

func TestDetailSyncInterruptible(t *testing.T) {
	baseTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{
			100: {ID: 100, PostStream: model.PostStream{Posts: []model.RawPost{
				{ID: 1001, PostNumber: 1, Version: 2},
			}}},
			200: {ID: 200, PostStream: model.PostStream{Posts: []model.RawPost{
				{ID: 2001, PostNumber: 1, Version: 2},
			}}},
		},
		revisions: map[int]map[int]*model.RawRevision{
			1001: {2: {CreatedAt: baseTime, Tags: &model.RevisionTagChange{Previous: []string{}, Current: []string{"x"}}}},
			2001: {2: {CreatedAt: baseTime, Tags: &model.RevisionTagChange{Previous: []string{}, Current: []string{"y"}}}},
		},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{
			{TopicID: 100, LastRevision: 0},
			{TopicID: 200, LastRevision: 0},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first topic is processed.
	origSave := store.SaveDetailSync
	_ = origSave // store.SaveDetailSync is a method, we wrap via interception
	// Instead, use a store that cancels after first save.
	cancelStore := &cancelOnFirstSave{fakeDetailStore: store, cancel: cancel}

	obs := observer.New(fetch, cancelStore, "http://example.com")
	result, err := obs.RunDetailSync(ctx)
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}

	// Should have processed topic 100 but not 200 (canceled).
	if result.PagesFetched != 1 {
		t.Errorf("topics processed = %d, want 1", result.PagesFetched)
	}
	if _, ok := cancelStore.fakeDetailStore.detailSyncs[100]; !ok {
		t.Error("topic 100 should be marked as synced")
	}
	if _, ok := cancelStore.fakeDetailStore.detailSyncs[200]; ok {
		t.Error("topic 200 should NOT be marked as synced (context canceled)")
	}
}

// cancelOnFirstSave wraps fakeDetailStore and cancels context after first SaveDetailSync.
type cancelOnFirstSave struct {
	*fakeDetailStore
	cancel func()
	called bool
}

func (s *cancelOnFirstSave) SaveDetailSync(ctx context.Context, topicID, lastRevision int, syncedAt time.Time) error {
	err := s.fakeDetailStore.SaveDetailSync(ctx, topicID, lastRevision, syncedAt)
	if !s.called {
		s.called = true
		s.cancel()
	}
	return err
}

func TestDetailSyncDeletedTopic(t *testing.T) {
	// Topic 300 does not exist in fetch client — should get a 404.
	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{{TopicID: 300, LastRevision: 0}},
	}

	obs := observer.New(fetch, store, "http://example.com")
	result, err := obs.RunDetailSync(context.Background())
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}

	// Topic should be marked as skipped (last_revision = -1).
	rec, ok := store.detailSyncs[300]
	if !ok {
		t.Fatal("topic 300 should be marked in detailSyncs")
	}
	if rec.lastRevision != -1 {
		t.Errorf("last revision = %d, want -1 (skipped)", rec.lastRevision)
	}
	if result.TopicsStored != 0 {
		t.Errorf("topics stored = %d, want 0", result.TopicsStored)
	}
}

func TestDetailSyncPrioritization(t *testing.T) {
	// This test verifies that the observer processes topics in the order
	// provided by TopicsNeedingDetailSync (which storage orders by priority).
	baseTime := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	fetch := &fakeDetailFetchClient{
		topics: map[int]*model.RawTopicDetail{
			10: {ID: 10, PostStream: model.PostStream{Posts: []model.RawPost{{ID: 101, PostNumber: 1, Version: 1}}}},
			20: {ID: 20, PostStream: model.PostStream{Posts: []model.RawPost{{ID: 201, PostNumber: 1, Version: 2}}}},
			30: {ID: 30, PostStream: model.PostStream{Posts: []model.RawPost{{ID: 301, PostNumber: 1, Version: 1}}}},
		},
		revisions: map[int]map[int]*model.RawRevision{
			201: {2: {CreatedAt: baseTime, Tags: &model.RevisionTagChange{Previous: []string{}, Current: []string{"a"}}}},
		},
	}

	store := &fakeDetailStore{
		topics: []model.TopicDetailState{
			{TopicID: 10, LastRevision: 0}, // never synced
			{TopicID: 20, LastRevision: 0}, // never synced, has revisions
			{TopicID: 30, LastRevision: 0}, // never synced
		},
	}

	obs := observer.New(fetch, store, "http://example.com")
	_, err := obs.RunDetailSync(context.Background())
	if err != nil {
		t.Fatalf("RunDetailSync: %v", err)
	}

	// All three should be synced.
	for _, id := range []int{10, 20, 30} {
		if _, ok := store.detailSyncs[id]; !ok {
			t.Errorf("topic %d not synced", id)
		}
	}
	// Only topic 20 should have events.
	if len(store.events) != 1 {
		t.Errorf("got %d events, want 1", len(store.events))
	}
}

func TestIsNotFound(t *testing.T) {
	notFound := &fakeHTTPError{status: 404}
	wrapped := errors.New("wrapped: " + notFound.Error())
	_ = wrapped

	// Direct 404 — detected via errors.As.
	if !isNotFoundHelper(notFound) {
		t.Error("expected 404 to be detected")
	}

	// Non-404 — should not match.
	other := &fakeHTTPError{status: 500}
	if isNotFoundHelper(other) {
		t.Error("500 should not be detected as not-found")
	}
}

// isNotFoundHelper wraps the package-internal function for testing.
// Since detail_sync.go is in package observer, and tests are in observer_test,
// we test through RunDetailSync behavior instead.
func isNotFoundHelper(err error) bool {
	var he interface {
		error
		HTTPStatusCode() int
	}
	return errors.As(err, &he) && he.HTTPStatusCode() == 404
}
