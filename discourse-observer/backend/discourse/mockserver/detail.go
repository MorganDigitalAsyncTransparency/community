// Spec: specs/observer/detail-sync.md (DS-15, DS-16)
// Tests: backend/observer/detail_sync_test.go
package mockserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

// postInfo holds mock data for a topic's first post and its revisions.
type postInfo struct {
	PostID    int
	Version   int
	Revisions map[int]model.RawRevision // keyed by version (2..N)
}

// buildTopicMap creates a lookup from topic ID to RawTopic.
func buildTopicMap(rawTopics []model.RawTopic) map[int]model.RawTopic {
	m := make(map[int]model.RawTopic, len(rawTopics))
	for i := range rawTopics {
		m[rawTopics[i].ID] = rawTopics[i]
	}
	return m
}

// buildPostData creates mock post and revision data for each topic.
// Topics with tags get version > 1 (simulating tag addition revisions).
// Topics with multiple tags get additional revisions.
func buildPostData(topics []model.Topic, rawTopics []model.RawTopic, categories []model.RawCategory) map[int]postInfo {
	catByID := make(map[int]string, len(categories))
	for _, c := range categories {
		catByID[c.ID] = c.Name
	}

	topicByID := make(map[int]model.Topic, len(topics))
	for i := range topics {
		topicByID[topics[i].ID] = topics[i]
	}

	data := make(map[int]postInfo, len(rawTopics))
	for i := range rawTopics {
		rt := &rawTopics[i]
		postID := rt.ID*10 + 1 // deterministic post ID
		tp := topicByID[rt.ID]
		revs := buildRevisions(&tp, rt, catByID)
		data[rt.ID] = postInfo{
			PostID:    postID,
			Version:   len(revs) + 1, // version 1 is original
			Revisions: revs,
		}
	}
	return data
}

// buildRevisions creates plausible revision data for a topic.
func buildRevisions(tp *model.Topic, rt *model.RawTopic, catByID map[int]string) map[int]model.RawRevision {
	revs := make(map[int]model.RawRevision)
	version := 2
	baseTime := tp.CreatedAt.Add(time.Hour)

	// If topic has tags, simulate a tag addition revision.
	if len(tp.Tags) > 0 {
		revs[version] = model.RawRevision{
			CreatedAt: baseTime,
			Tags: &model.RevisionTagChange{
				Previous: []string{},
				Current:  tp.Tags,
			},
		}
		version++
	}

	// If topic changed category (has a non-default category), simulate a move.
	if tp.CategoryName == "Bug Reports" {
		revs[version] = model.RawRevision{
			CreatedAt: baseTime.Add(30 * time.Minute),
			CategoryID: &model.RevisionIntChange{
				Previous: 1, // "Support"
				Current:  rt.CategoryID,
			},
		}
	}

	return revs
}

func handleTopicDetail(topicMap map[int]model.RawTopic, postData map[int]postInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimSuffix(r.PathValue("id"), ".json")
		topicID, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if _, ok := topicMap[topicID]; !ok {
			http.NotFound(w, r)
			return
		}

		pd := postData[topicID]
		resp := model.RawTopicDetail{
			ID: topicID,
			PostStream: model.PostStream{
				Posts: []model.RawPost{{
					ID:         pd.PostID,
					PostNumber: 1,
					Version:    pd.Version,
				}},
			},
		}
		writeJSON(w, resp)
	}
}

func handlePostRevision(postData map[int]postInfo) http.HandlerFunc {
	// Build a reverse lookup: postID → topicID
	postToTopic := make(map[int]int, len(postData))
	for topicID, pd := range postData {
		postToTopic[pd.PostID] = topicID
	}

	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.Atoi(r.PathValue("postID"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		version, err := strconv.Atoi(strings.TrimSuffix(r.PathValue("version"), ".json"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		topicID, ok := postToTopic[postID]
		if !ok {
			http.NotFound(w, r)
			return
		}

		pd := postData[topicID]
		rev, ok := pd.Revisions[version]
		if !ok {
			http.NotFound(w, r)
			return
		}

		// Wrap in the Discourse response envelope.
		resp := struct {
			model.RawRevision
			CurrentRevision int `json:"current_revision"`
			VersionCount    int `json:"version_count"`
		}{
			RawRevision:     rev,
			CurrentRevision: pd.Version,
			VersionCount:    pd.Version,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
