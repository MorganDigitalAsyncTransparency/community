# API — Traceability Matrix

This matrix shows how use cases decompose into API contract requirements and verification artifacts.

---

## Queue

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-1: Identify topics waiting longest | [api-contract.md](api-contract.md) | AC-12, AC-13 | [api-contract_verification.md](api-contract_verification.md) |
| UC-2: See all unreplied topics | [api-contract.md](api-contract.md) | AC-12, AC-13 | [api-contract_verification.md](api-contract_verification.md) |
| UC-3: Detect untagged topics | [api-contract.md](api-contract.md) | AC-12, AC-14 | [api-contract_verification.md](api-contract_verification.md) |
| UC-18: Detect stalled topics | [api-contract.md](api-contract.md) | AC-15 | [api-contract_verification.md](api-contract_verification.md) |

---

## Response metrics

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-4: Time to first reply | [api-contract.md](api-contract.md) | AC-16 | [api-contract_verification.md](api-contract_verification.md) |
| UC-5: Time to resolution | [api-contract.md](api-contract.md) | AC-16 | [api-contract_verification.md](api-contract_verification.md) |
| UC-6: Solved vs self-closed | [api-contract.md](api-contract.md) | AC-16 | [api-contract_verification.md](api-contract_verification.md) |
| UC-7: Answer rate | [api-contract.md](api-contract.md) | AC-16 | [api-contract_verification.md](api-contract_verification.md) |
| UC-8: Response time trends | [api-contract.md](api-contract.md) | AC-18 | [api-contract_verification.md](api-contract_verification.md) |
| UC-17: Topic intake volume | [api-contract.md](api-contract.md) | AC-17 | [api-contract_verification.md](api-contract_verification.md) |
| UC-20: Response time distribution | [api-contract.md](api-contract.md) | AC-19 | [api-contract_verification.md](api-contract_verification.md) |

---

## Distribution

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-9: Highest-volume tags | [api-contract.md](api-contract.md) | AC-20 | [api-contract_verification.md](api-contract_verification.md) |
| UC-10: Slowest tags | [api-contract.md](api-contract.md) | AC-21 | [api-contract_verification.md](api-contract_verification.md) |
| UC-11: Accumulating backlogs | [api-contract.md](api-contract.md) | AC-22, AC-23 | [api-contract_verification.md](api-contract_verification.md) |

---

## SLO

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-13: Flag topics exceeding SLO | [api-contract.md](api-contract.md) | AC-24 | [api-contract_verification.md](api-contract_verification.md) |
| UC-14: Evaluate SLO compliance | [api-contract.md](api-contract.md) | AC-25 | [api-contract_verification.md](api-contract_verification.md) |

---

## Activity

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-19: Peak activity periods | [api-contract.md](api-contract.md) | AC-26 | [api-contract_verification.md](api-contract_verification.md) |

---

## Workflow insights

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| Triage responsiveness | [triage-time.md](triage-time.md) | TT-1 through TT-13 | `backend/domain/triage_unit_test.go`, `backend/api/triage-time_contract_test.go` |
| Tag flow patterns | [tag-flows.md](tag-flows.md) | TF-1 through TF-21 | `backend/domain/tagflows_unit_test.go`, `backend/api/tag-flows_contract_test.go` |
| Escalation patterns | [escalations.md](escalations.md) | EP-1 through EP-12 | `backend/domain/escalations_unit_test.go`, `backend/api/escalations_contract_test.go` |

---

## Filtering and navigation

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| UC-12: Filter by time period | [api-contract.md](api-contract.md) | AC-8, AC-9, AC-11, AC-31 | [api-contract_verification.md](api-contract_verification.md) |
| UC-15: Filter by tag | [api-contract.md](api-contract.md) | AC-10, AC-30 | [api-contract_verification.md](api-contract_verification.md) |
| UC-16: Navigate tags by area | [api-contract.md](api-contract.md) | AC-27 | [api-contract_verification.md](api-contract_verification.md) |
| UC-24: URL state persistence | (frontend-only) | — | — |

---

## Infrastructure

| Use case | Spec | Requirements | Verification |
|----------|------|-------------|--------------|
| (operational) | [api-contract.md](api-contract.md) | AC-28 (status) | [api-contract_verification.md](api-contract_verification.md) |
| (operational) | [api-contract.md](api-contract.md) | AC-33 (sync log) | `GET /api/v1/sync-log` returns entries |

---

## Gaps

None. All use cases except UC-24 (frontend-only) are covered by API contract requirements.
