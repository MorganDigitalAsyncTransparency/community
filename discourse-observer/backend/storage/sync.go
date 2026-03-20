// Spec: specs/observer/sync-metadata.md
// Tests: backend/storage/sqlite_test.go
package storage

import (
	"context"
	"database/sql"
	"time"
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
		`INSERT OR REPLACE INTO sync_state (key, value) VALUES ('last_completed_page', CAST(? AS TEXT))`,
		page)
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

// SaveDetailSync records when a topic was last detail-synced.
func (s *SQLiteStore) SaveDetailSync(ctx context.Context, topicID int, syncedAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO topic_detail_sync (topic_id, synced_at) VALUES (?, ?)`,
		topicID, syncedAt.Format(time.RFC3339))
	return err
}

// TopicsNeedingDetailSync returns topic IDs that need detail enrichment,
// ordered by priority: never synced first, then stale (oldest synced_at).
// Limit controls how many IDs to return.
func (s *SQLiteStore) TopicsNeedingDetailSync(ctx context.Context, limit int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id
		FROM topics t
		LEFT JOIN topic_detail_sync d ON t.id = d.topic_id
		ORDER BY d.synced_at IS NOT NULL, d.synced_at ASC
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
