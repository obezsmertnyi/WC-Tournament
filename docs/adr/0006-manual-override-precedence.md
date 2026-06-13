# ADR-0006: API is the source of truth; manual entry is an outage fallback

**Status:** Accepted · **Date:** 2026-06-13

## Context
Match results come from the FIFA API (ADR-0002). The owner clarified the role of manual admin entry: it is needed **only when the API is unavailable but a match still needs to be scored** — not as a routine override. Among friends, an incorrect or stuck result that shifts the leaderboard is a source of disputes, so changes must be transparent.

## Decision
- **The provider (FIFA, then fallbacks) is the source of truth** for results.
- **Manual admin entry is a fallback** used when the API is unavailable but scoring must proceed. Such a row is marked `result_source = 'manual'`.
- **Reconciliation on API recovery:** when a provider later returns a result for a match that was filled manually, the system compares them. If they match, the source is upgraded to the provider silently. **If they differ, the discrepancy is logged to the audit feed and the provider (authoritative) value is adopted and re-scored** — so a stopgap entry can't permanently diverge from reality. An admin may explicitly **lock** a row (e.g. a genuine API error) to keep a manual value; locking is itself an audited action.
- Every result write — provider, manual, or reconciliation — appends an immutable `audit_log` row (who/what/when, old→new, ADR-0009).
- Any change to an already-`finished` match **triggers an idempotent re-score** of affected predictions.

## Consequences
- Scoring never blocks on an API outage — the admin fills in, the game continues.
- The authoritative API result wins once available, so manual stopgaps self-correct; mismatches are visible, not silent.
- The explicit lock is the escape hatch for the rare case the API is genuinely wrong.
- Re-scoring on result change prevents "stuck" points after corrections.

## Alternatives considered
- **Manual always wins, never overwritten** (earlier draft) — contradicts the owner's intent that manual is only an outage stopgap; would let a hurried entry diverge from the official result. Rejected.
- **Provider always wins, no manual path** — leaves scoring blocked during an API outage. Rejected.
