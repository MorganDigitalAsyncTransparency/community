// Spec: specs/discourse/discourse-source-model.md
// Tests: backend/discourse/client_test.go, backend/pipeline_test.go, backend/sync_test.go
//
// Package mockserver provides an HTTP server that mimics the Discourse API
// using the project's existing mock dataset. It serves /latest.json and
// /categories.json in the same JSON shape as a real Discourse instance.
//
// The server supports pagination: /latest.json?page=N returns a page of
// topics (default 30 per page) and includes more_topics_url when more
// pages are available.
//
// Usage in tests:
//
//	srv := mockserver.New()
//	defer srv.Close()
//	client := discourse.NewClient(srv.URL, "", "")
//
// Usage as a standalone server:
//
//	http.ListenAndServe(":9920", mockserver.Handler())
package mockserver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/model"
)

const defaultPageSize = 30

// Handler returns an http.Handler serving Discourse-format responses
// built from mock.Topics(). Suitable for both httptest.Server and
// standalone HTTP servers.
func Handler() http.Handler {
	return HandlerWithPageSize(defaultPageSize)
}

// HandlerWithPageSize returns an http.Handler with a custom page size.
func HandlerWithPageSize(pageSize int) http.Handler {
	topics := mock.Topics()
	categories := buildCategories(topics)
	rawTopics := convertTopics(topics, categories)
	sortByBumpedAtDesc(rawTopics)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /about.json", handleAbout(len(rawTopics)))
	mux.HandleFunc("GET /latest.json", handleLatest(rawTopics, pageSize))
	mux.HandleFunc("GET /categories.json", handleCategories(categories))

	return mux
}

// New starts an httptest.Server that serves Discourse-format responses
// built from mock.Topics().
func New() *httptest.Server {
	return httptest.NewServer(Handler())
}

// NewWithPageSize starts an httptest.Server with a custom page size.
// Useful for tests that need to verify pagination with small pages.
func NewWithPageSize(pageSize int) *httptest.Server {
	return httptest.NewServer(HandlerWithPageSize(pageSize))
}

func handleLatest(topics []model.RawTopic, pageSize int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := 0
		if p := r.URL.Query().Get("page"); p != "" {
			if n, err := strconv.Atoi(p); err == nil && n >= 0 {
				page = n
			}
		}

		start := page * pageSize
		if start >= len(topics) {
			start = len(topics)
		}
		end := start + pageSize
		if end > len(topics) {
			end = len(topics)
		}
		pageTopics := topics[start:end]

		resp := struct {
			TopicList struct {
				Topics        []model.RawTopic `json:"topics"`
				MoreTopicsURL string           `json:"more_topics_url,omitempty"`
			} `json:"topic_list"`
		}{}
		resp.TopicList.Topics = pageTopics

		if end < len(topics) {
			resp.TopicList.MoreTopicsURL = "/latest.json?page=" + strconv.Itoa(page+1)
		}

		writeJSON(w, resp)
	}
}

func handleAbout(topicCount int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp := struct {
			About struct {
				Stats struct {
					TopicCount int `json:"topic_count"`
				} `json:"stats"`
			} `json:"about"`
		}{}
		resp.About.Stats.TopicCount = topicCount
		writeJSON(w, resp)
	}
}

func handleCategories(categories []model.RawCategory) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp := struct {
			CategoryList struct {
				Categories []model.RawCategory `json:"categories"`
			} `json:"category_list"`
		}{}
		resp.CategoryList.Categories = categories
		writeJSON(w, resp)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("mockserver: write failed: %v", err)
	}
}

// buildCategories extracts unique category names from topics and assigns IDs.
func buildCategories(topics []model.Topic) []model.RawCategory {
	seen := map[string]int{}
	var cats []model.RawCategory
	nextID := 1
	for i := range topics {
		name := topics[i].CategoryName
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = nextID
		cats = append(cats, model.RawCategory{
			ID:   nextID,
			Name: name,
			Slug: slugify(name),
		})
		nextID++
	}
	return cats
}

// convertTopics maps model.Topic values to Discourse API RawTopic format.
func convertTopics(topics []model.Topic, cats []model.RawCategory) []model.RawTopic {
	catMap := make(map[string]int, len(cats))
	for _, c := range cats {
		catMap[c.Name] = c.ID
	}

	raw := make([]model.RawTopic, len(topics))
	for i := range topics {
		tp := &topics[i]
		bump := effectiveBump(tp)
		raw[i] = model.RawTopic{
			ID:           tp.ID,
			Title:        tp.Title,
			Slug:         slugify(tp.Title),
			CreatedAt:    tp.CreatedAt,
			CategoryID:   catMap[tp.CategoryName],
			Tags:         tp.Tags,
			ReplyCount:   tp.ReplyCount,
			PostsCount:   tp.ReplyCount + 1,
			LastPostedAt: bump,
			BumpedAt:     bump,
			FirstReplyAt: tp.FirstReplyAt,
		}
		if raw[i].Tags == nil {
			raw[i].Tags = []string{}
		}

		switch tp.Outcome {
		case "solved":
			raw[i].HasAcceptedAnswer = true
			raw[i].Closed = true
			raw[i].AcceptedAnswerAt = tp.ResolvedAt
		case "self-closed":
			raw[i].Closed = true
			raw[i].ClosedAt = tp.ResolvedAt
		}
	}
	return raw
}

// effectiveBump returns the best approximation of Discourse's bumped_at.
// Real Discourse always populates bumped_at; for topics with no activity it
// equals created_at.
func effectiveBump(tp *model.Topic) *time.Time {
	switch {
	case tp.LastActivityAt != nil:
		return tp.LastActivityAt
	case tp.ResolvedAt != nil:
		return tp.ResolvedAt
	case tp.FirstReplyAt != nil:
		return tp.FirstReplyAt
	default:
		t := tp.CreatedAt
		return &t
	}
}

// sortByBumpedAtDesc orders topics by BumpedAt descending, matching real
// Discourse /latest.json behavior. Ties are broken by ID descending.
func sortByBumpedAtDesc(topics []model.RawTopic) {
	sort.Slice(topics, func(i, j int) bool {
		bi, bj := topics[i].BumpedAt, topics[j].BumpedAt
		switch {
		case bi == nil && bj == nil:
			return topics[i].ID > topics[j].ID
		case bi == nil:
			return false
		case bj == nil:
			return true
		case bi.Equal(*bj):
			return topics[i].ID > topics[j].ID
		default:
			return bi.After(*bj)
		}
	})
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	// Collapse consecutive dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
