# Requirements traceability matrix

> **Generated** by `scripts/gen-traceability.mjs` from `docs/requirements.md`
> + `@trace <FR-id>` annotations. Do not edit by hand — CI fails if stale
> (`node scripts/gen-traceability.mjs --check`). See [ADR-0014](../adr/0014-requirement-id-grammar.md).

Coverage: **22/24** functional requirements traced (MVP: **21/22**).

| FR | Tag | Requirement | Traced by |
|----|-----|-------------|-----------|
| FR-001 | MVP | A signed-in player can submit/edit a score prediction (home, away; 0–30) for any fixture before its kickoff. | `backend/internal/api/m2_predictions_test.go` |
| FR-002 | MVP | Predictions lock at kickoff; after kickoff a non-admin can no longer create or change a prediction. | `backend/internal/api/m2_predictions_test.go` |
| FR-003 | MVP | For a knockout fixture predicted as a **regulation draw**, the player must also pick which team advances; a decisive scoreline implies the winner and needs no separate pick. | `backend/internal/api/m2_predictions_test.go` |
| FR-010 | MVP | An exact-score prediction scores **3** points. | `backend/internal/scoring/scoring_evals_test.go`<br>`backend/internal/scoring/scoring_test.go` |
| FR-011 | MVP | A correct outcome (right winner, or a draw) with a wrong scoreline scores **1** point. | `backend/internal/scoring/scoring_evals_test.go`<br>`backend/internal/scoring/scoring_test.go` |
| FR-012 | MVP | A wrong outcome scores **0**. | `backend/internal/scoring/scoring_evals_test.go`<br>`backend/internal/scoring/scoring_test.go` |
| FR-013 | MVP | On a knockout fixture, **+1** is added **only when** the prediction was a regulation draw **and** the actual regulation result was a draw **and** the player's advancer pick matches the team that actually advanced (extra time / penalties). | `backend/internal/scoring/scoring_evals_test.go`<br>`backend/internal/scoring/scoring_test.go` |
| FR-014 | MVP | Champion/finalist/top-scorer bonuses are time-tiered and awarded **only if** the pick proves correct after the relevant result is known. | `backend/internal/resolve/resolve_test.go` |
| FR-015 | MVP | Scores are recomputed deterministically whenever a result arrives or is overridden. | `backend/internal/scoring/scoring_evals_test.go` |
| FR-030 | MVP | Players sign in with Google OAuth; an admin signs in with a password (constant-time compare); the server refuses to boot without a ≥32-byte `JWT_SECRET`. | `backend/internal/auth/jwt_test.go` |
| FR-031 | MVP | When **demo mode** is on, a non-admin's effective access is their tier: `none` (browse UI only), `ro` (also see other players' data), `rw` (also participate). | `backend/internal/auth/access_test.go`<br>`frontend/src/lib/access.test.ts` |
| FR-032 | MVP | When demo mode is off, or for an admin, effective access is always `rw` (pre-demo behavior preserved). | `backend/internal/auth/access_test.go`<br>`frontend/src/lib/access.test.ts` |
| FR-033 | MVP | New self-service Google sign-ups land in `none` while demo mode is on; admin-provisioned players are `rw`. | ⚠️ **untraced** |
| FR-070 | MVP | A read-only MCP server exposes the public pool state to agents as tools: today's/❲range❳ fixtures, group standings, leaderboard, the knockout bracket, and a named player's revealed predictions. | `mcp/evals/tools.eval.test.ts` |
| FR-071 | MVP | The MCP server never exposes secrets, never mutates state, and validates all tool inputs (rejects unknown params, clamps ranges). | `mcp/evals/tools.eval.test.ts` |
| FR-072 | Future | The MCP server surfaces a per-match "who predicted what" only after kickoff (mirrors the app's reveal lock). | `mcp/evals/tools.eval.test.ts` |
| FR-080 | MVP | For a kicked-off/finished fixture the app shows a generated, natural-language recap built from the match facts (teams, scoreline, stage, status) and, when available, who nailed the exact score. | `frontend/src/lib/recap.test.ts` |
| FR-081 | MVP | The recap is **fact-grounded**: a guardrail rejects any candidate recap that introduces a team or number not present in the match facts (no hallucination), bounds length, and neutralizes injection — whatever produced the text. | `frontend/src/lib/recap.test.ts` |
| FR-082 | Future | A pluggable LLM provider may produce the recap prose; its output passes through the FR-081 guardrail before display (never shown raw). | ⚠️ **untraced** |
| FR-090 | MVP | An authenticated user can chat with a football assistant (Gemini 2.5 Flash, Vertex AI via keyless WIF); replies stream token-by-token. | `backend/internal/gemini/gemini_test.go` |
| FR-091 | MVP | The assistant answers **only** football / FIFA World Cup topics (players, clubs, national teams, history, matches, WC2026); anything off-topic is refused in one sentence with a football redirect. A layered guardrail enforces this: input hygiene → a `flash-lite` topic-classifier gate → a hardened system instruction (anti-injection: user text is data, never reveal/obey it) → output validation. The system prompt is never disclosed. | `backend/internal/gemini/gemini_test.go` |
| FR-092 | MVP | The user can request a **club or player card**; the assistant returns a structured card (name, country, club, position, achievements, summary, confidence) rendered as a UI card, flagged "may be outdated" when confidence < high. | `backend/internal/gemini/gemini_test.go` |
| FR-093 | MVP | AI endpoints require authentication (anonymous → 401) but are **not tier-gated** — every logged-in user incl. the `none` demo tier may use them. Per-user rate limits + input-length caps apply; if the AI backend is unavailable the endpoints return `503` without affecting the rest of the app. | `backend/internal/gemini/gemini_test.go` |
| FR-100 | MVP | The assistant is **grounded in the app's own live tournament data** (ADR-0018): via Gemini function-calling it can read a one-call **tournament overview** (primary — for a quick championship summary), recent results, a team's matches, group standings, and the prediction-pool leaderboard, and answers current-tournament questions from that data — never claiming the tournament "hasn't started" or that it lacks current data. Answers are grounded in returned tool data (no invented results); a tool/grounding failure degrades gracefully to an ungrounded reply. | `backend/internal/gemini/gemini_test.go` |

