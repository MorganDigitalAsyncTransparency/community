// Spec: specs/discourse/discourse-source-model.md
package discourse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/code-community/discourse-observer/backend/discourse"
	"github.com/code-community/discourse-observer/backend/discourse/mockserver"
	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/model"
)

func TestFetchTopicsPagesCollectsAllTopics(t *testing.T) {
	srv := mockserver.NewWithPageSize(10)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	var allTopics []model.RawTopic
	var pagesSeen []int

	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{}, func(topics []model.RawTopic, page int) error {
		allTopics = append(allTopics, topics...)
		pagesSeen = append(pagesSeen, page)
		return nil
	})
	if err != nil {
		t.Fatalf("FetchTopicsPages: %v", err)
	}

	expected := len(mock.Topics())
	if len(allTopics) != expected {
		t.Errorf("got %d topics, want %d", len(allTopics), expected)
	}

	// With 44 topics and page size 10: pages 0,1,2,3,4
	wantPages := (expected + 9) / 10
	if len(pagesSeen) != wantPages {
		t.Errorf("got %d pages, want %d", len(pagesSeen), wantPages)
	}
	for i, p := range pagesSeen {
		if p != i {
			t.Errorf("page %d: got page number %d", i, p)
		}
	}
}

func TestFetchTopicsPagesStartPage(t *testing.T) {
	srv := mockserver.NewWithPageSize(10)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	var pagesSeen []int
	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{StartPage: 2}, func(_ []model.RawTopic, page int) error {
		pagesSeen = append(pagesSeen, page)
		return nil
	})
	if err != nil {
		t.Fatalf("FetchTopicsPages: %v", err)
	}

	if len(pagesSeen) == 0 {
		t.Fatal("no pages seen")
	}
	if pagesSeen[0] != 2 {
		t.Errorf("first page = %d, want 2", pagesSeen[0])
	}
}

func TestFetchTopicsPagesCallbackError(t *testing.T) {
	srv := mockserver.NewWithPageSize(10)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	callbackErr := context.Canceled
	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{}, func(_ []model.RawTopic, _ int) error {
		return callbackErr
	})
	if err != callbackErr {
		t.Errorf("got error %v, want %v", err, callbackErr)
	}
}

func TestFetchTopicsPagesContextCanceled(t *testing.T) {
	srv := mockserver.NewWithPageSize(10)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	ctx, cancel := context.WithCancel(context.Background())
	var pages int
	err := client.FetchTopicsPagesWithConfig(ctx, discourse.PageConfig{Delay: 50 * time.Millisecond}, func(_ []model.RawTopic, _ int) error {
		pages++
		if pages >= 2 {
			cancel()
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
}

func TestFetchTopicsPagesHTTP429(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			// Retry-After: 0 means retry immediately (valid per RFC 7231).
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		// Return a single page with no more_topics_url.
		resp := struct {
			TopicList struct {
				Topics []model.RawTopic `json:"topics"`
			} `json:"topic_list"`
		}{}
		resp.TopicList.Topics = []model.RawTopic{{ID: 1, Title: "test"}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	var gotTopics int
	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{MaxRetries: 3, RetryDelay: 10 * time.Millisecond}, func(topics []model.RawTopic, _ int) error {
		gotTopics += len(topics)
		return nil
	})
	if err != nil {
		t.Fatalf("FetchTopicsPages: %v", err)
	}
	if gotTopics != 1 {
		t.Errorf("got %d topics, want 1", gotTopics)
	}
	if c := calls.Load(); c != 2 {
		t.Errorf("server got %d calls, want 2 (1 x 429 + 1 success)", c)
	}
}

func TestFetchTopicsPagesHTTP429FallbackDelay(t *testing.T) {
	// When 429 has no Retry-After and retryDelay is 0, the client must not
	// spin in a tight loop — it should use rateLimitFallback (10s).
	// We verify by canceling the context quickly and checking that only
	// one retry was attempted (the fallback sleep was interrupted).
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_ = client.FetchTopicsPagesWithConfig(ctx, discourse.PageConfig{}, func(_ []model.RawTopic, _ int) error {
		return nil
	})

	// With a 10s fallback, only 1 call should happen before the 200ms timeout.
	// A tight loop would produce hundreds.
	if c := calls.Load(); c > 2 {
		t.Errorf("server got %d calls — fallback delay not working (expected ≤ 2)", c)
	}
}

func TestFetchTopicsPagesHTTP5xxRetry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		resp := struct {
			TopicList struct {
				Topics []model.RawTopic `json:"topics"`
			} `json:"topic_list"`
		}{}
		resp.TopicList.Topics = []model.RawTopic{{ID: 1, Title: "test"}}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{MaxRetries: 3, RetryDelay: 10 * time.Millisecond}, func(_ []model.RawTopic, _ int) error {
		return nil
	})
	if err != nil {
		t.Fatalf("FetchTopicsPages: %v", err)
	}
	if c := calls.Load(); c != 3 {
		t.Errorf("server got %d calls, want 3", c)
	}
}

func TestFetchTopicsPagesHTTP5xxExhaustsRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{MaxRetries: 2, RetryDelay: 10 * time.Millisecond}, func(_ []model.RawTopic, _ int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
}

func TestFetchTopicsPagesHTTP4xxNoRetry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{MaxRetries: 3, RetryDelay: 10 * time.Millisecond}, func(_ []model.RawTopic, _ int) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error on 403")
	}
	if c := calls.Load(); c != 1 {
		t.Errorf("server got %d calls, want 1 (no retry on 4xx)", c)
	}
}

