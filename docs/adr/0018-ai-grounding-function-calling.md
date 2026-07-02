# ADR-0018: Ground the AI assistant in live app data via function-calling

**Status:** Accepted · **Date:** 2026-07-01

## Context
The football assistant (ADR-0017) answered from the model's stale knowledge. In this
app the World Cup 2026 pool is **already underway** (matches played and scored in-app),
so the model gave wrong/embarrassing answers: "the tournament hasn't started", "I can't
access current data", talking about players in the past tense. The headline feature the
user wants is a **quick summary of *this* championship**; live stats are a bonus. The app
already holds the authoritative data in Postgres.

## Decision
- **Function-calling grounding** (`google.golang.org/genai` tools), not RAG/pgvector:
  the data is structured, so SQL-backed tools give exact, traceable answers. Tools
  (`backend/internal/gemini/grounding.go`, impl `internal/api/ai_tools.go` over storage):
  `tournament_overview` (**primary** — one call: current stage, played/total, recent
  results, upcoming, pool leader → for a quick summary), `recent_results`, `team_matches`,
  `group_standings`, `pool_leaderboard`.
- **Decoupled seam:** the `gemini` package defines a `Tools` interface + plain fact
  structs; the `api` layer implements it over `*storage.Store` (reusing the standings
  computation). The gemini package never imports storage — keeps the guardrail unit-testable.
- **Loop** (`vertexGen.generateGrounded`): non-streaming GenerateContent with the tool
  declarations; execute each requested `FunctionCall`, feed `FunctionResponse` back, up to
  4 rounds, then a final answer. The grounded reply is delivered whole (streaming UX traded
  for correctness).
- **Grounded system prompt** tells the model the tournament is LIVE and to answer strictly
  from tool data — used **only** with tools attached, so it can't invent results.
- **Fail-safe:** if grounding errors or returns empty, `StreamChat` falls back to the
  ungrounded streaming path — chat never hard-fails because of a grounding hiccup.
- The guardrail (ADR-0017) is unchanged and still wraps the grounded path (classifier gate,
  sanitized history, safety settings).

## Update (as-built, 2026-07-02) — web grounding for external football facts
Function-calling over our DB grounds *tournament/pool* questions, but not external
facts (a player's current club, history) — those were stale from model memory. Vertex
**refuses to combine `googleSearch` with `functionDeclarations`** in one request
("Multiple tools supported only when all search tools"), so we **route by intent**: the
L1 classifier now also returns a `scope` (`tournament` | `general`); `StreamChat` sends
`general` to a **Google-Search-grounded** call (`groundedSearch`, current web facts) and
`tournament` to DB function-calling. **Cards** likewise use Google Search grounding
(`groundedCard`) for current facts (e.g. a 2026 transfer / national-team switch), parsed
from JSON-in-text (search is incompatible with a JSON `responseSchema`), with fallback to
the model-knowledge card. Added the `top_scorers` DB tool. Greetings + identity questions
are allowed (not refused). All verified with live smokes before deploy.

### Card delivery + quality fixes (2026-07-02)
Two bugs kept cards from working end-to-end:
- **Write-deadline truncation.** `groundedCard` takes 10-21s, but the HTTP server's
  10s `WriteTimeout` reset the connection before `c.JSON` wrote the body — the server
  logged 200 while the client's `fetch` failed and the UI showed the generic error.
  `aiCardHandler` now extends the response write deadline (`aiTimeout+10s`), mirroring
  the chat handler. Regression: `internal/api/ai_card_test.go` (real net.Listener +
  short WriteTimeout; a Recorder cannot reproduce this).
- **Thinking budget starving output.** On 2.5-flash, thinking tokens (~1300-1500)
  share `MaxOutputTokens`; for a high-info subject (Messi) they flakily starved the
  JSON so it truncated mid-object → `parseCard` nil → silent low-confidence fallback
  (no stats). `groundedCard` now disables thinking (`ThinkingConfig.ThinkingBudget=0`):
  a factual card needs no chain-of-thought. Result: deterministic complete JSON with
  stats at high confidence, and ~2.5× faster (≈19s→7s). Regression guard added to the
  live smoke `TestLiveGroundedCard` (high-info Messi case: conf≠low, stats>0).

## Consequences
- The assistant answers "як зіграли Англійці", "підсумуй чемпіонат", standings, and pool
  questions from real data (FR-100).
- Deterministic parts (tool dispatch/routing, grounded-vs-fallback decision) are unit-tested
  with a stubbed generator + stubbed tools; the live function-calling round-trip needs WIF
  and is verified in prod (safe because of the fallback).
- Grounded replies are non-streamed (one delivery) — acceptable; the guardrail + caps still apply.

## Alternatives considered
- **pgvector / RAG** — for exact structured facts, SQL tools beat vector similarity; deferred (see the persistence research).
- **Google Search grounding** — external traffic, weakens the football-only guardrail; rejected.
- **Streaming the grounded answer** — the SDK tool-loop is cleanest non-streaming; streaming a post-tool answer added complexity for little gain. Deferred.
