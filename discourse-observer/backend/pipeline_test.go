// Spec: specs/observer/observer-behavior.md
// Tests the full pipeline: mock server → fetch → observe → store → SQLite.
package main_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/storage"
)

func TestPipelineEndToEnd(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	client := discourse.NewClient(srv.URL, "", "")
	obs := observer.New(client, store, srv.URL)

	ctx := context.Background()
	if err := obs.Run(ctx); err != nil {
		t.Fatalf("pipeline run: %v", err)
	}

	topics, err := store.LoadTopics(ctx)
	if err != nil {
		t.Fatalf("load topics: %v", err)
	}

	expected := mock.Topics()
	if len(topics) != len(expected) {
		t.Fatalf("got %d topics, want %d", len(topics), len(expected))
	}

	// Build lookup by ID for field-level checks.
	byID := map[int]int{}
	for i, tp := range topics {
		byID[tp.ID] = i
	}

	for _, want := range expected {
		idx, ok := byID[want.ID]
		if !ok {
			t.Errorf("topic %d missing from stored results", want.ID)
			continue
		}
		got := topics[idx]

		if got.Title != want.Title {
			t.Errorf("topic %d title = %q, want %q", want.ID, got.Title, want.Title)
		}
		if got.CategoryName != want.CategoryName {
			t.Errorf("topic %d category = %q, want %q", want.ID, got.CategoryName, want.CategoryName)
		}
		if got.Outcome != want.Outcome {
			t.Errorf("topic %d outcome = %q, want %q", want.ID, got.Outcome, want.Outcome)
		}
		if got.ReplyCount != want.ReplyCount {
			t.Errorf("topic %d replyCount = %d, want %d", want.ID, got.ReplyCount, want.ReplyCount)
		}
		if len(got.Tags) != len(want.Tags) {
			t.Errorf("topic %d tags = %v, want %v", want.ID, got.Tags, want.Tags)
		}
		if !got.CreatedAt.Equal(want.CreatedAt) {
			t.Errorf("topic %d createdAt = %v, want %v", want.ID, got.CreatedAt, want.CreatedAt)
		}
		if (got.FirstReplyAt == nil) != (want.FirstReplyAt == nil) {
			t.Errorf("topic %d firstReplyAt nil mismatch: got nil=%v, want nil=%v",
				want.ID, got.FirstReplyAt == nil, want.FirstReplyAt == nil)
		} else if got.FirstReplyAt != nil && !got.FirstReplyAt.Equal(*want.FirstReplyAt) {
			t.Errorf("topic %d firstReplyAt = %v, want %v", want.ID, *got.FirstReplyAt, *want.FirstReplyAt)
		}
		if (got.ResolvedAt == nil) != (want.ResolvedAt == nil) {
			t.Errorf("topic %d resolvedAt nil mismatch: got nil=%v, want nil=%v",
				want.ID, got.ResolvedAt == nil, want.ResolvedAt == nil)
		} else if got.ResolvedAt != nil && !got.ResolvedAt.Equal(*want.ResolvedAt) {
			t.Errorf("topic %d resolvedAt = %v, want %v", want.ID, *got.ResolvedAt, *want.ResolvedAt)
		}
	}
}

func TestPipelineIdempotent(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	client := discourse.NewClient(srv.URL, "", "")
	obs := observer.New(client, store, srv.URL)
	ctx := context.Background()

	// Run pipeline twice.
	for i := range 2 {
		if err := obs.Run(ctx); err != nil {
			t.Fatalf("run %d: %v", i+1, err)
		}
	}

	topics, err := store.LoadTopics(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(topics) != len(mock.Topics()) {
		t.Errorf("after 2 runs: got %d topics, want %d", len(topics), len(mock.Topics()))
	}
}

func TestPipelineTopicURLs(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	client := discourse.NewClient(srv.URL, "", "")
	obs := observer.New(client, store, srv.URL)

	if err := obs.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}

	topics, err := store.LoadTopics(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, tp := range topics {
		if tp.TopicURL == "" {
			t.Errorf("topic %d has empty URL", tp.ID)
		}
	}
}
