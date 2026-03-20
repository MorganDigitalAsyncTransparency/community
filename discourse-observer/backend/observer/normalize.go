// Spec: specs/observer/observer-behavior.md
// Tests: backend/pipeline_test.go, backend/sync_test.go
package observer

import (
	"strconv"

	"github.com/code-community/discourse-observer/backend/model"
)

// Normalize transforms a raw Discourse topic into a domain Topic.
func Normalize(raw *model.RawTopic, categories map[int]string, baseURL string) model.Topic {
	outcome := deriveOutcome(raw)

	t := model.Topic{
		ID:           raw.ID,
		Title:        raw.Title,
		CreatedAt:    raw.CreatedAt,
		Tags:         raw.Tags,
		CategoryName: categories[raw.CategoryID],
		ReplyCount:   raw.ReplyCount,
		Outcome:      outcome,
		FirstReplyAt: raw.FirstReplyAt,
		TopicURL:     baseURL + "/t/" + strconv.Itoa(raw.ID),
	}

	if t.Tags == nil {
		t.Tags = []string{}
	}

	switch outcome {
	case "solved":
		t.ResolvedAt = raw.AcceptedAnswerAt
	case "self-closed":
		t.ResolvedAt = raw.ClosedAt
	}

	// Use LastPostedAt as last activity; fall back to BumpedAt.
	if raw.LastPostedAt != nil {
		t.LastActivityAt = raw.LastPostedAt
	} else if raw.BumpedAt != nil {
		t.LastActivityAt = raw.BumpedAt
	}

	return t
}

// normalizeAll converts a slice of raw topics to domain topics.
func normalizeAll(raws []model.RawTopic, catMap map[int]string, baseURL string) []model.Topic {
	topics := make([]model.Topic, len(raws))
	for i := range raws {
		topics[i] = Normalize(&raws[i], catMap, baseURL)
	}
	return topics
}

func deriveOutcome(raw *model.RawTopic) string {
	if raw.HasAcceptedAnswer {
		return "solved"
	}
	if raw.Closed {
		return "self-closed"
	}
	return ""
}

func buildCategoryMap(cats []model.RawCategory) map[int]string {
	m := make(map[int]string, len(cats))
	for _, c := range cats {
		m[c.ID] = c.Name
	}
	return m
}
