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
func (s *SQLiteStore) SaveSyncLogEntry(ctx context.Context, e model.SyncLogEntry) error {
	hasChanges := 0
	if e.HasChanges {
		hasChanges = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sync_log (timestamp, mode, pages, topics, duration_s, has_changes) VALUES (?, ?, ?, ?, ?, ?)`,
		e.Timestamp.Format(time.RFC3339), e.Mode, e.Pages, e.Topics, e.Duration.Seconds(), hasChanges)
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
		`SELECT timestamp, mode, pages, topics, duration_s, has_changes FROM sync_log ORDER BY timestamp DESC`)
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
		if err := rows.Scan(&raw, &e.Mode, &e.Pages, &e.Topics, &durS, &hc); err != nil {
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

// SaveDetailSync records when a topic was last detail-synced.
func (s *SQLiteStore) SaveDetailSync(ctx context.Context, topicID int, syncedAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO topic_detail_sync (topic_id, synced_at) VALUES (?, ?)`,
		topicID, syncedAt.Format(time.RFC3339))
	return err
}

// TopicsNeedingDetailSync returns topic IDs that need detail enrichment,
// ordered by priority: never synced first, then stale (oldest synced_at),
// with ties broken by topic ID ascending. Only topics present in the
// topics table are considered. Limit controls how many IDs to return.
func (s *SQLiteStore) TopicsNeedingDetailSync(ctx context.Context, limit int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id
		FROM topics t
		LEFT JOIN topic_detail_sync d ON t.id = d.topic_id
		ORDER BY d.synced_at IS NOT NULL, d.synced_at ASC, t.id ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
