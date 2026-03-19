// Spec: specs/discourse/discourse-source-model.md
// Tests: backend/pipeline_test.go
package model

import "time"

// RawTopic represents a topic as returned by the Discourse API.
// Fields match the /latest.json topic shape with additional fields
// from topic detail endpoints and the Solved plugin.
type RawTopic struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	Slug              string     `json:"slug"`
	CreatedAt         time.Time  `json:"created_at"`
	CategoryID        int        `json:"category_id"`
	Tags              []string   `json:"tags"`
	ReplyCount        int        `json:"reply_count"`
	PostsCount        int        `json:"posts_count"`
	Closed            bool       `json:"closed"`
	HasAcceptedAnswer bool       `json:"has_accepted_answer"`
	LastPostedAt      *time.Time `json:"last_posted_at,omitempty"`
	BumpedAt          *time.Time `json:"bumped_at,omitempty"`
	// Extended fields — available from topic detail or mock server.
	// Real Discourse requires per-topic detail fetches for these.
	FirstReplyAt     *time.Time `json:"first_reply_at,omitempty"`
	AcceptedAnswerAt *time.Time `json:"accepted_answer_at,omitempty"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
}

// RawCategory represents a category from the Discourse API /categories.json.
type RawCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
