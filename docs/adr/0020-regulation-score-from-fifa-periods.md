# ADR-0020: Derive the knockout regulation scoreline from FIFA goal periods

**Status:** Accepted · **Date:** 2026-07-04 · **Refines:** ADR-0006, ADR-0008

## Context
Knockout scoring is defined against the **90-minute (regulation) result**: an
exact regulation score is worth 3 points, the correct outcome 1, and the "+1
advancer" bonus is awarded **only when both the prediction and the actual
regulation result are draws** (ADR-0008, `docs/features/scoring/spec.md`).

The FIFA calendar feed (`fifa_types.go`) reports only the **final aet-inclusive
score** (`HomeTeamScore`/`AwayTeamScore`) plus a separate penalty score — there
is **no regulation-90 field**. A knockout that was a regulation *draw* but won by
an **extra-time goal** is therefore stored as a decisive score (e.g. `3:2`),
which the regulation-based scoring reads as decisive — silently robbing correct
`2:2`-draw + advancer predictions and, symmetrically, handing a `+1` to everyone
when it should be a decisive result. This bit us twice: BEL–SEN (#81) and
ARG–CPV (#87), each patched by a **manual override** (ADR-0006) that reset the
score to the regulation value and left `winner_team_id` = advancer.

The manual override is a per-match band-aid that requires an admin to notice the
mis-score. We need the score to be **correct by construction**.

The signal exists: the FIFA **live** endpoint tags every goal event with a
`Period` code. Verified against the raw API for ARG 3:2 CPV
(`live/football/17/285023/289287/400021521`): Messi **P3** + two extra-time
goals **P7/P9** for Argentina; Duarte **P5** + one extra-time goal **P7** for
Cabo Verde → regulation = **1:1**. Codes: **3** = first half, **5** = second half
(regulation); **7/9** = extra-time halves; **10** = end state. `Minute` is *not*
a usable proxy — stoppage time means regulation can read as "102'".

## Decision
- Add `Period` to the parsed goal event (`results.LiveGoal`).
- Add `results.RegulationScore(goals, finalHome, finalAway) → (h, a, ok)`:
  the regulation scoreline is the count of goals per side with **Period ∈ {3,5}**.
- **Completeness safeguard.** `ok` is true only when every goal carries a
  `Period` and a known side, **and** the goals in periods `{3,5,7,9}` sum exactly
  to the stored aet score. When `ok` is false the caller keeps the stored score —
  many finished matches carry a winner but no goal timeline, and we must never
  zero a real scoreline from missing data.
- **Where it runs.** `winners.Run` already fetches the live detail per finished
  knockout to resolve `winner_team_id`; in the *same* pass it now derives the
  regulation score and, when it differs from the stored score, writes it via
  `CorrectKnockoutRegulationScore`. `winner_team_id` still records the actual
  advancer, so the "+1" bonus is unaffected.
- **Persistence guard.** A corrected row is marked `result_source =
  'fifa_regulation'`; `UpsertMatch` will not overwrite `manual` **or**
  `fifa_regulation` rows, so the next calendar sync (which carries the aet score)
  cannot undo the correction. A manual admin override still wins over everything
  (`CorrectKnockoutRegulationScore` never touches a `manual` row) — ADR-0006 holds.

## Consequences
- Extra-time-goal knockouts self-correct to their regulation scoreline the moment
  the advancer is resolved; **no manual override needed** for the common case.
- Penalty shootouts are handled uniformly: regulation `1:1`, aet `1:1` → no change
  (the common case); the rare shootout *with* extra-time goals (aet `2:2`,
  regulation `1:1`) is now also recovered correctly, which the old feed could not.
- The `manual` override remains the escape hatch for a straggler whose goal
  timeline was incomplete at resolution time (safeguard `ok=false`), and for any
  genuine FIFA data error.
- Adds one distinct `result_source` value (`fifa_regulation`); it is backend-only
  (not surfaced in the frontend).

## Limitation (accepted)
`winners.Run` only processes knockouts with `winner_team_id IS NULL`, so the
correction gets **one shot** — at advancer resolution. If the goal timeline is
incomplete at that instant (safeguard fails), the winner is still set and the
match drops out of the queue; it will not be re-corrected automatically and needs
a manual override. In practice the live feed's goal events are complete once the
match reads finished, so this is the rare tail. Documented rather than solved to
avoid re-fetching the live detail for every finished knockout on every tick.

## Alternatives considered
- **Minute-based cutoff (goals ≤ 90').** Rejected: stoppage time makes the
  90-minute boundary fuzzy (a regulation half can end at "45+7'", extra time can
  read past 102'); the owner explicitly flagged this.
- **Separate `regulation_home/away_score` columns, keep aet for display.**
  Rejected: contradicts the established #81/#87 behavior (the visible score *is*
  the regulation result) and adds scoring-source ambiguity + a migration for no
  product gain; the full aet timeline is still visible in the match detail view.
- **Keep the manual override as the only remedy.** Rejected: relies on a human
  noticing every ET-goal mis-score — the root cause, not the symptom, is fixable.
