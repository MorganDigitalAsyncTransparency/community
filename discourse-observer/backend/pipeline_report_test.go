// Spec: specs/observer/observer-behavior.md
// Generates docs/test-reports/pipeline-report.md with concrete pipeline output.
package main_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/model"
	"github.com/code-community/discourse-observer/backend/observer"
	"github.com/code-community/discourse-observer/backend/storage"
)

func TestPipelineReport(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	dbPath := filepath.Join(t.TempDir(), "report.db")
	store, err := storage.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer func() { _ = store.Close() }()

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

	// Run pipeline again for idempotency check.
	if err := obs.Run(ctx); err != nil {
		t.Fatalf("second run: %v", err)
	}
	topicsAfter, err := store.LoadTopics(ctx)
	if err != nil {
		t.Fatalf("load after second run: %v", err)
	}

	expected := mock.Topics()
	report := buildReport(topics, topicsAfter, expected, srv.URL)

	reportDir := filepath.Join("..", "docs", "test-reports")
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		t.Fatalf("create report dir: %v", err)
	}
	reportPath := filepath.Join(reportDir, "pipeline-report.md")
	if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	t.Logf("Report written to %s", reportPath)
}

func buildReport(topics, topicsAfter, expected []model.Topic, baseURL string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Pipeline Test Report\n\n")
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "Source: `go test ./backend/... -run TestPipelineReport -v`\n\n")

	fmt.Fprintf(&b, "## Summary\n\n")
	fmt.Fprintf(&b, "| Metric | Value |\n")
	fmt.Fprintf(&b, "|--------|-------|\n")
	fmt.Fprintf(&b, "| Mock server base URL | `%s` |\n", baseURL)
	fmt.Fprintf(&b, "| Topics in mock dataset | %d |\n", len(expected))
	fmt.Fprintf(&b, "| Topics stored in SQLite | %d |\n", len(topics))
	fmt.Fprintf(&b, "| Topics after second run (idempotency) | %d |\n", len(topicsAfter))
	fmt.Fprintf(&b, "| Match | %v |\n\n", len(topics) == len(expected))

	writeOutcomeCounts(&b, topics)
	writeCategoryCounts(&b, topics)
	writeNullStats(&b, topics)
	writeFieldMismatches(&b, topics, expected)
	writeSampleTopics(&b, topics)

	return b.String()
}

func writeOutcomeCounts(b *strings.Builder, topics []model.Topic) {
	counts := map[string]int{}
	for i := range topics {
		tp := &topics[i]
		switch {
		case tp.Outcome == "solved":
			counts["solved"]++
		case tp.Outcome == "self-closed":
			counts["self-closed"]++
		case len(tp.Tags) == 0:
			counts["untagged"]++
		case tp.ReplyCount == 0:
			counts["unreplied"]++
		default:
			counts["replied-open"]++
		}
	}

	fmt.Fprintf(b, "## Topics by status\n\n")
	fmt.Fprintf(b, "| Status | Count |\n")
	fmt.Fprintf(b, "|--------|-------|\n")
	for _, key := range []string{"solved", "self-closed", "unreplied", "untagged", "replied-open"} {
		fmt.Fprintf(b, "| %s | %d |\n", key, counts[key])
	}
	fmt.Fprintf(b, "\n")
}

func writeCategoryCounts(b *strings.Builder, topics []model.Topic) {
	counts := map[string]int{}
	for i := range topics {
		counts[topics[i].CategoryName]++
	}

	fmt.Fprintf(b, "## Topics by category\n\n")
	fmt.Fprintf(b, "| Category | Count |\n")
	fmt.Fprintf(b, "|----------|-------|\n")
	for cat, n := range counts {
		if cat == "" {
			cat = "(empty)"
		}
		fmt.Fprintf(b, "| %s | %d |\n", cat, n)
	}
	fmt.Fprintf(b, "\n")
}

func writeNullStats(b *strings.Builder, topics []model.Topic) {
	var nullFirst, nullResolved, nullLastActivity int
	for i := range topics {
		if topics[i].FirstReplyAt == nil {
			nullFirst++
		}
		if topics[i].ResolvedAt == nil {
			nullResolved++
		}
		if topics[i].LastActivityAt == nil {
			nullLastActivity++
		}
	}

	fmt.Fprintf(b, "## Null timestamp counts\n\n")
	fmt.Fprintf(b, "| Column | Null | Non-null |\n")
	fmt.Fprintf(b, "|--------|------|----------|\n")
	fmt.Fprintf(b, "| first_reply_at | %d | %d |\n", nullFirst, len(topics)-nullFirst)
	fmt.Fprintf(b, "| resolved_at | %d | %d |\n", nullResolved, len(topics)-nullResolved)
	fmt.Fprintf(b, "| last_activity_at | %d | %d |\n", nullLastActivity, len(topics)-nullLastActivity)
	fmt.Fprintf(b, "\n")
}

func writeFieldMismatches(b *strings.Builder, stored, expected []model.Topic) {
	byID := map[int]*model.Topic{}
	for i := range stored {
		byID[stored[i].ID] = &stored[i]
	}

	var mismatches []string
	for i := range expected {
		want := &expected[i]
		got, ok := byID[want.ID]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d: missing from SQLite", want.ID))
			continue
		}
		if got.Title != want.Title {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d title: got %q, want %q", want.ID, got.Title, want.Title))
		}
		if got.CategoryName != want.CategoryName {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d category: got %q, want %q", want.ID, got.CategoryName, want.CategoryName))
		}
		if got.Outcome != want.Outcome {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d outcome: got %q, want %q", want.ID, got.Outcome, want.Outcome))
		}
		if got.ReplyCount != want.ReplyCount {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d replyCount: got %d, want %d", want.ID, got.ReplyCount, want.ReplyCount))
		}
		if !got.CreatedAt.Equal(want.CreatedAt) {
			mismatches = append(mismatches, fmt.Sprintf("Topic %d createdAt: got %v, want %v", want.ID, got.CreatedAt, want.CreatedAt))
		}
	}

	fmt.Fprintf(b, "## Field verification\n\n")
	if len(mismatches) == 0 {
		fmt.Fprintf(b, "All %d topics match expected values (title, category, outcome, replyCount, createdAt).\n\n", len(expected))
	} else {
		fmt.Fprintf(b, "**%d mismatches found:**\n\n", len(mismatches))
		for _, m := range mismatches {
			fmt.Fprintf(b, "- %s\n", m)
		}
		fmt.Fprintf(b, "\n")
	}
}

func writeSampleTopics(b *strings.Builder, topics []model.Topic) {
	fmt.Fprintf(b, "## All topics\n\n")
	fmt.Fprintf(b, "| ID | Title | Category | Tags | Outcome | Replies | CreatedAt | FirstReply | Resolved | LastActivity |\n")
	fmt.Fprintf(b, "|----|-------|----------|------|---------|---------|-----------|------------|----------|-------------|\n")
	for i := range topics {
		tp := &topics[i]
		fmt.Fprintf(b, "| %d | %s | %s | %s | %s | %d | %s | %s | %s | %s |\n",
			tp.ID,
			truncate(tp.Title, 50),
			tp.CategoryName,
			strings.Join(tp.Tags, ", "),
			outcomeOrDash(tp.Outcome),
			tp.ReplyCount,
			fmtTime(tp.CreatedAt),
			fmtTimePtr(tp.FirstReplyAt),
			fmtTimePtr(tp.ResolvedAt),
			fmtTimePtr(tp.LastActivityAt),
		)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func outcomeOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func fmtTimePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}
