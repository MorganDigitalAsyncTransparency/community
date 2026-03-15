export interface Topic {
  id: number;
  title: string;
  createdAt: string;
  tags: string[];
  category: string;
  replyCount: number;
}

export interface DashboardData {
  unrepliedTopics: Topic[];
  untaggedTopics: Topic[];
  lastSyncedAt: string;
}

function daysAgo(days: number): string {
  return new Date(Date.now() - days * 86_400_000).toISOString();
}

function hoursAgo(hours: number): string {
  return new Date(Date.now() - hours * 3_600_000).toISOString();
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

export const MOCK_DATA: DashboardData = {
  unrepliedTopics,
  untaggedTopics,
  lastSyncedAt: hoursAgo(2),
};