func TestFetchTopicsPagesDelay(t *testing.T) {
	pageSize := 10
	srv := mockserver.NewWithPageSize(pageSize)
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")

	delay := 50 * time.Millisecond
	start := time.Now()
	var pages int
	err := client.FetchTopicsPagesWithConfig(context.Background(), discourse.PageConfig{Delay: delay}, func(_ []model.RawTopic, _ int) error {
		pages++
		return nil
	})
	if err != nil {
		t.Fatalf("FetchTopicsPages: %v", err)
	}
	elapsed := time.Since(start)

	// Delay applies between pages (not before the first), so expect (pages-1)*delay.
	minExpected := time.Duration(pages-1) * delay
	if elapsed < minExpected {
		t.Errorf("elapsed %v < expected minimum %v for %d pages with %v delay", elapsed, minExpected, pages, delay)
	}
}

func TestFetchTopicsCollectsAll(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	topics, err := client.FetchTopics(context.Background())
	if err != nil {
		t.Fatalf("FetchTopics: %v", err)
	}
	expected := len(mock.Topics())
	if len(topics) != expected {
		t.Errorf("got %d topics, want %d", len(topics), expected)
	}
}

func TestMockserverPaginationBeyondLastPage(t *testing.T) {
	srv := mockserver.NewWithPageSize(10)
	defer srv.Close()

	totalTopics := len(mock.Topics())
	lastPage := (totalTopics + 9) / 10

	// Request a page beyond the last.
	resp, err := http.Get(srv.URL + "/latest.json?page=" + strconv.Itoa(lastPage+5))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		TopicList struct {
			Topics        []model.RawTopic `json:"topics"`
			MoreTopicsURL string           `json:"more_topics_url"`
		} `json:"topic_list"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)

	if len(body.TopicList.Topics) != 0 {
		t.Errorf("got %d topics beyond last page, want 0", len(body.TopicList.Topics))
	}
	if body.TopicList.MoreTopicsURL != "" {
		t.Error("got more_topics_url beyond last page, want empty")
	}
}

func TestFetchTopicDetail(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	topics := mock.Topics()
	// Pick a topic with tags — it will have revisions.
	var taggedID int
	for _, tp := range topics {
		if len(tp.Tags) > 0 {
			taggedID = tp.ID
			break
		}
	}
	if taggedID == 0 {
		t.Fatal("no tagged topic found in mock data")
	}

	detail, err := client.FetchTopicDetail(context.Background(), taggedID)
	if err != nil {
		t.Fatalf("FetchTopicDetail: %v", err)
	}
	if detail.ID != taggedID {
		t.Errorf("detail ID = %d, want %d", detail.ID, taggedID)
	}
	if len(detail.PostStream.Posts) == 0 {
		t.Fatal("expected at least one post")
	}
	if detail.PostStream.Posts[0].Version < 2 {
		t.Errorf("tagged topic should have version >= 2, got %d", detail.PostStream.Posts[0].Version)
	}
}

func TestFetchTopicDetail404(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	_, err := client.FetchTopicDetail(context.Background(), 999999)
	if err == nil {
		t.Fatal("expected error for non-existent topic")
	}
}

func TestFetchPostRevision(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	topics := mock.Topics()
	// Find a tagged topic to get its post ID and revision.
	var taggedID int
	for _, tp := range topics {
		if len(tp.Tags) > 0 {
			taggedID = tp.ID
			break
		}
	}

	detail, err := client.FetchTopicDetail(context.Background(), taggedID)
	if err != nil {
		t.Fatalf("FetchTopicDetail: %v", err)
	}
	postID := detail.PostStream.Posts[0].ID
	version := detail.PostStream.Posts[0].Version
	if version < 2 {
		t.Skip("topic has no revisions")
	}

	rev, err := client.FetchPostRevision(context.Background(), postID, 2)
	if err != nil {
		t.Fatalf("FetchPostRevision: %v", err)
	}
	if rev.CreatedAt.IsZero() {
		t.Error("revision created_at should not be zero")
	}
	// Tagged topic's revision 2 should have tag changes.
	if rev.Tags == nil {
		t.Error("expected tags_changes in revision 2 of a tagged topic")
	}
}

func TestMockServerTopicDetail(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	topics := mock.Topics()
	client := discourse.NewClient(srv.URL, "", "")

	for _, tp := range topics {
		detail, err := client.FetchTopicDetail(context.Background(), tp.ID)
		if err != nil {
			t.Errorf("FetchTopicDetail(%d): %v", tp.ID, err)
			continue
		}
		if detail.ID != tp.ID {
			t.Errorf("detail.ID = %d, want %d", detail.ID, tp.ID)
		}
		if len(detail.PostStream.Posts) == 0 {
			t.Errorf("topic %d: no posts", tp.ID)
		}
	}
}

func TestMockServerRevisions(t *testing.T) {
	srv := mockserver.New()
	defer srv.Close()

	client := discourse.NewClient(srv.URL, "", "")
	topics := mock.Topics()

	var revisionsFound int
	for _, tp := range topics {
		detail, err := client.FetchTopicDetail(context.Background(), tp.ID)
		if err != nil {
			t.Fatalf("FetchTopicDetail(%d): %v", tp.ID, err)
		}
		post := detail.PostStream.Posts[0]
		for v := 2; v <= post.Version; v++ {
			rev, err := client.FetchPostRevision(context.Background(), post.ID, v)
			if err != nil {
				t.Errorf("FetchPostRevision(%d, %d): %v", post.ID, v, err)
				continue
			}
			if rev.CreatedAt.IsZero() {
				t.Errorf("revision %d/%d has zero created_at", post.ID, v)
			}
			revisionsFound++
		}
	}

	if revisionsFound == 0 {
		t.Error("expected at least some revisions in mock data")
	}
}
