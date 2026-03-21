// Spec: specs/observer/detail-sync.md (DS-11, DS-12, DS-13, DS-14)
// Tests: backend/observer/detail_sync_test.go
package observer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

const detailSyncBatchSize = 50

// RunDetailSync enriches topics with revision history.
// It fetches topic detail and post revisions for topics that need
// enrichment, extracting tag change, category move, and title edit events.
// The method is interruptible — it checks ctx between topics.
func (o *Observer) RunDetailSync(ctx context.Context) (SyncResult, error) {
	start := time.Now()
	result := SyncResult{Mode: "detail"}

	topics, err := o.store.TopicsNeedingDetailSync(ctx, detailSyncBatchSize)
	if err != nil {
		return result, fmt.Errorf("load topics needing detail sync: %w", err)
	}

	for _, ts := range topics {
		if err := ctx.Err(); err != nil {
			break
		}
		enriched, err := o.enrichTopic(ctx, ts)
		if err != nil {
			if isNotFound(err) {
				if markErr := o.markSkipped(ctx, ts.TopicID); markErr != nil {
					return result, fmt.Errorf("mark skipped topic %d: %w", ts.TopicID, markErr)
				}
				continue
			}
			return result, fmt.Errorf("enrich topic %d: %w", ts.TopicID, err)
		}
		if enriched {
			result.TopicsStored++
		}
		result.PagesFetched++ // reuse as "topics processed" counter
	}

	result.Duration = time.Since(start)
	return result, nil
}

// enrichTopic fetches detail and revisions for a single topic.
// Returns true if the topic was enriched (new revisions found or marked).
func (o *Observer) enrichTopic(ctx context.Context, ts model.TopicDetailState) (bool, error) {
	detail, err := o.fetch.FetchTopicDetail(ctx, ts.TopicID)
	if err != nil {
		return false, err
	}

	if len(detail.PostStream.Posts) == 0 {
		return false, o.store.SaveDetailSync(ctx, ts.TopicID, 1, time.Now())
	}

	firstPost := detail.PostStream.Posts[0]
	currentVersion := firstPost.Version

	// No revisions exist — mark as synced at version 1.
	if currentVersion <= 1 {
		return false, o.store.SaveDetailSync(ctx, ts.TopicID, 1, time.Now())
	}

	// Already up to date.
	if currentVersion <= ts.LastRevision {
		return false, o.store.SaveDetailSync(ctx, ts.TopicID, currentVersion, time.Now())
	}

	startVersion := maxInt(ts.LastRevision+1, 2)
	events, err := o.fetchRevisions(ctx, ts.TopicID, firstPost.ID, startVersion, currentVersion)
	if err != nil {
		return false, err
	}

	if len(events) > 0 {
		if err := o.store.SaveTopicEvents(ctx, events); err != nil {
			return false, fmt.Errorf("save events: %w", err)
		}
	}

	if err := o.store.SaveDetailSync(ctx, ts.TopicID, currentVersion, time.Now()); err != nil {
		return false, fmt.Errorf("save detail sync: %w", err)
	}

	return len(events) > 0, nil
}

// fetchRevisions fetches revision data for versions [start, end] and
// extracts topic events.
func (o *Observer) fetchRevisions(ctx context.Context, topicID, postID, start, end int) ([]model.TopicEvent, error) {
	var events []model.TopicEvent

	for v := start; v <= end; v++ {
		if err := ctx.Err(); err != nil {
			return events, err
		}

		rev, err := o.fetch.FetchPostRevision(ctx, postID, v)
		if err != nil {
			if isNotFound(err) {
				continue // revision may have been removed
			}
			return events, fmt.Errorf("revision %d: %w", v, err)
		}

		events = append(events, extractEvents(topicID, rev)...)
	}

	return events, nil
}

// extractEvents pulls tag change, category move, and title edit events
// from a single revision.
func extractEvents(topicID int, rev *model.RawRevision) []model.TopicEvent {
	var events []model.TopicEvent

	if rev.Tags != nil {
		detail, _ := json.Marshal(rev.Tags)
		events = append(events, model.TopicEvent{
			TopicID:    topicID,
			EventType:  "tag_change",
			HappenedAt: rev.CreatedAt,
			Detail:     string(detail),
		})
	}

	if rev.CategoryID != nil {
		detail, _ := json.Marshal(rev.CategoryID)
		events = append(events, model.TopicEvent{
			TopicID:    topicID,
			EventType:  "category_move",
			HappenedAt: rev.CreatedAt,
			Detail:     string(detail),
		})
	}

	if rev.Title != nil {
		detail, _ := json.Marshal(rev.Title)
		events = append(events, model.TopicEvent{
			TopicID:    topicID,
			EventType:  "title_edit",
			HappenedAt: rev.CreatedAt,
			Detail:     string(detail),
		})
	}

	return events
}

// markSkipped marks a deleted topic so it is not re-selected.
// Uses last_revision = -1 as the "skip" sentinel.
func (o *Observer) markSkipped(ctx context.Context, topicID int) error {
	return o.store.SaveDetailSync(ctx, topicID, -1, time.Now())
}

// HTTPStatusError is satisfied by any error that reports an HTTP status code.
// The discourse package's HTTPError has a StatusCode field; this interface
// allows the observer to detect 404s without importing discourse.
type HTTPStatusError interface {
	error
	HTTPStatusCode() int
}

// isNotFound checks if an error chain contains a 404 HTTP error.
func isNotFound(err error) bool {
	var he HTTPStatusError
	return errors.As(err, &he) && he.HTTPStatusCode() == 404
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
