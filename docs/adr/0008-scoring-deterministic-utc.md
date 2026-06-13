# ADR-0008: Scoring model, penalties, frozen rules, Ukraine time

**Status:** Accepted · **Date:** 2026-06-13

## Context
Scoring fairness is the core of the pool. The method is carried over from the WC2022 sheet (recovered by reverse-engineering its cumulative points column) and clarified by the owner for the knockout stage. Two recurring failure modes must be avoided: rules argued *after* matches are played, and prediction locks firing at the wrong moment due to time-zone confusion.

## Decision

### Scoring model
A prediction is a **regular-time score** (`home_pred : away_pred`, i.e. "основний час"). Points per match:
- **Exact regular-time score → 3 points.**
- Else **correct outcome only** (W/D/L of the regular-time scoreline) → **1 point.**
- Else **0**.

**Knockout matches add a winner pick.** Because a knockout tie can be drawn in regular time but still has a team that advances (via extra time / penalties), the player additionally predicts **which team advances**. Correct winner pick → **+1 point**, stacked on top of the score points above.
- Example: predict **1:1 regular time** and **France advances** → 3 (exact reg-time score) **+ 1** (correct advancer) = **4 points**.
- **Penalties/extra time count** only for resolving the advancer (the `winner_pick`), never as part of the predicted score.
- Group-stage matches have no winner pick (draws are final) → max 3 points. Knockout matches → max 4 points.

**Optional tournament bonus picks.** Each is independently toggleable in `bonus_rules` (all default off, integer points): **champion** (default 5), **finalist** (reaches the final), **top scorer**. Each enabled kind lets a player make one pick (`tournament_picks`), locked at an admin-set deadline. Champion/finalist picks are team IDs; the top-scorer pick is a player reference, admin-confirmed at tournament end (from the FIFA top-scorer list). Correct pick → fixed integer bonus added to the total. None of this touches per-match math — all integers, no fractions.

Values live in `scoring_rules` (defaults exact=3, outcome=1, knockout_winner=1, champion_bonus off/5) and are versioned. **Rules are frozen and confirmed by all players before scoring starts.** Re-scoring is **idempotent** (`recompute-scores` CLI) and re-runs on any result change.

### Time policy
- **All timestamps stored and compared in UTC.** The prediction lock is enforced **server-side** against `kickoff_at` (UTC); the client clock is never trusted.
- **Display and deadline presentation use Ukraine time (`Europe/Kyiv`).** Conversion happens only at the presentation layer.

## Consequences
- No "the rules were different" disputes; penalty/ET handling is explicit; knockout draws are scoreable.
- The prediction form must expose a winner pick on knockout matches (see data model `predictions.winner_pick`).
- No off-by-one-hour lock failures; everyone sees Kyiv time.
- Backfill of matches already played since 2026-06-11 is a separate decision (architecture Open Questions) — scoring likely starts from an agreed matchday.

## Alternatives considered
- **Score knockouts purely on the 90'/120' line, ignore the shootout** — rejected: the owner wants the advancer to count (+1), which is what makes a predicted draw still meaningful in a knockout.
- **Add a goal-difference tier** — not part of the WC2022 method; rejected.
- **Store local Kyiv times in the DB** — invites TZ/DST bugs; rejected in favor of UTC-only storage with Kyiv display.
