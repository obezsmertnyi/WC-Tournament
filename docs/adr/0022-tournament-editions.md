# ADR-0022: Multi-edition support — a first-class "tournament" entity

**Status:** Accepted · **Date:** 2026-07-07 · **Related:** ADR-0006, ADR-0008, ADR-0012

## Context
The pool runs every 4 years (WC 2026, 2030, 2034, …). The owner wants, at the
2030 edition, to **archive 2026 read-only (still browsable)** and **start fresh
for 2030** with the same group of friends. The app is currently implicitly
single-tournament: nothing scopes teams, matches, predictions, points or bonuses
to an edition. We need editions to coexist in one app so history is preserved and
browsable, accounts are shared, and an all-time view is possible later.

## Decision
Introduce a first-class **`tournaments`** (edition) entity and scope all
per-edition data to it. Exactly one edition is **active** (accepts play); past
editions are **read-only and browsable** via an edition switcher.

**Shared across editions (NOT scoped):** `users` (the same friends), and truly
global settings.

**Per-edition (scoped by `tournament_id`):** `teams`, `matches`, `predictions`,
`points`, `tournament_picks`, `bonus_rules`, `champion_tiers`, edition-specific
`audit_log` rows, and edition-specific `app_state` flags (namespaced per edition).

### Schema
```
tournaments(id, code UNIQUE, name, year, is_active, created_at, updated_at)
  -- partial unique index: at most one is_active = true
  -- e.g. WC2026 (archived), WC2030 (active)
```
- Add `tournament_id bigint NOT NULL REFERENCES tournaments(id)` to each scoped
  table.
- Uniqueness that assumed a single tournament gets edition-qualified:
  `matches (tournament_id, fifa_id)`, `teams (tournament_id, fifa_id)`,
  `tournament_picks (tournament_id, user_id, kind)`. `predictions`/`points`
  keep `(user_id, match_id)` — `match_id` already belongs to exactly one edition.
- The **active** edition = `tournaments.is_active`. Reads for "current play"
  filter on it; an optional `?tournament=<code>` selects a past edition
  (read-only).

### Behavior
- **Write-guard:** predict / set-bonus / admin-result reject when the target
  edition is not active (409 "edition archived") — same shape as the bonus lock.
- **All edition management lives in the admin panel** (admin-only, audited — never
  CLI or manual DB). New admin endpoints under `/api/admin/tournaments`:
  - `GET` — list editions + which is active (admin console table).
  - `POST` — create an edition (`code`, `name`, `year`), created inactive.
  - `POST /:id/load-fixtures` — run a tournament-scoped FIFA sync to populate that
    edition's teams + fixtures (reuses the existing sync, scoped to the edition).
  - `POST /:id/activate` — make it the active edition (atomically archives the
    current one); guarded by a **confirmation** in the UI since it flips who can
    play. Audited as `admin_tournament_activate`.
  The admin UI gets a "Tournaments" section: list, create, load fixtures, and
  activate/archive — the whole 2030 rollout is done by the owner from the panel.
- **Archive view (all users):** a read-only **edition switcher** lets anyone
  browse a past edition's bracket, standings, leaderboard, and every player's
  (now-revealed) predictions and bonuses. Browsing is public; management is admin.
- **Optional later:** an all-time "hall of fame" aggregating results across
  editions per user (the scoping makes this a straight cross-edition query).

### Migration (backfill, no behavior change on landing)
1. Create `tournaments`; insert `WC2026` as `is_active = true`.
2. Add `tournament_id` nullable to each scoped table; backfill every existing row
   to the WC2026 id; then `SET NOT NULL` + add FK.
3. Swap the edition-qualified unique constraints.
Single active edition = WC2026 ⇒ every current query behaves exactly as before.

## Consequences
- True archive with full browsing, shared accounts, and a clean 2030 slate in one
  deployment; cross-edition features become possible.
- Every edition-scoped read/write must filter by edition — a broad but mechanical
  change; the active-edition default keeps call sites simple.
- Larger migration than the alternatives, done once.

## Alternatives considered
- **Freeze 2026 read-only + a separate 2030 instance.** Least work, zero risk to
  2026, but two deployments, no shared accounts/leaderboard, archive = the old
  site. Rejected for the unified-history goal (viable fallback if effort must be
  minimal).
- **Snapshot-archive + reset (wipe live tables, load 2030).** Keeps the live
  schema single-tournament, but the archive is a frozen copy needing its own
  viewer, the wipe is risky, and cross-edition features are hard. Rejected.
