// Spec: docs/decisions/0006-analytical-storage.md
// Tests: backend/pipeline_test.go, backend/api/contract_test.go, backend/sync_test.go
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/code-community/discourse-observer/backend/model"

	_ "modernc.org/sqlite"
)

// SQLiteStore persists normalized topics to a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at path and
// runs schema migrations.
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

// StoreTopics upserts topics into the database. Existing rows with the
// same ID are replaced, making the operation idempotent.
func (s *SQLiteStore) StoreTopics(ctx context.Context, topics []model.Topic) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO topics
			(id, title, created_at, category_name, tags, reply_count,
			 outcome, first_reply_at, resolved_at, last_activity_at, topic_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for i := range topics {
		tp := &topics[i]
		tags, _ := json.Marshal(tp.Tags)
		_, err := stmt.ExecContext(ctx,
			tp.ID,
			tp.Title,
			tp.CreatedAt.Format(time.RFC3339),
			tp.CategoryName,
			string(tags),
			tp.ReplyCount,
			tp.Outcome,
			formatTime(tp.FirstReplyAt),
			formatTime(tp.ResolvedAt),
			formatTime(tp.LastActivityAt),
			tp.TopicURL,
		)
		if err != nil {
			return fmt.Errorf("insert topic %d: %w", tp.ID, err)
		}
	}
	return tx.Commit()
}

// LoadTopics reads all topics from the database, ordered by ID.
func (s *SQLiteStore) LoadTopics(ctx context.Context) ([]model.Topic, error) {
	return s.QueryTopics(ctx, model.QueryOpts{})
}

// QueryTopics reads topics filtered by the given options.
// Time bounds filter on created_at. Tag filters using json_each on the
// tags JSON array. Returns topics ordered by ID.
func (s *SQLiteStore) QueryTopics(ctx context.Context, opts model.QueryOpts) ([]model.Topic, error) {
	var (
		where []string
		args  []any
	)

	if opts.From != nil {
		where = append(where, "created_at >= ?")
		args = append(args, opts.From.Format(time.RFC3339))
	}
	if opts.To != nil {
		where = append(where, "created_at <= ?")
		args = append(args, opts.To.Format(time.RFC3339))
	}
	if opts.Tag != "" {
		where = append(where, "EXISTS (SELECT 1 FROM json_each(tags) WHERE json_each.value = ?)")
		args = append(args, opts.Tag)
	}

	query := `SELECT id, title, created_at, category_name, tags, reply_count,
	                 outcome, first_reply_at, resolved_at, last_activity_at, topic_url
	          FROM topics`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTopics(rows)
}

func scanTopics(rows *sql.Rows) ([]model.Topic, error) {
	var topics []model.Topic
	for rows.Next() {
		var (
			t                                  model.Topic
			createdAt, tags                    string
			firstReply, resolved, lastActivity sql.NullString
		)
		err := rows.Scan(
			&t.ID, &t.Title, &createdAt, &t.CategoryName, &tags,
			&t.ReplyCount, &t.Outcome,
			&firstReply, &resolved, &lastActivity, &t.TopicURL,
		)
		if err != nil {
			return nil, err
		}

		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		if err := json.Unmarshal([]byte(tags), &t.Tags); err != nil {
			t.Tags = []string{}
		}
		t.FirstReplyAt = parseNullTime(firstReply)
		t.ResolvedAt = parseNullTime(resolved)
		t.LastActivityAt = parseNullTime(lastActivity)

		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
			id               INTEGER PRIMARY KEY,
			title            TEXT    NOT NULL,
			created_at       TEXT    NOT NULL,
			category_name    TEXT    NOT NULL DEFAULT '',
			tags             TEXT    NOT NULL DEFAULT '[]',
			reply_count      INTEGER NOT NULL DEFAULT 0,
			outcome          TEXT    NOT NULL DEFAULT '',
			first_reply_at   TEXT,
			resolved_at      TEXT,
			last_activity_at TEXT,
			topic_url        TEXT    NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS sync_state (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS topic_detail_sync (
			topic_id  INTEGER PRIMARY KEY,
			synced_at TEXT    NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sync_log (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp   TEXT    NOT NULL,
			mode        TEXT    NOT NULL,
			pages       INTEGER NOT NULL,
			topics      INTEGER NOT NULL,
			duration_s  REAL    NOT NULL,
			has_changes INTEGER NOT NULL DEFAULT 1
		)
	`)
	return err
}

func formatTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339)
}

func parseNullTime(ns sql.NullString) *time.Time {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, ns.String)
	if err != nil {
		return nil
	}
	return &t
}
