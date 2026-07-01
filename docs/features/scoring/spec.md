# Spec — Scoring (CAP-02)

Behavioral spec for how a prediction is scored against a result. The core is a
**pure, deterministic** function (NFR-001); these scenarios are the contract its
unit tests and golden-fixture evals trace to.

Owns: FR-010, FR-011, FR-012, FR-013, FR-014, FR-015.
Decision record: [ADR-0008](../../adr/0008-scoring-deterministic-utc.md).

## Group / regulation outcomes

### FR-010 — exact score
- **Given** a finished fixture `2:1`
- **When** the player predicted `2:1`
- **Then** the prediction scores **3**.

### FR-011 — correct outcome, wrong scoreline
- **Given** a finished fixture `2:1`
- **When** the player predicted `3:1` (home win, wrong score) — or `1:1` vs an actual `2:2` draw
- **Then** the prediction scores **1**.

### FR-012 — wrong outcome
- **Given** a finished fixture `0:1`
- **When** the player predicted `1:0`
- **Then** the prediction scores **0**.

## Knockout advancer (+1)

### FR-013 — advancer awarded
- **Given** a knockout fixture whose **regulation** result is a draw `1:1`, decided for team A on penalties
- **When** the player predicted `1:1` **and** picked **A** to advance
- **Then** the prediction scores **3 (exact) + 1 (advancer) = 4**.

### FR-013 — advancer NOT awarded on a decisive scoreline
- **Given** a knockout fixture that ends `2:1` (decisive in regulation)
- **When** the player predicted `2:1`
- **Then** the prediction scores **3** with **no** separate +1 (the winner was implied by the scoreline).

### FR-013 — advancer NOT awarded when regulation was not a draw
- **Given** a knockout fixture `0:1`
- **When** the player predicted `1:1` and picked the home team to advance
- **Then** the prediction scores **0** — the +1 applies only when both the prediction and the actual regulation result are draws.

## Bonuses (time-tiered)

> Note: the pure `Score()` function covers per-match points (FR-010–013, 015).
> The bonus **award/tier resolution** (FR-014) runs after the relevant result in
> the `internal/resolve` package and is traced by `resolve_test.go` — it is
> listed here because it is conceptually part of scoring, but it is verified
> where it is implemented.

### FR-014 — bonus only if correct
- **Given** the tournament has a known champion
- **When** a player picked that champion during the group stage (early tier)
- **Then** the player is awarded the higher champion tier; a wrong pick is awarded **0**.

## Recompute

### FR-015 — deterministic recompute
- **Given** a fixture result is set or later overridden by an admin
- **When** scores are recomputed
- **Then** every affected prediction's points are re-derived purely from (prediction, result) with no dependence on wall-clock or evaluation order — identical inputs always yield identical points.
