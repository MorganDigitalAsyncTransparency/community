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
//
// This dataset mirrors the original frontend mock data (frontend/src/mock/data.ts)
// with relative dates anchored to 2026-03-19T12:00:00Z. Keeping the same topics,
// titles, tags, and time relationships lets us compare backend-served output against
// the former frontend-computed output.
func Topics() []model.Topic {
	return []model.Topic{

		// ---------------------------------------------------------------
		// Unreplied (tagged) — 11 topics
		// ---------------------------------------------------------------

		{ID: 1041, Title: "Cannot authenticate with API key after upgrade",
			Tags: []string{"authentication"}, CategoryName: "Support",
			CreatedAt: t("2026-03-05T12:00:00Z"), TopicURL: topicURL(1041)},

		{ID: 1055, Title: "Installation fails on Ubuntu 24.04 — missing libssl dependency",
			Tags: []string{"installation"}, CategoryName: "Support",
			CreatedAt: t("2026-03-08T12:00:00Z"), TopicURL: topicURL(1055)},

		{ID: 1063, Title: "SSO login redirect loop with SAML provider",
			Tags: []string{"authentication", "sso"}, CategoryName: "Support",
			CreatedAt: t("2026-03-10T12:00:00Z"), TopicURL: topicURL(1063)},

		{ID: 1078, Title: "Webhook delivery silently drops events over 1 MB",
			Tags: []string{"webhooks"}, CategoryName: "Support",
			CreatedAt: t("2026-03-12T12:00:00Z"), TopicURL: topicURL(1078)},

		{ID: 1089, Title: "Search index not rebuilding after plugin update",
			Tags: []string{"search"}, CategoryName: "Support",
			CreatedAt: t("2026-03-14T12:00:00Z"), TopicURL: topicURL(1089)},

		{ID: 1094, Title: "Email notifications delayed by several hours",
			Tags: []string{"email"}, CategoryName: "Support",
			CreatedAt: t("2026-03-15T12:00:00Z"), TopicURL: topicURL(1094)},

		{ID: 1102, Title: "Rate limiting returns 429 even below documented threshold",
			Tags: []string{"api"}, CategoryName: "Support",
			CreatedAt: t("2026-03-16T12:00:00Z"), TopicURL: topicURL(1102)},

		{ID: 1108, Title: "Bulk import fails with timeout on large CSV",
			Tags: []string{"data-import"}, CategoryName: "Support",
			CreatedAt: t("2026-03-17T12:00:00Z"), TopicURL: topicURL(1108)},

		{ID: 1115, Title: "Markdown rendering broken in post preview",
			Tags: []string{"editor"}, CategoryName: "Support",
			CreatedAt: t("2026-03-18T12:00:00Z"), TopicURL: topicURL(1115)},

		{ID: 980, Title: "Custom domain SSL certificate not renewing automatically",
			Tags: []string{"ssl", "configuration"}, CategoryName: "Support",
			CreatedAt: t("2026-02-02T12:00:00Z"), TopicURL: topicURL(980)},

		{ID: 921, Title: "Admin panel inaccessible after version 3.0 upgrade",
			Tags: []string{"administration"}, CategoryName: "Support",
			CreatedAt: t("2025-02-12T12:00:00Z"), TopicURL: topicURL(921)},

		// ---------------------------------------------------------------
		// Untagged — 7 topics
		// ---------------------------------------------------------------

		{ID: 1044, Title: "How do I change the default theme colors?",
			Tags: []string{}, CategoryName: "General", ReplyCount: 3,
			CreatedAt: t("2026-03-07T12:00:00Z"), TopicURL: topicURL(1044)},

		{ID: 1067, Title: "Sidebar navigation disappeared after update",
			Tags: []string{}, CategoryName: "Bug Reports", ReplyCount: 1,
			CreatedAt: t("2026-03-11T12:00:00Z"), TopicURL: topicURL(1067)},

		{ID: 1085, Title: "Best practices for category permissions?",
			Tags: []string{}, CategoryName: "General", ReplyCount: 5,
			CreatedAt: t("2026-03-13T12:00:00Z"), TopicURL: topicURL(1085)},

		{ID: 1098, Title: "Mobile layout issues on Galaxy S24",
			Tags: []string{}, CategoryName: "Bug Reports", ReplyCount: 2,
			CreatedAt: t("2026-03-16T12:00:00Z"), TopicURL: topicURL(1098)},

		{ID: 1112, Title: "User group sync with external directory not working",
			Tags: []string{}, CategoryName: "Support",
			CreatedAt: t("2026-03-18T12:00:00Z"), TopicURL: topicURL(1112)},

		{ID: 977, Title: "Category description missing on mobile view",
			Tags: []string{}, CategoryName: "Bug Reports",
			CreatedAt: t("2026-02-02T12:00:00Z"), TopicURL: topicURL(977)},

		{ID: 910, Title: "Emoji picker not loading on Safari 15",
			Tags: []string{}, CategoryName: "Bug Reports", ReplyCount: 2,
			CreatedAt: t("2025-02-12T12:00:00Z"), TopicURL: topicURL(910)},

		// ---------------------------------------------------------------
		// Resolved — solved — 13 topics
		// ---------------------------------------------------------------

		{ID: 1001, Title: "API rate limit not resetting after cooldown period",
			Tags: []string{"api"}, CategoryName: "Support", ReplyCount: 4,
			CreatedAt:    t("2026-02-19T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-19T15:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-21T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1001)},

		{ID: 1005, Title: "OAuth2 token refresh fails silently",
			Tags: []string{"authentication"}, CategoryName: "Support", ReplyCount: 6,
			CreatedAt:    t("2026-02-21T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-21T20:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-26T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1005)},

		{ID: 1008, Title: "Docker setup crashes on Apple Silicon",
			Tags: []string{"installation"}, CategoryName: "Support", ReplyCount: 3,
			CreatedAt:    t("2026-02-22T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-22T13:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-23T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1008)},

		{ID: 1016, Title: "Email digest contains duplicate entries",
			Tags: []string{"email"}, CategoryName: "Bug Reports", ReplyCount: 5,
			CreatedAt:    t("2026-02-25T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-26T00:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-01T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1016)},

		{ID: 1019, Title: "Full-text search returns stale results",
			Tags: []string{"search"}, CategoryName: "Support", ReplyCount: 2,
			CreatedAt:    t("2026-02-26T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-27T12:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-05T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1019)},

		{ID: 1025, Title: "Initial setup wizard skips database migration step",
			Tags: []string{"setup"}, CategoryName: "Support", ReplyCount: 7,
			CreatedAt:    t("2026-03-01T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-01T14:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-04T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1025)},

		{ID: 1028, Title: "API pagination returns wrong total count",
			Tags: []string{"api"}, CategoryName: "Bug Reports", ReplyCount: 3,
			CreatedAt:    t("2026-03-02T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-02T18:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-04T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1028)},

		{ID: 1033, Title: "Webhook retry logic sends duplicate payloads",
			Tags: []string{"webhooks"}, CategoryName: "Support", ReplyCount: 4,
			CreatedAt:    t("2026-03-05T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-06T06:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-11T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1033)},

		{ID: 1036, Title: "Email templates not rendering HTML correctly",
			Tags: []string{"email"}, CategoryName: "Support", ReplyCount: 2,
			CreatedAt:    t("2026-03-07T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-09T00:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-15T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1036)},

		{ID: 1042, Title: "Plugin installation fails with permission error",
			Tags: []string{"installation", "setup"}, CategoryName: "Support", ReplyCount: 3,
			CreatedAt:    t("2026-03-11T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-11T16:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-12T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1042)},

		{ID: 1046, Title: "Editor toolbar disappears in fullscreen mode",
			Tags: []string{"editor"}, CategoryName: "Bug Reports", ReplyCount: 1,
			CreatedAt:    t("2026-03-13T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-15T12:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-16T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1046)},

		{ID: 1052, Title: "Webhook endpoint health check returns false positive",
			Tags: []string{"webhooks"}, CategoryName: "Support", ReplyCount: 5,
			CreatedAt:    t("2026-03-16T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-03-16T14:00:00Z")),
			ResolvedAt:   ptr(t("2026-03-17T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(1052)},

		{ID: 975, Title: "Backup export silently truncates topics over 10 MB",
			Tags: []string{"backup"}, CategoryName: "Support", ReplyCount: 3,
			CreatedAt:    t("2026-02-02T12:00:00Z"),
			FirstReplyAt: ptr(t("2026-02-02T18:00:00Z")),
			ResolvedAt:   ptr(t("2026-02-06T12:00:00Z")),
			Outcome:      "solved", TopicURL: topicURL(975)},

		// ---------------------------------------------------------------
		// Resolved — self-closed — 6 topics
		// ---------------------------------------------------------------

		{ID: 1012, Title: "Webhook signature validation mismatch",
			Tags: []string{"webhooks"}, CategoryName: "Support",
			CreatedAt:  t("2026-02-23T12:00:00Z"),
			ResolvedAt: ptr(t("2026-02-26T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1012)},

		{ID: 1022, Title: "Editor loses content on browser back navigation",
			Tags: []string{"editor"}, CategoryName: "Bug Reports",
			CreatedAt:  t("2026-02-27T12:00:00Z"),
			ResolvedAt: ptr(t("2026-03-05T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1022)},

		{ID: 1030, Title: "Two-factor authentication codes rejected intermittently",
			Tags: []string{"authentication"}, CategoryName: "Support",
			CreatedAt:  t("2026-03-04T12:00:00Z"),
			ResolvedAt: ptr(t("2026-03-08T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1030)},

		{ID: 1038, Title: "Search filters ignore category parameter",
			Tags: []string{"search"}, CategoryName: "Bug Reports",
			CreatedAt:  t("2026-03-09T12:00:00Z"),
			ResolvedAt: ptr(t("2026-03-14T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1038)},

		{ID: 1049, Title: "API key rotation does not invalidate old keys",
			Tags: []string{"api", "authentication"}, CategoryName: "Support",
			CreatedAt:  t("2026-03-14T12:00:00Z"),
			ResolvedAt: ptr(t("2026-03-16T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(1049)},

		{ID: 908, Title: "Permalink resolver returns 404 for archived categories",
			Tags: []string{"permalinks"}, CategoryName: "Bug Reports",
			CreatedAt:  t("2025-02-12T12:00:00Z"),
			ResolvedAt: ptr(t("2025-02-22T12:00:00Z")),
			Outcome:    "self-closed", TopicURL: topicURL(908)},

		// ---------------------------------------------------------------
		// Replied-open (tagged) — 7 topics
		// ---------------------------------------------------------------

		{ID: 1060, Title: "Intermittent 502 errors on API gateway after load balancer change",
			Tags: []string{"api"}, CategoryName: "Support", ReplyCount: 3,
			CreatedAt:      t("2026-02-12T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-12T16:00:00Z")),
			LastActivityAt: ptr(t("2026-02-25T12:00:00Z")),
			TopicURL:       topicURL(1060)},

		{ID: 1065, Title: "SSO session not persisting across subdomains",
			Tags: []string{"authentication", "sso"}, CategoryName: "Support", ReplyCount: 5,
			CreatedAt:      t("2026-02-17T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-17T14:00:00Z")),
			LastActivityAt: ptr(t("2026-03-01T12:00:00Z")),
			TopicURL:       topicURL(1065)},

		{ID: 1071, Title: "Webhook delivery order not guaranteed for batch events",
			Tags: []string{"webhooks"}, CategoryName: "Support", ReplyCount: 2,
			CreatedAt:      t("2026-02-22T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-22T18:00:00Z")),
			LastActivityAt: ptr(t("2026-03-03T12:00:00Z")),
			TopicURL:       topicURL(1071)},

		{ID: 1075, Title: "Search autocomplete suggestions lag behind index updates",
			Tags: []string{"search"}, CategoryName: "Support", ReplyCount: 4,
			CreatedAt:      t("2026-02-27T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-28T00:00:00Z")),
			LastActivityAt: ptr(t("2026-03-09T12:00:00Z")),
			TopicURL:       topicURL(1075)},

		{ID: 1080, Title: "Email bounces not updating user suppression list",
			Tags: []string{"email"}, CategoryName: "Support", ReplyCount: 1,
			CreatedAt:      t("2026-03-01T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-03-01T20:00:00Z")),
			LastActivityAt: ptr(t("2026-03-14T12:00:00Z")),
			TopicURL:       topicURL(1080)},

		{ID: 1083, Title: "Plugin compatibility issue after core update — admin notified",
			Tags: []string{"installation", "closed"}, CategoryName: "Support", ReplyCount: 3,
			CreatedAt:      t("2026-02-19T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-02-19T13:00:00Z")),
			LastActivityAt: ptr(t("2026-02-27T12:00:00Z")),
			TopicURL:       topicURL(1083)},

		{ID: 1087, Title: "Editor font rendering inconsistent on HiDPI displays",
			Tags: []string{"editor"}, CategoryName: "Bug Reports", ReplyCount: 2,
			CreatedAt:      t("2026-03-04T12:00:00Z"),
			FirstReplyAt:   ptr(t("2026-03-05T12:00:00Z")),
			LastActivityAt: ptr(t("2026-03-16T12:00:00Z")),
			TopicURL:       topicURL(1087)},
	}
}
