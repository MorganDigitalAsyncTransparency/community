// Spec: specs/observer/detail-sync.md (DS-1, DS-2, DS-3)
// Tests: backend/observer/detail_sync_test.go
package model

import "time"

// RawTopicDetail represents relevant fields from /t/{id}.json.
type RawTopicDetail struct {
	ID         int       `json:"id"`
	PostStream PostStream `json:"post_stream"`
}

// PostStream holds the posts array from a topic detail response.
type PostStream struct {
	Posts []RawPost `json:"posts"`
}

// RawPost represents a post within a topic detail response.
type RawPost struct {
	ID         int `json:"id"`
	PostNumber int `json:"post_number"`
	Version    int `json:"version"`
}

// RawRevision represents a single revision from
// /posts/{id}/revisions/{v}.json. Field names match the Discourse API.
type RawRevision struct {
	CreatedAt  time.Time          `json:"created_at"`
	Title      *RevisionChange    `json:"title_changes,omitempty"`
	CategoryID *RevisionIntChange `json:"category_id_changes,omitempty"`
	Tags       *RevisionTagChange `json:"tags_changes,omitempty"`
}

// RevisionChange holds previous/current string values for a revision field.
type RevisionChange struct {
	Previous string `json:"previous"`
	Current  string `json:"current"`
}

// RevisionIntChange holds previous/current int values for a revision field.
type RevisionIntChange struct {
	Previous int `json:"previous"`
	Current  int `json:"current"`
}

// RevisionTagChange holds previous/current string slices for tag changes.
type RevisionTagChange struct {
	Previous []string `json:"previous"`
	Current  []string `json:"current"`
}

// TopicDetailState holds a topic's detail sync progress.
type TopicDetailState struct {
	TopicID      int
	LastRevision int
}

// TopicEvent represents a stored event extracted from a revision.
type TopicEvent struct {
	ID        int       `json:"id,omitempty"`
	TopicID   int       `json:"topic_id"`
	EventType string    `json:"event_type"`
	HappenedAt time.Time `json:"happened_at"`
	Detail    string    `json:"detail"`
}
