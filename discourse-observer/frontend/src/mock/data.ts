export interface Topic {
  id: number;
  title: string;
  createdAt: string;
  tags: string[];
  category: string;
  replyCount: number;
  firstReplyAt?: string;
  resolvedAt?: string;
  outcome?: "solved" | "self-closed";
}

export interface DashboardData {
  unrepliedTopics: Topic[];
  untaggedTopics: Topic[];
  resolvedTopics: Topic[];
  lastSyncedAt: string;
}

function daysAgo(days: number): string {
  return new Date(Date.now() - days * 86_400_000).toISOString();
}

function hoursAgo(hours: number): string {
  return new Date(Date.now() - hours * 3_600_000).toISOString();
}

function hoursAfter(base: string, hours: number): string {
  return new Date(new Date(base).getTime() + hours * 3_600_000).toISOString();
}

function daysAfter(base: string, days: number): string {
  return new Date(new Date(base).getTime() + days * 86_400_000).toISOString();
}

const unrepliedTopics: Topic[] = [
  {
    id: 1041,
    title: "Cannot authenticate with API key after upgrade",
    createdAt: daysAgo(14),
    tags: ["authentication"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1055,
    title: "Installation fails on Ubuntu 24.04 — missing libssl dependency",
    createdAt: daysAgo(11),
    tags: ["installation"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1063,
    title: "SSO login redirect loop with SAML provider",
    createdAt: daysAgo(9),
    tags: ["authentication", "sso"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1078,
    title: "Webhook delivery silently drops events over 1 MB",
    createdAt: daysAgo(7),
    tags: ["webhooks"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1089,
    title: "Search index not rebuilding after plugin update",
    createdAt: daysAgo(5),
    tags: ["search"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1094,
    title: "Email notifications delayed by several hours",
    createdAt: daysAgo(4),
    tags: ["email"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1102,
    title: "Rate limiting returns 429 even below documented threshold",
    createdAt: daysAgo(3),
    tags: ["api"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1108,
    title: "Bulk import fails with timeout on large CSV",
    createdAt: daysAgo(2),
    tags: ["data-import"],
    category: "Support",
    replyCount: 0,
  },
  {
    id: 1115,
    title: "Markdown rendering broken in post preview",
    createdAt: daysAgo(1),
    tags: ["editor"],
    category: "Support",
    replyCount: 0,
  },
];

const untaggedTopics: Topic[] = [
  {
    id: 1044,
    title: "How do I change the default theme colors?",
    createdAt: daysAgo(12),
    tags: [],
    category: "General",
    replyCount: 3,
  },
  {
    id: 1067,
    title: "Sidebar navigation disappeared after update",
    createdAt: daysAgo(8),
    tags: [],
    category: "Bug Reports",
    replyCount: 1,
  },
  {
    id: 1085,
    title: "Best practices for category permissions?",
    createdAt: daysAgo(6),
    tags: [],
    category: "General",
    replyCount: 5,
  },
  {
    id: 1098,
    title: "Mobile layout issues on Galaxy S24",
    createdAt: daysAgo(3),
    tags: [],
    category: "Bug Reports",
    replyCount: 2,
  },
  {
    id: 1112,
    title: "User group sync with external directory not working",
    createdAt: daysAgo(1),
    tags: [],
    category: "Support",
    replyCount: 0,
  },
];

const resolvedTopics: Topic[] = [
  {
    id: 1001,
    title: "API rate limit not resetting after cooldown period",
    createdAt: daysAgo(28),
    tags: ["api"],
    category: "Support",
    replyCount: 4,
    firstReplyAt: hoursAfter(daysAgo(28), 3),
    resolvedAt: daysAfter(daysAgo(28), 2),
    outcome: "solved",
  },
  {
    id: 1005,
    title: "OAuth2 token refresh fails silently",
    createdAt: daysAgo(26),
    tags: ["authentication"],
    category: "Support",
    replyCount: 6,
    firstReplyAt: hoursAfter(daysAgo(26), 8),
    resolvedAt: daysAfter(daysAgo(26), 5),
    outcome: "solved",
  },
  {
    id: 1008,
    title: "Docker setup crashes on Apple Silicon",
    createdAt: daysAgo(25),
    tags: ["installation"],
    category: "Support",
    replyCount: 3,
    firstReplyAt: hoursAfter(daysAgo(25), 1),
    resolvedAt: daysAfter(daysAgo(25), 1),
    outcome: "solved",
  },
  {
    id: 1012,
    title: "Webhook signature validation mismatch",
    createdAt: daysAgo(24),
    tags: ["webhooks"],
    category: "Support",
    replyCount: 0,
    resolvedAt: daysAfter(daysAgo(24), 3),
    outcome: "self-closed",
  },
  {
    id: 1016,
    title: "Email digest contains duplicate entries",
    createdAt: daysAgo(22),
    tags: ["email"],
    category: "Bug Reports",
    replyCount: 5,
    firstReplyAt: hoursAfter(daysAgo(22), 12),
    resolvedAt: daysAfter(daysAgo(22), 4),
    outcome: "solved",
  },
  {
    id: 1019,
    title: "Full-text search returns stale results",
    createdAt: daysAgo(21),
    tags: ["search"],
    category: "Support",
    replyCount: 2,
    firstReplyAt: hoursAfter(daysAgo(21), 24),
    resolvedAt: daysAfter(daysAgo(21), 7),
    outcome: "solved",
  },
  {
    id: 1022,
    title: "Editor loses content on browser back navigation",
    createdAt: daysAgo(20),
    tags: ["editor"],
    category: "Bug Reports",
    replyCount: 0,
    resolvedAt: daysAfter(daysAgo(20), 6),
    outcome: "self-closed",
  },
  {
    id: 1025,
    title: "Initial setup wizard skips database migration step",
    createdAt: daysAgo(18),
    tags: ["setup"],
    category: "Support",
    replyCount: 7,
    firstReplyAt: hoursAfter(daysAgo(18), 2),
    resolvedAt: daysAfter(daysAgo(18), 3),
    outcome: "solved",
  },
  {
    id: 1028,
    title: "API pagination returns wrong total count",
    createdAt: daysAgo(17),
    tags: ["api"],
    category: "Bug Reports",
    replyCount: 3,
    firstReplyAt: hoursAfter(daysAgo(17), 6),
    resolvedAt: daysAfter(daysAgo(17), 2),
    outcome: "solved",
  },
  {
    id: 1030,
    title: "Two-factor authentication codes rejected intermittently",
    createdAt: daysAgo(15),
    tags: ["authentication"],
    category: "Support",
    replyCount: 0,
    resolvedAt: daysAfter(daysAgo(15), 4),
    outcome: "self-closed",
  },
  {
    id: 1033,
    title: "Webhook retry logic sends duplicate payloads",
    createdAt: daysAgo(14),
    tags: ["webhooks"],
    category: "Support",
    replyCount: 4,
    firstReplyAt: hoursAfter(daysAgo(14), 18),
    resolvedAt: daysAfter(daysAgo(14), 6),
    outcome: "solved",
  },
  {
    id: 1036,
    title: "Email templates not rendering HTML correctly",
    createdAt: daysAgo(12),
    tags: ["email"],
    category: "Support",
    replyCount: 2,
    firstReplyAt: hoursAfter(daysAgo(12), 36),
    resolvedAt: daysAfter(daysAgo(12), 8),
    outcome: "solved",
  },
  {
    id: 1038,
    title: "Search filters ignore category parameter",
    createdAt: daysAgo(10),
    tags: ["search"],
    category: "Bug Reports",
    replyCount: 0,
    resolvedAt: daysAfter(daysAgo(10), 5),
    outcome: "self-closed",
  },
  {
    id: 1042,
    title: "Plugin installation fails with permission error",
    createdAt: daysAgo(8),
    tags: ["installation", "setup"],
    category: "Support",
    replyCount: 3,
    firstReplyAt: hoursAfter(daysAgo(8), 4),
    resolvedAt: daysAfter(daysAgo(8), 1),
    outcome: "solved",
  },
  {
    id: 1046,
    title: "Editor toolbar disappears in fullscreen mode",
    createdAt: daysAgo(6),
    tags: ["editor"],
    category: "Bug Reports",
    replyCount: 1,
    firstReplyAt: hoursAfter(daysAgo(6), 48),
    resolvedAt: daysAfter(daysAgo(6), 3),
    outcome: "solved",
  },
  {
    id: 1049,
    title: "API key rotation does not invalidate old keys",
    createdAt: daysAgo(5),
    tags: ["api", "authentication"],
    category: "Support",
    replyCount: 0,
    resolvedAt: daysAfter(daysAgo(5), 2),
    outcome: "self-closed",
  },
  {
    id: 1052,
    title: "Webhook endpoint health check returns false positive",
    createdAt: daysAgo(3),
    tags: ["webhooks"],
    category: "Support",
    replyCount: 5,
    firstReplyAt: hoursAfter(daysAgo(3), 2),
    resolvedAt: daysAfter(daysAgo(3), 1),
    outcome: "solved",
  },
];

export const MOCK_DATA: DashboardData = {
  unrepliedTopics,
  untaggedTopics,
  resolvedTopics,
  lastSyncedAt: hoursAgo(2),
};
