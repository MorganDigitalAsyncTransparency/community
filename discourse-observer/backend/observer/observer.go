// Spec: specs/observer/observer-behavior.md
// Tests: backend/pipeline_test.go
package observer

import (
	"context"
	"fmt"
	"strconv"

	"github.com/code-community/discourse-observer/backend/model"
)

// FetchClient fetches raw data from a Discourse-compatible API.
type FetchClient interface {
	FetchTopics(ctx context.Context) ([]model.RawTopic, error)
	FetchCategories(ctx context.Context) ([]model.RawCategory, error)
}

// StorageBackend persists normalized topics.
type StorageBackend interface {
	StoreTopics(ctx context.Context, topics []model.Topic) error
	Close() error
}

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

// Run executes one full sync cycle: fetch categories, fetch topics,
// normalize to domain types, and store.
func (o *Observer) Run(ctx context.Context) error {
	cats, err := o.fetch.FetchCategories(ctx)
	if err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}
	catMap := buildCategoryMap(cats)

	raws, err := o.fetch.FetchTopics(ctx)
	if err != nil {
		return fmt.Errorf("fetch topics: %w", err)
	}

	topics := make([]model.Topic, len(raws))
	for i, raw := range raws {
		topics[i] = Normalize(raw, catMap, o.baseURL)
	}

	if err := o.store.StoreTopics(ctx, topics); err != nil {
		return fmt.Errorf("store topics: %w", err)
	}
	return nil
}

// Normalize transforms a raw Discourse topic into a domain Topic.
func Normalize(raw model.RawTopic, categories map[int]string, baseURL string) model.Topic {
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

func deriveOutcome(raw model.RawTopic) string {
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
