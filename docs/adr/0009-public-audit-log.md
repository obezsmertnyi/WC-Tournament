# ADR-0009: Public, immutable audit log with timed disclosure

**Status:** Accepted · **Date:** 2026-06-13

## Context
Among friends, trust depends on transparency: everyone should be able to see who predicted what and when, and — equally important — every admin action (result overrides, rule changes, bracket edits). This must be **public**, in its own tab, and cover both player and admin activity. ADR-0006 already requires auditing admin overrides; this ADR generalizes that into a single public log.

The tension: prediction *values* must stay secret until a match locks, otherwise players copy each other. So we cannot simply make the whole log public the instant a prediction is written.

## Decision
A single **insert-only `audit_log`** table records every prediction write and every admin/system action. Rows are **never updated or deleted** (corrections are new rows). Each row carries `public_visible_at`:
- **Predictions:** `public_visible_at = match.kickoff_at`. The *fact and timestamp* ("Player X submitted/edited a prediction for match Y at 14:32") is public immediately; the *predicted score* is masked until kickoff.
- **Admin/system actions:** `public_visible_at = created_at` — public at once, with full `before_json → after_json`.

A dedicated **public "Activity / Audit" tab** renders this feed, read-only, filterable by player / match / action type. The API enforces `now() >= public_visible_at` before exposing masked fields — server-side, never trusting the client.

## Consequences
- Full transparency and dispute resolution: any leaderboard movement is traceable to a logged, timestamped action.
- Anti-cheat preserved: predicted scores cannot be seen before lock, even though the activity is otherwise public.
- Immutability means the log itself can't be quietly edited — a corrected result is a visible new entry.
- Slight extra write on every prediction/admin mutation — negligible at this scale (≤15 players, 104 matches).

## Alternatives considered
- **Admin-only audit** — fails the explicit "must be public" requirement. Rejected.
- **Fully public predictions in real time** — destroys the game (copying). Rejected.
- **Mutable log / overwrite on correction** — defeats the purpose of an audit trail. Rejected.
