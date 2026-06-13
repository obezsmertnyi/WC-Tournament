# ADR-0010: One-time backfill window for the already-played opening matches

**Status:** Accepted · **Date:** 2026-06-13

## Context
The app is being built after the tournament already started (2026-06-11). The owner wants scoring to count **from the very first match**, but players cannot have submitted predictions for 11–13.06 because they hadn't registered yet. This conflicts with the normal invariant that predictions lock at kickoff (ADR-0008).

## Decision
A **one-time, admin-controlled backfill window** allows predictions to be recorded for matches that have already kicked off, **only** for the opening matchdays up to the agreed "live" cutoff. After the cutoff, the normal server-side kickoff lock applies to everyone.

- Scoring counts **from match #1** (2026-06-11).
- Backfilled predictions are entered either **by the admin on a player's behalf** (while players are still registering) or **by the players themselves** once registered — whichever is convenient.
- **Every backfilled prediction is written to the public audit log** (ADR-0009), flagged as backfill, with actor and timestamp — so it's transparent who entered what and when.
- This is **trust-based**: among friends, with the result already public, honesty is assumed and the audit trail is the safeguard, not a technical lock.
- The window is closed explicitly by the admin once everyone is caught up; from then on backfill is disabled.

## Consequences
- Scoring can include the opening matches without excluding late-registering players.
- Full transparency via the audit feed; any dispute is traceable.
- Accepted risk: backfilled predictions are entered after results are known — acceptable for a private, trust-based friends pool, not for anything competitive/public.
- The backfill flag + window state must be modeled (a `backfill_until` config and an `is_backfill` marker on the audit/prediction).

## Alternatives considered
- **Start scoring from the first not-yet-played match** (drop 11–13.06) — simplest and cheat-proof, but the owner wants the opening matches to count. Not chosen.
- **Open backfill to players with no audit** — removes the only safeguard. Rejected.
