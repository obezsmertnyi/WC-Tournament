# Requirements — WC-Tournament

Numbered, testable requirements. IDs are **stable and never renumbered** (see
[ADR-0014](adr/0014-requirement-id-grammar.md)); each states **one** behavior and
is tagged **MVP** or **Future**. Every FR is owned by exactly one capability slice
in [`mvp-capability-plan.md`](mvp-capability-plan.md), cited in a capability spec
under `docs/features/`, and traced from a test via `@trace <FR-id>`.

Grammar: `FR` functional · `NFR` non-functional · `TC` technical constraint ·
`BC` business constraint.

## Functional (FR)

### Predictions — CAP-01
- **FR-001** (MVP) A signed-in player can submit/edit a score prediction (home, away; 0–30) for any fixture before its kickoff.
- **FR-002** (MVP) Predictions lock at kickoff; after kickoff a non-admin can no longer create or change a prediction.
- **FR-003** (MVP) For a knockout fixture predicted as a **regulation draw**, the player must also pick which team advances; a decisive scoreline implies the winner and needs no separate pick.

### Scoring — CAP-02
- **FR-010** (MVP) An exact-score prediction scores **3** points.
- **FR-011** (MVP) A correct outcome (right winner, or a draw) with a wrong scoreline scores **1** point.
- **FR-012** (MVP) A wrong outcome scores **0**.
- **FR-013** (MVP) On a knockout fixture, **+1** is added **only when** the prediction was a regulation draw **and** the actual regulation result was a draw **and** the player's advancer pick matches the team that actually advanced (extra time / penalties).
- **FR-014** (MVP) Champion/finalist/top-scorer bonuses are time-tiered and awarded **only if** the pick proves correct after the relevant result is known.
- **FR-015** (MVP) Scores are recomputed deterministically whenever a result arrives or is overridden.

### Auth & access — CAP-05
- **FR-030** (MVP) Players sign in with Google OAuth; an admin signs in with a password (constant-time compare); the server refuses to boot without a ≥32-byte `JWT_SECRET`.
- **FR-031** (MVP) When **demo mode** is on, a non-admin's effective access is their tier: `none` (browse UI only), `ro` (also see other players' data), `rw` (also participate).
- **FR-032** (MVP) When demo mode is off, or for an admin, effective access is always `rw` (pre-demo behavior preserved).
- **FR-033** (MVP) New self-service Google sign-ups land in `none` while demo mode is on; admin-provisioned players are `rw`.

### Tools / MCP — CAP-08 (bonus)
- **FR-070** (MVP) A read-only MCP server exposes the public pool state to agents as tools: today's/❲range❳ fixtures, group standings, leaderboard, the knockout bracket, and a named player's revealed predictions.
- **FR-071** (MVP) The MCP server never exposes secrets, never mutates state, and validates all tool inputs (rejects unknown params, clamps ranges).
- **FR-072** (Future) The MCP server surfaces a per-match "who predicted what" only after kickoff (mirrors the app's reveal lock).

### AI match recap — CAP-09 (bonus, GenUI)
- **FR-080** (MVP) For a kicked-off/finished fixture the app shows a generated, natural-language recap built from the match facts (teams, scoreline, stage, status) and, when available, who nailed the exact score.
- **FR-081** (MVP) The recap is **fact-grounded**: a guardrail rejects any candidate recap that introduces a team or number not present in the match facts (no hallucination), bounds length, and neutralizes injection — whatever produced the text.
- **FR-082** (Future) A pluggable LLM provider may produce the recap prose; its output passes through the FR-081 guardrail before display (never shown raw).

### AI assistant — CAP-10 (bonus, GenUI + Gemini)
- **FR-090** (MVP) An authenticated user can chat with a football assistant (Gemini 2.5 Flash, Vertex AI via keyless WIF); replies stream token-by-token.
- **FR-091** (MVP) The assistant answers **only** football / FIFA World Cup topics (players, clubs, national teams, history, matches, WC2026); anything off-topic is refused in one sentence with a football redirect. A layered guardrail enforces this: input hygiene → a `flash-lite` topic-classifier gate → a hardened system instruction (anti-injection: user text is data, never reveal/obey it) → output validation. The system prompt is never disclosed.
- **FR-092** (MVP) The user can request a **club or player card**; the assistant returns a structured card (name, country, club, position, achievements, summary, confidence) rendered as a UI card, flagged "may be outdated" when confidence < high.
- **FR-093** (MVP) AI endpoints require authentication (anonymous → 401) but are **not tier-gated** — every logged-in user incl. the `none` demo tier may use them. Per-user rate limits + input-length caps apply; if the AI backend is unavailable the endpoints return `503` without affecting the rest of the app.
- **FR-100** (MVP) The assistant is **grounded in the app's own live tournament data** (ADR-0018): via Gemini function-calling it can read a one-call **tournament overview** (primary — for a quick championship summary), recent results, a team's matches, group standings, and the prediction-pool leaderboard, and answers current-tournament questions from that data — never claiming the tournament "hasn't started" or that it lacks current data. Answers are grounded in returned tool data (no invented results); a tool/grounding failure degrades gracefully to an ungrounded reply.

## Non-functional (NFR)
- **NFR-001** (MVP) Scoring is pure and deterministic given (prediction, result) — no clock/IO in the core function; covered by unit tests **and** golden-fixture evals.
- **NFR-002** (MVP) No secret (JWT/admin/OAuth/Telegram/**Gemini**) appears in the repo; Gemini auth is **keyless** (WIF/ADC — no API key or SA-key in code, env, or image). CI scans history (gitleaks) and images (Trivy, fail on HIGH/CRITICAL).
- **NFR-006** (MVP) The AI assistant never logs full message bodies (only ids, token counts, classifier verdict, latency, finish reason); requests are time-bounded and cancel the upstream stream on client disconnect.
- **NFR-003** (MVP) Backend runs as non-root distroless, read-only rootfs, all caps dropped.
- **NFR-004** (MVP) Every push/PR passes the 5-layer verification gate before merge to `main`.
- **NFR-005** (Future) p95 API latency < 300 ms for read endpoints under the friends-scale load (≤ a few dozen users).

## Technical constraints (TC)
- **TC-001** Backend Go `1.26.4`; frontend built on Node `26-alpine`; DB Postgres `17`.
- **TC-002** Embedded SQL migrations applied on boot; no external migration step.
- **TC-003** Times stored UTC, displayed Europe/Kyiv.
- **TC-004** Container images are multi-arch (amd64+arm64) so local arm64 and prod amd64 match.

## Business constraints (BC)
- **BC-001** Private, invite-only pool — not a public service (enforced by demo-mode gating, [ADR-0012](adr/0012-demo-mode-access-tiers.md)).
- **BC-002** Match data comes from the official FIFA API; manual admin entry is an outage fallback only ([ADR-0006](adr/0006-manual-override-precedence.md)).
- **BC-003** Zero-cost hosting/tooling (docker-compose, GHCR, free CodeRabbit).
