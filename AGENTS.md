# AGENTS.md — WC-Tournament

Single source of truth for project context, shared across coding agents
(Claude Code imports this via `@AGENTS.md` in `CLAUDE.md`; Codex and other tools
read this file directly). Edit **here** — do not duplicate rules into `CLAUDE.md`.

Keep it lean (this is the always-loaded budget); per-session state lives in the
**dynamic** context (see boundary below), not here.

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
- `docs/` — SDD + decisions: `product-brief`, `requirements`, `mvp-capability-plan`, `architecture`, `features/<cap>/spec.md`, `adr/`, `qa/`, `diagrams/`
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
- Major structural changes get an **ADR** first. Run `make ci` after changes.
- Commits are SSH-signed; branch names `feat/* fix/* chore/*`.

## Static vs dynamic context (the boundary)
- **Static** (here + committed): this file (`AGENTS.md`, imported by `CLAUDE.md`),
  `docs/` (brief, requirements, capability plan, specs, ADRs, architecture).
  Stable across sessions.
- **Dynamic** (per session/slice): `docs/memory/activeContext.md` +
  `progress.md`, `current-state.md` (fast handoff), `.workflow-state.toon`
  (loop state), `CHECKLIST.md`. Read these to resume; don't bloat the static
  context with them.

## Spec-driven flow
Change a behavior → update/author the capability spec (`docs/features/<cap>/spec.md`,
Given/When/Then) and requirements (`docs/requirements.md`, FR/NFR ids) → implement
→ test (`@trace FR-id`) → verify (`make ci` + evals) → separate review → commit
with `Slice:`/`Refs:` trailers. See `WORKFLOW.md` / `LOOP.md`.
