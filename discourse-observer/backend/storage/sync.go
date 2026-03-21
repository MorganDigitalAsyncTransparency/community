// Spec: specs/observer/sync-metadata.md
// Tests: backend/storage/sqlite_test.go
package storage

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// SaveWatermark persists the high-water mark timestamp.
func (s *SQLiteStore) SaveWatermark(ctx context.Context, t time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sync_state (key, value) VALUES ('watermark', ?)`,
		t.Format(time.RFC3339))
	return err
}

// LoadWatermark returns the stored watermark, or nil if none exists.
func (s *SQLiteStore) LoadWatermark(ctx context.Context) (*time.Time, error) {
	var raw string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM sync_state WHERE key = 'watermark'`).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// SaveLastPage records the last completed page number (for initial sync resume).
func (s *SQLiteStore) SaveLastPage(ctx context.Context, page int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sync_state (key, value) VALUES ('last_completed_page', ?)`,
		strconv.Itoa(page))
	return err
}

// LoadLastPage returns the stored page number, or -1 if none exists.
func (s *SQLiteStore) LoadLastPage(ctx context.Context) (int, error) {
	var raw int
	err := s.db.QueryRowContext(ctx,
		`SELECT CAST(value AS INTEGER) FROM sync_state WHERE key = 'last_completed_page'`).Scan(&raw)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return 0, err
	}
	return raw, nil
}

// ClearLastPage removes the stored page number (called when initial sync completes).
func (s *SQLiteStore) ClearLastPage(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM sync_state WHERE key = 'last_completed_page'`)
	return err
}

const maxLogPerType = 20

// SaveSyncLogEntry appends a sync log entry, keeping at most 20 per mode.
// No-change entries (has_changes=0) are deduplicated: only the most recent
// no-change entry per mode is kept.
func (s *SQLiteStore) SaveSyncLogEntry(ctx context.Context, e *model.SyncLogEntry) error {
	hasChanges := 0
	if e.HasChanges {
		hasChanges = 1
	}

	// Before inserting a no-change entry, remove all previous no-change entries for this mode.
	// Error entries are never deduplicated — each failure is kept individually.
	if !e.HasChanges && e.Error == "" {
		if _, err := s.db.ExecContext(ctx,
			`DELETE FROM sync_log WHERE mode = ? AND has_changes = 0 AND error = ''`, e.Mode); err != nil {
			return err
		}
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sync_log (timestamp, mode, pages, topics, duration_s, has_changes, error) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.Timestamp.Format(time.RFC3339), e.Mode, e.Pages, e.Topics, e.Duration.Seconds(), hasChanges, e.Error)
	if err != nil {
		return err
	}
	// Trim to maxLogPerType per mode: delete oldest entries beyond the limit.
	_, err = s.db.ExecContext(ctx, `
		DELETE FROM sync_log WHERE id IN (
			SELECT id FROM sync_log WHERE mode = ?
			ORDER BY timestamp DESC
			LIMIT -1 OFFSET ?
		)`, e.Mode, maxLogPerType)
	return err
}

// LoadSyncLog returns all stored sync log entries, newest first.
func (s *SQLiteStore) LoadSyncLog(ctx context.Context) ([]model.SyncLogEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT timestamp, mode, pages, topics, duration_s, has_changes, error FROM sync_log ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []model.SyncLogEntry
	for rows.Next() {
		var raw string
		var e model.SyncLogEntry
		var durS float64
		var hc int
		if err := rows.Scan(&raw, &e.Mode, &e.Pages, &e.Topics, &durS, &hc, &e.Error); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, err
		}
		e.Timestamp = t
		e.Duration = time.Duration(durS * float64(time.Second))
		e.HasChanges = hc != 0
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SaveDetailSync records when a topic was last detail-synced and the
// highest revision version fetched.
func (s *SQLiteStore) SaveDetailSync(ctx context.Context, topicID, lastRevision int, syncedAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO topic_detail_sync (topic_id, synced_at, last_revision) VALUES (?, ?, ?)`,
		topicID, syncedAt.Format(time.RFC3339), lastRevision)
	return err
}

// TopicsNeedingDetailSync returns topics that need detail enrichment,
// ordered by priority: never synced first, then stale (oldest synced_at),
// with ties broken by topic ID ascending. Returns both topic ID and last
// fetched revision version. Limit controls how many to return.
func (s *SQLiteStore) TopicsNeedingDetailSync(ctx context.Context, limit int) ([]model.TopicDetailState, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, COALESCE(d.last_revision, 0)
		FROM topics t
		LEFT JOIN topic_detail_sync d ON t.id = d.topic_id
		ORDER BY d.synced_at IS NOT NULL, d.synced_at ASC, t.id ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var states []model.TopicDetailState
	for rows.Next() {
		var s model.TopicDetailState
		if err := rows.Scan(&s.TopicID, &s.LastRevision); err != nil {
			return nil, err
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

// SaveTopicEvents stores extracted revision events for a topic.
// Existing events for the same topic are not duplicated — events are
// matched by topic_id + event_type + happened_at.
func (s *SQLiteStore) SaveTopicEvents(ctx context.Context, events []model.TopicEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO topic_events (topic_id, event_type, happened_at, detail)
		SELECT ?, ?, ?, ?
		WHERE NOT EXISTS (
			SELECT 1 FROM topic_events
			WHERE topic_id = ? AND event_type = ? AND happened_at = ?
		)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for i := range events {
		e := &events[i]
		ts := e.HappenedAt.Format(time.RFC3339)
		_, err := stmt.ExecContext(ctx,
			e.TopicID, e.EventType, ts, e.Detail,
			e.TopicID, e.EventType, ts)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadTopicEvents returns all stored events for a topic, ordered by time.
func (s *SQLiteStore) LoadTopicEvents(ctx context.Context, topicID int) ([]model.TopicEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, topic_id, event_type, happened_at, detail
		FROM topic_events WHERE topic_id = ?
		ORDER BY happened_at ASC`, topicID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []model.TopicEvent
	for rows.Next() {
		var e model.TopicEvent
		var ts string
		if err := rows.Scan(&e.ID, &e.TopicID, &e.EventType, &ts, &e.Detail); err != nil {
			return nil, err
		}
		e.HappenedAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
