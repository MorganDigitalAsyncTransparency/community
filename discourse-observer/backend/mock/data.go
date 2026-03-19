// Spec: specs/api/api-contract.md
// Tests: backend/api/contract_test.go
package mock

import (
	"fmt"
	"time"

	"github.com/code-community/discourse-observer/backend/model"
)

const ForumBaseURL = "https://forum.example.com"

func topicURL(id int) string {
	return fmt.Sprintf("%s/t/%d", ForumBaseURL, id)
}

func ptr(t time.Time) *time.Time { return &t }

func t(s string) time.Time {
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic("invalid time: " + s)
	}
	return parsed
}

// Topics returns the full set of mock topics for the API.
func Topics() []model.Topic {
	return []model.Topic{
		// Unreplied (tagged)
		{ID: 1001, Title: "Cannot connect to API endpoint", Tags: []string{"api"},
			CategoryName: "Integration", CreatedAt: t("2026-03-16T10:00:00Z"),
			TopicURL: topicURL(1001)},
		{ID: 1002, Title: "SSO login fails with SAML", Tags: []string{"authentication", "sso"},
			CategoryName: "Access", CreatedAt: t("2026-03-08T14:30:00Z"),
			TopicURL: topicURL(1002)},
		{ID: 1003, Title: "Editor toolbar missing buttons", Tags: []string{"editor", "plugin"},
			CategoryName: "Content", CreatedAt: t("2026-02-01T09:00:00Z"),
			TopicURL: topicURL(1003)},
		{ID: 1004, Title: "Installation guide unclear", Tags: []string{"installation"},
			CategoryName: "Infrastructure", CreatedAt: t("2026-03-17T16:00:00Z"),
			TopicURL: topicURL(1004)},
		{ID: 1005, Title: "API rate limiting not working", Tags: []string{"api"},
			CategoryName: "Integration", CreatedAt: t("2025-02-10T08:00:00Z"),
			TopicURL: topicURL(1005)},

		// Resolved — solved
		{ID: 1006, Title: "Webhook payload format changed", Tags: []string{"api", "webhooks"},
			CategoryName: "Integration", ReplyCount: 2,
			CreatedAt:    t("2026-03-13T10:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-13T12:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-14T10:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1006)},
		{ID: 1007, Title: "Password reset email not sent", Tags: []string{"authentication"},
			CategoryName: "Access", ReplyCount: 3,
			CreatedAt:    t("2026-03-03T08:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-03T14:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-05T08:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1007)},
		{ID: 1008, Title: "Search results not updating", Tags: []string{"search"},
			CategoryName: "Content", ReplyCount: 1,
			CreatedAt:    t("2026-03-10T11:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-10T12:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-10T23:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1008)},
		{ID: 1009, Title: "Plugin compatibility issue", Tags: []string{"plugin"},
			CategoryName: "Content", ReplyCount: 4,
			CreatedAt:    t("2026-03-15T09:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-15T19:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-18T09:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1009)},
		{ID: 1010, Title: "Webhook delivery failures", Tags: []string{"api", "webhooks"},
			CategoryName: "Integration", ReplyCount: 2,
			CreatedAt:    t("2026-02-26T10:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-26T15:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-27T22:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1010)},

		// Resolved — self-closed
		{ID: 1011, Title: "SSL certificate renewal fails", Tags: []string{"ssl"},
			CategoryName: "Access", ReplyCount: 1,
			CreatedAt:    t("2026-03-11T09:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-11T12:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-14T09:00:00Z")),
			Outcome:      "self-closed", TopicURL: topicURL(1011)},
		{ID: 1012, Title: "Data import timeout", Tags: []string{"data-import"},
			CategoryName: "Infrastructure", ReplyCount: 0,
			CreatedAt:  t("2026-03-06T10:00:00Z"),
			ResolvedAt: ptr(t("2026-03-07T10:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1012)},
		{ID: 1013, Title: "Migration script error", Tags: []string{"migration"},
			CategoryName: "Other", ReplyCount: 2,
			CreatedAt:    t("2026-02-21T10:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-23T10:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-28T10:00:00Z")),
			Outcome:      "self-closed", TopicURL: topicURL(1013)},

		// Replied-open (tagged)
		{ID: 1014, Title: "API documentation outdated", Tags: []string{"api"},
			CategoryName: "Integration", ReplyCount: 3,
			CreatedAt:      t("2026-02-26T14:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-26T22:00:00Z")),
			LastActivityAt: ptr(t("2026-02-28T14:00:00Z")),
			TopicURL:       topicURL(1014)},
		{ID: 1015, Title: "Editor crashes on large posts", Tags: []string{"editor"},
			CategoryName: "Content", ReplyCount: 1,
			CreatedAt:      t("2026-02-16T10:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-16T14:00:00Z")),
			LastActivityAt: ptr(t("2026-03-15T10:00:00Z")),
			TopicURL:       topicURL(1015)},
		{ID: 1016, Title: "SSO session timeout too short", Tags: []string{"authentication", "sso"},
			CategoryName: "Access", ReplyCount: 2,
			CreatedAt:      t("2026-03-08T08:00:00Z"),
			FirstReplyAt:   ptr(t("2026-03-08T20:00:00Z")),
			LastActivityAt: ptr(t("2026-03-13T08:00:00Z")),
			TopicURL:       topicURL(1016)},

		// Untagged
		{ID: 1017, Title: "How to customize theme colors", Tags: []string{},
			CategoryName: "General", CreatedAt: t("2026-03-13T15:00:00Z"),
			TopicURL: topicURL(1017)},
		{ID: 1018, Title: "Forum loading slowly", Tags: []string{},
			CategoryName: "Support", CreatedAt: t("2026-03-03T09:00:00Z"),
			TopicURL: topicURL(1018)},
		{ID: 1019, Title: "Cannot upload images", Tags: []string{},
			CategoryName: "General", ReplyCount: 1,
			CreatedAt:    t("2026-03-15T11:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-15T17:00:00Z")),
			TopicURL:     topicURL(1019)},
	}
}
