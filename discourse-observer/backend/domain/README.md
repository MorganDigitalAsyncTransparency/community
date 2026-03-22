# Domain

Pure calculation functions implementing domain aggregates for the API.

## Responsibility

- Median calculation (truncated for even-length inputs)
- Time period and tag filtering
- Time bucketing (daily/weekly granularity, gap filling)
- Queue summary (unreplied count, untagged count, oldest age)
- Stalled topic detection (threshold lookup, closedTag exclusion)
- Response metrics summary (median reply/resolution, outcome counts, answer rate)
- Response time distribution (histogram bucketing by configured ceilings)
- Volume and median trend bucketing
- Tag rankings (volume, resolution time, backlog)
- Weekly backlog trend
- SLO violation detection and compliance computation
- Peak activity heatmap (7x24 UTC grid)
- Triage time analysis (median duration from creation to first tag)
- Tag flow analysis (transitions, co-occurring pairs, instability)
- Tag configuration resolution (merging per-tag overrides with defaults)

## Does not

- Handle HTTP concerns — those live in `backend/api/`
- Perform I/O (no file access, no network calls)
- Import any package other than `backend/model/` and the standard library

## Dependencies

- `backend/model/` — Topic and config types
