# AGENTS.md — WC-Tournament

Single source of truth for project context, shared across coding agents
(Claude Code imports this via `@AGENTS.md` in `CLAUDE.md`; Codex and other tools
read this file directly). Edit **here** — do not duplicate rules into `CLAUDE.md`.

Keep it lean (this is the always-loaded budget); per-session state lives in the
**dynamic** context (see boundary below), not here.

**New session? Orient here:** [`docs/REPO-MAP.md`](docs/REPO-MAP.md) (indexes every
file by purpose) → `docs/memory/` (dynamic context / what's in flight).

**About to change behavior? That's a *slice* — load [`LOOP.md`](LOOP.md) and follow the
cycle:** spec → implement → trace (`@trace FR-id`) → verify → review (maker≠checker)
→ commit. Entry commands: `/new-capability`, `/verify`, `/trace`, `/review`. A red
gate is a **STOP** — fix it, never bypass (deterministic gates fire on their own via
hooks + CI; the *order* and the *review* are on you).

## What this is
A private, friends-only FIFA World Cup 2026 score-prediction pool. Players
predict fixtures, the app scores them against official FIFA results, a live
leaderboard + Telegram bot keep everyone in the loop. Bilingual (UA/EN),
mobile-first. Live at `wc2026.mtgrd-das.app`.

## Stack (pinned)
- **Backend:** Go `1.26.4` · Gin · `pgx/v5` · embedded SQL migrations · distroless image
- **Frontend:** React 19 · Vite · TypeScript · Tailwind 3 · framer-motion · react-i18next · built on Node `26-alpine`
- **DB:** Postgres `17`
- **Auth:** Google OAuth (OIDC) + admin password · JWT cookie
- **Data:** official FIFA API (ESPN / football-data.org fallbacks)
- **Notifications:** Telegram Bot API
- **Delivery:** docker-compose · GHCR multi-arch images · GitHub Releases (`v*` tags)

## Layout
- `backend/` — Go service (`cmd/server`, `internal/{api,auth,scoring,storage,…}`, `migrations/`)
- `frontend/` — React SPA (`src/{pages,components,lib,auth,…}`)
- `docs/` — SDD + decisions: `project-brief`, `requirements`, `mvp-capability-plan`, `architecture`, `features/<cap>/spec.md`, `adr/`, `qa/`, `diagrams/`
- `mcp/` — read-only MCP server exposing the pool to agents
- `evals/` — golden-fixture evals (the quality bar)
- `.github/` — CI (`ci.yml`) + release (`release.yml`) + dependabot

## Commands (see `Makefile`)
```bash
make build        # backend + frontend
make test         # backend tests (set DATABASE_URL for integration)
make ci           # local equivalent of CI gates
cd frontend && npm run build && npm test   # tsc + vite + vitest
cd backend && go test -tags=evals ./...    # scoring evals
```

## Guardrails (always)
- **No secrets in code.** `JWT_SECRET`, `ADMIN_PASSWORD`, `GOOGLE_OAUTH_*`,
  `TELEGRAM_*` live only in `.env` (gitignored) / host env. `.env.example`
  documents shapes with placeholders.
- **Fix the root cause — no workarounds.** Never disable/skip a failing gate or
  write a test that asserts buggy behavior. Bump to a patched version, fix the
  cause, prove it green.
- **Scoring invariants** (see `docs/adr/0008` + `docs/features/scoring/spec.md`):
  exact 3 · outcome 1 · knockout advancer +1 **only on a predicted regulation
  draw**. Bonuses time-tiered, awarded only if correct. Don't change without an ADR.
- **Demo-mode access gating** (`docs/adr/0012`): when demo mode is on, non-admin
  access tiers (none/ro/rw) gate the API; admins/`demo-off` resolve to `rw`.
- **Manual result override beats a later FIFA sync** (`docs/adr/0006`).
- **UI is design-first** (ADR-0021): produce a reviewable design artifact before
  implementing; honor existing tokens (`frontend/tailwind.config.js`) + real i18n
  copy; **no emoji** anywhere (app or docs) — use inline SVG icons or typographic
  marks. Playbook: `docs/design-first-workflow.md`.
- **Multi-edition (ADR-0022): per-edition data is scoped by `tournament_id`; exactly
  one edition is active (`tournaments.is_active`).** Phase 1 (schema + backfill,
  active = WC2026) is live and non-breaking. Edition-scoping of queries/upserts, the
  archived-edition write-guard, edition-qualified uniques, admin UI + switcher are
  **PENDING** (Phases 2–4). **Until Phase 2 lands, reads are NOT yet edition-filtered,
  so do NOT create a second edition on prod** — a second edition's rows would leak
  into current views and every `fifa_id` lookup would be ambiguous (fifa_ids repeat
  across editions). When you do add an edition-scoped query, filter by
  `active_tournament_id()` and scope every `fifa_id` lookup to the edition. Runbook:
  `docs/repeat-tournament-runbook.md`.
- Major structural changes get an **ADR** first. Run `make ci` after changes.
- Commits are SSH-signed; branch names `feat/* fix/* chore/*`.

## Correctness rules (learned from bugs)
Hard-won invariants — each cost a real bug. Do not "simplify" them away.
- **FIFA `MatchStatus` is unreliable:** `0` = played/finished, `1` = not started,
  `3` = LIVE (stays `3` through extra time / penalties). Never treat code `3` or
  a merely-present score as full-time — finish only on `0` (+ score + past kickoff).
- **Knockout advancer +1** is awarded **only when both the prediction and the
  actual regulation result are draws** and the pick matches. A decisive result
  (either side) gets no +1 (`0:1` predicted `1:1` must score 0, not +1).
- **Extra-time GOAL wins are auto-corrected from FIFA goal periods (ADR-0020).**
  The FIFA calendar feed (`fifa_types.go`) carries only the *final*
  `HomeTeamScore/AwayTeamScore` (aet-inclusive) + a separate penalty score —
  there is **no regulation-90 field**. A knockout won by an ET goal would store
  the aet score (e.g. `3:2`) and the regulation-based scoring would read it as
  *decisive*, robbing correct `2:2`-draw + advancer predictions. Root fix: the
  **live** goal events carry a `Period` (3/5 = regulation halves, 7/9 = extra
  time), so `results.RegulationScore` recovers the 90-minute scoreline and
  `winners.Run` writes it (`result_source='fifa_regulation'`, protected from the
  next calendar sync) while keeping `winner_team_id` = advancer. Safeguard: only
  trust the derivation when every goal has a period+side **and** the {3,5,7,9}
  goals sum to the aet score — else keep the stored score (never zero a match
  with a missing goal timeline). A **manual override** (ADR-0006) still wins and
  remains the escape hatch for an incomplete-timeline straggler. (History: #81
  BEL–SEN and #87 ARG–CPV were manually patched before the root fix landed
  2026-07-04.)
- **Bracket order = tree geometry**, not FIFA match number (feeder→parent slot);
  ordering by number misaligns the connectors.
- **AI recap grounding is scoreline-only** — keep stage labels digit-free, and a
  non-grounded provider without a team registry must fall back to the template
  (never display raw model output).
- **AI chat grounding must state the outcome explicitly — never let the model
  infer who won.** WC 2026 is past the model's training cutoff, so every fact
  comes from the tools. Given only a raw home:away `score`, the model inverted
  win/loss (reported teams *losing* matches they won, e.g. Colombia). Fix: the
  grounding `MatchFact` carries an explicit `winner` (or `"draw"`) and, for
  knockouts, `advanced` — computed in Go from the authoritative scores +
  `winner_team_id` (`ai_tools.go`), and the master prompt tells the model to use
  those fields and never derive the result from the score itself.
- **Admin "predict on behalf":** the `<select>` value is a string, the API sends
  the id as a number — compare as strings (`String(a) === b`).
- **Bounds-check before indexing** (e.g. `medals[i]`) — the digest panicked with
  >3 players. Guard first.
- **Ship multi-arch images** (amd64+arm64) via CI/GHCR — never `docker save|load`
  a local arm64 image onto amd64 prod (`exec format error`).

## Static vs dynamic context (the boundary)
- **Static** (here + committed): this file (`AGENTS.md`, imported by `CLAUDE.md`),
  `docs/` (brief, requirements, capability plan, specs, ADRs, architecture).
  Stable across sessions.
- **Dynamic** (per session/slice): `docs/memory/activeContext.md` +
  `progress.md`, `.workflow-state.toon` (loop state). Read these to resume;
  don't bloat the static context with them.

## Spec-driven flow
Change a behavior → update/author the capability spec (`docs/features/<cap>/spec.md`,
Given/When/Then) and requirements (`docs/requirements.md`, FR/NFR ids) → implement
→ test (`@trace FR-id`) → verify (`make ci` + evals) → separate review → commit
with `Slice:`/`Refs:` trailers. See `WORKFLOW.md` / `LOOP.md`.
