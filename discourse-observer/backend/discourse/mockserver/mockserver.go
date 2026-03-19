// Spec: specs/discourse/discourse-source-model.md
// Tests: backend/pipeline_test.go
//
// Package mockserver provides an HTTP server that mimics the Discourse API
// using the project's existing mock dataset. It serves /latest.json and
// /categories.json in the same JSON shape as a real Discourse instance.
//
// Usage in tests:
//
//	srv := mockserver.New()
//	defer srv.Close()
//	client := discourse.NewClient(srv.URL, "", "")
package mockserver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/code-community/discourse-observer/backend/mock"
	"github.com/code-community/discourse-observer/backend/model"
)

// New starts an httptest.Server that serves Discourse-format responses
// built from mock.Topics().
func New() *httptest.Server {
	topics := mock.Topics()
	categories := buildCategories(topics)
	rawTopics := convertTopics(topics, categories)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /latest.json", handleLatest(rawTopics))
	mux.HandleFunc("GET /categories.json", handleCategories(categories))

	return httptest.NewServer(mux)
}

func handleLatest(topics []model.RawTopic) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp := struct {
			TopicList struct {
				Topics []model.RawTopic `json:"topics"`
			} `json:"topic_list"`
		}{}
		resp.TopicList.Topics = topics
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
		raw[i] = model.RawTopic{
			ID:           tp.ID,
			Title:        tp.Title,
			Slug:         slugify(tp.Title),
			CreatedAt:    tp.CreatedAt,
			CategoryID:   catMap[tp.CategoryName],
			Tags:         tp.Tags,
			ReplyCount:   tp.ReplyCount,
			PostsCount:   tp.ReplyCount + 1,
			LastPostedAt: tp.LastActivityAt,
			BumpedAt:     tp.LastActivityAt,
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
