# Architecture — WC2026 Prediction Pool (Local Solution)

**Status:** Draft v0.1 · **Date:** 2026-06-13 · **Owner:** o.bezsmertnyi
**Scope:** Local-first deployment via `docker-compose`. Production hosting is out of scope for this iteration.

> ⏰ **Time context:** the tournament is already underway (group stage started 2026-06-11). Target: a working MVP in days. Backfill predictions for already-played matchdays is an admin decision (see Open Questions).
>
> ℹ️ **Tournament structure is NOT hardcoded.** The number of teams, groups, and knockout rounds is whatever the data source reports (ADR-0007). The only things carried over from the 2022 sheet are **the participants (the friends who play)** and **the scoring method** — see §5.

## 1. Decisions locked
- **Backend:** Go 1.26 + Gin + Postgres 17 (matches Everstake stack; `pgx`).
- **Frontend:** React + Vite + **TypeScript**, football-themed UI (Tailwind + shadcn/ui), served by nginx. Must be **production-ready and polished** — not a throwaway. Country flags are **mandatory** on every team; player avatars on leaderboard/profiles.
- **Auth:** **Sign in with Google** (OAuth/OIDC) → backend JWT session. Players set a **nickname** in their profile (the display name everywhere). A private pool is gated by **demo mode + per-user access tiers** (`none`/`ro`/`rw`): when demo mode is on, a self-service Google sign-up can only browse the UI until an admin grants access; admins and demo-off always resolve to `rw`. Enforced by one global `DemoGate` middleware. (ADR-0005, ADR-0012)
- **Deployment:** `docker-compose` — `db`, `backend`, `frontend`. Single host, local.
- **Results/calendar source:** **official FIFA API** `api.fifa.com/api/v3` (`IdCompetition=17`, `IdSeason=285023` — verified live for WC2026), behind a pluggable provider interface, with manual admin override always winning. ESPN hidden API (`site.api.espn.com/.../soccer/fifa.world`) and football-data.org (`/v4/competitions/WC`) as fallbacks. sport.ua scraping dropped — the official API is cleaner and gives flags + bracket placeholders.

## 2. System overview

```
┌──────────────┐      HTTPS/JSON      ┌──────────────────────────┐
│  Frontend    │  ───────────────▶    │  Backend (Go / Gin)      │
│ React+nginx  │                      │  REST API + JWT          │
│ calendar /   │  ◀───────────────    │  scoring engine          │
│ table / bracket                     │  scheduler (cron)        │
└──────────────┘                      └─────────┬───────┬────────┘
                                                 │       │
                                       ┌─────────▼──┐ ┌──▼──────────────┐
                                       │ Postgres   │ │ ResultsProvider │
                                       │ (state)    │ │ FIFA API (+f/b) │
                                       └────────────┘ └─────────────────┘
```

Three containers. The results sync runs as an in-process scheduled job inside the backend for the MVP (no separate worker) — split out later if needed.

## 3. Components

### 3.1 Backend (Go / Gin)
Internal package layout (mirrors the claimer convention):
- `cmd/server/` — entrypoint, CLI (`serve`, `scrape-once`, `recompute-scores`, `seed`).
- `internal/api/` — Gin handlers + JWT middleware.
- `internal/predictions/` — submit / lock logic.
- `internal/scoring/` — deterministic scoring engine (rules from config).
- `internal/results/` — `ResultsProvider` interface + `fifa` (primary) / `espn` / `footballdata` implementations + `manual` override.
- `internal/standings/` — leaderboard + group-table computation.
- `internal/bracket/` — knockout progression (auto-derive + admin override).
- `internal/storage/` — Postgres (pgx/sqlx), migrations.
- `internal/scheduler/` — cron (`robfig/cron`) for periodic sync + reminder dispatch.
- `internal/notifier/` — Telegram bot notifier (Telegram only — no email). Sends **friendly, Ukrainian** reminder copy. Mirrors the claimer's notifier pattern.
- `internal/achievements/` — badges + per-matchday winner computation.
- `internal/config/` — env-based config (incl. `TELEGRAM_BOT_TOKEN`, never committed).

### 3.2 ResultsProvider (decoupled source)
```go
type ResultsProvider interface {
    Fixtures(ctx context.Context) ([]Fixture, error) // calendar
    Results(ctx context.Context) ([]MatchResult, error)
    Standings(ctx context.Context) ([]GroupStanding, error) // may be derived locally
}
```
- **`fifa` (primary):** official `api.fifa.com/api/v3`. Verified live for WC2026 — `IdCompetition=17`, `IdSeason=285023`.
  - `/calendar/matches?idCompetition=17&idSeason=285023&count=200` → full schedule + scores + status (walk `ContinuationToken`).
  - `/stages?...` → knockout stage IDs (R32 `289287`, R16 `289288`, QF `289289`, SF `289290`, 3rd `289291`, Final `289292`).
  - `/live/football/17/285023/{idStage}/{idMatch}` → live match detail.
  - Names come as localized arrays (`TeamName[].Description`) — pick `en`. Flags via `PictureUrl`. Bracket pre-draw via `PlaceHolderA/B`.
  - No auth; CORS open; Akamai-fronted with 5-min edge cache → reuse one `*http.Client` with a cookie jar, rate-limit (`golang.org/x/time/rate`), poll live ~30–60s, schedule sync ~daily.
- **`espn` (fallback):** `site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard?dates=YYYYMMDD` — no key, near-real-time, used if FIFA schema shifts or throttles.
- **`footballdata` (optional fallback):** `api.football-data.org/v4/competitions/WC/matches` + `/standings`, `X-Auth-Token`, 10 req/min — has clean standings.
- **Standings** are **derived locally** from finished group-stage results (FIFA's standings path returns null pre-completion); only 104 matches, cheap to compute. Tie-break: points → GD → goals → (fair play / draw).
- The **provider is the source of truth**. **Manual admin entry is an outage fallback** (`result_source='manual'`) used only when the API is down but a match must be scored. On API recovery the system reconciles: equal → upgrade source silently; differ → log discrepancy to audit, adopt the provider value, re-score. An admin can explicitly **lock** a row to keep a manual value (audited). (ADR-0006)
> ⚖️ Hitting the undocumented FIFA API for a private, non-commercial, low-frequency pool is low-risk but a ToS gray area — keep it personal, cache aggressively. **Legal review required** before any non-personal use.

### 3.3 Frontend (React + Vite)
Pages:
- **Calendar / Fixtures** — grouped **Group stage (by group A–L) → Knockout (R32 → Final)**. Each fixture card shows: date + kickoff time in **Ukraine time (Europe/Kyiv)**, **both team flags + names**, **venue: city + stadium** (and host country flag), live status/score, and your prediction inline (editable until kickoff). Knockout cards also take a **winner pick** (who advances). Before the draw, knockout cards render `placeholder_home/away` ("Winner Group A").
- **Leaderboard** — ranked table of players with points, deltas, per-stage breakdown; **live-updating** during matches (poll/websocket). Shows badges and the per-matchday winner.
- **Bracket** — knockout tree, auto-populated after groups, admin-editable.
- **Activity / Audit (public)** — chronological feed visible to all players: who placed/changed a prediction and when (the score itself appears only after that match's kickoff), plus every admin action (result overrides with old→new, rule changes, bracket edits, user creation). Read-only, filterable by player/match/action.
- **Profile** — player sets a **nickname**, uploads a **photo** (overrides the Google avatar), picks a **favorite-team flag**, links **Telegram** (for reminders), and makes the enabled **bonus picks** (champion / finalist / top scorer). Shows the player's badges.
- **Admin** — enter/override results, manage users, trigger sync, edit scoring rules. Every action here is written to the public audit feed.
Visual language: dark, premium-minimal "expensive" theme — near-black charcoal + spotlight glow, champagne-gold accent (`#C9A24B`), glass cards, subtle framer-motion, mobile-first. Team flags everywhere. (ADR-0003)

## 4. Data model (Postgres)

```
users(id, google_sub UNIQUE, email, nickname,
      avatar_url,                      -- Google picture by default, overridable by uploaded photo
      favorite_team_code,              -- chosen flag (ISO code) shown next to the nickname
      telegram_chat_id,                -- set after the player /starts the bot; for DM reminders
      role[player|admin], approved BOOLEAN, created_at)
teams(id, name, code, flag_url, group_label)
matches(id, stage[group|r32|r16|qf|sf|final|third], group_label, match_number,
        home_team_id, away_team_id, kickoff_at, status[scheduled|live|finished],
        home_score, away_score, home_pens, away_pens,
        venue_stadium, venue_city, venue_country, venue_country_code,
        placeholder_home, placeholder_away,   -- "Winner Group A" before draw
        result_source[fifa|espn|footballdata|manual], updated_at)
predictions(id, user_id, match_id, home_pred, away_pred,   -- regular-time score
            winner_pick_team_id,                            -- knockout only: who advances (incl. ET/pens)
            created_at, updated_at, UNIQUE(user_id, match_id))   -- locked at matches.kickoff_at
scoring_rules(id, exact_score_pts DEFAULT 3, outcome_pts DEFAULT 1,
              knockout_winner_pts DEFAULT 1,
              active)                                       -- no goal-diff tier, no multiplier
bonus_rules(id, kind[champion|finalist|topscorer], enabled BOOLEAN, pts)  -- optional tournament bonuses
tournament_picks(id, user_id, kind[champion|finalist|topscorer],
                 pick_ref,                                  -- team_id (champion/finalist) or player ref (topscorer, admin-confirmed)
                 locked_at, created_at, updated_at, UNIQUE(user_id, kind))
points(id, user_id, match_id, points, breakdown_json)  -- materialized per scored match
achievements(id, user_id, kind[most_exact|perfect_round|matchday_winner|...], context, awarded_at)
notifications(id, user_id, kind[nudge|round_start|round_summary|bonus_deadline],
              channel[telegram], sent_at, status)   -- Telegram only
bracket_slots(id, stage, slot_label, team_id, source[derived|manual])
audit_log(id, actor_user_id, actor_role[player|admin],
          action[prediction_create|prediction_update|result_override|
                 match_upsert|user_create|rules_change|bracket_override|login],
          match_id, target_user_id, before_json, after_json,
          created_at, public_visible_at)
```

**Audit policy (public transparency):** every prediction write and every admin action appends an immutable `audit_log` row (insert-only, never updated/deleted). `public_visible_at` controls disclosure: for predictions it equals the match `kickoff_at` (so the *fact + timestamp* of a prediction is public immediately, but the *predicted score* is masked until lock — prevents copying); for admin/system actions it equals `created_at` (public at once). The public Activity tab reads only rows where `now() >= public_visible_at` for sensitive fields.

Standings/leaderboard are computed from `points` (view or on-demand aggregate), not stored denormalized.

## 5. Scoring engine
A prediction is a **regular-time score** ("основний час"). Method recovered from WC2022 + owner's knockout clarification (full rationale in ADR-0008):
- **Exact regular-time score → 3 points.**
- Else **correct outcome only** (W/D/L) → **1 point.** Else **0**.
- **Knockout matches add a winner pick** (which team advances, decided incl. extra time / penalties): correct → **+1**, stacked on the score points.
  - e.g. predict `1:1` reg-time **and** France advances → 3 + 1 = **4**. Group matches have no winner pick → max 3; knockout → max 4.
- **No goal-difference tier, no multiplier.** Penalties/ET count only for the winner pick, never as part of the predicted score.

Deterministic, pure function of `(prediction, result, rules)`. Values in `scoring_rules` (defaults exact=3, outcome=1, knockout_winner=1), versioned and **frozen + confirmed by all players before scoring starts**. Recompute is idempotent (`recompute-scores` CLI), re-runs on any result change.

## 6. Prediction lifecycle
1. Results sync (FIFA) loads fixtures → `matches`.
2. Player submits `home_pred/away_pred` (+ `winner_pick` on knockout) per match; editable until `kickoff_at`.
3. At kickoff the API rejects edits (server-side check vs UTC `kickoff_at` — never trust the client). Exception: the one-time, admin-controlled **backfill window** for the opening matchdays, fully audited (ADR-0010).
4. Provider sets `home_score/away_score`, `status=finished`. If the API is down, the admin fills it manually (`result_source='manual'`); the provider reconciles on recovery (ADR-0006).
5. Scoring engine writes `points`; leaderboard updates. Re-runs idempotently on any later result change.

## 7. Knockout progression
- **We do not hardcode the format.** Groups, the set of knockout rounds, and which teams advance are read from the data source (FIFA API `/stages` + `/calendar/matches`); ADR-0007.
- Group tables shown in the UI are derived locally **for display only** — never used to decide who advances.
- The admin can override any bracket slot (`source=manual`).
- Before teams are known, knockout fixtures render their `placeholder_home/away` ("Winner Group A") straight from the source.

## 8. docker-compose layout
```
wc2026/
├── docker-compose.yml          # db, backend, frontend
├── backend/                    # Go service + Dockerfile
│   ├── cmd/server/
│   ├── internal/...
│   └── migrations/
├── frontend/                   # React+Vite + Dockerfile (nginx)
└── docs/
```
- `db`: postgres:17, named volume, healthcheck.
- `backend`: depends_on db (healthy), runs migrations on boot, exposes `:8080`.
- `frontend`: nginx serving the built SPA, proxies `/api` → backend.
- Config via `.env` (no secrets committed).

## 9. Build order (PoC, timeboxed for the live tournament)
1. **M0 — Skeleton:** docker-compose up with db + Go health endpoint + React shell.
2. **M1 — Data + calendar:** schema; `fifa` sync fills teams/fixtures/venues; calendar view (group→knockout) with flags, venue, Kyiv time. *(Read-first — simpler and more urgent than admin CRUD.)*
3. **M2 — Auth + predictions:** Google sign-in, profile (nickname/photo/flag/Telegram link), submit score + knockout winner pick, server-side kickoff lock + one-time backfill window.
4. **M3 — Results + scoring + leaderboard + public audit:** FIFA results → scoring engine (3/1/+1) → leaderboard + the public Activity/Audit feed. *(Usable end-to-end here.)*
5. **M4 — Bracket:** knockout structure from FIFA + admin override; manual-entry outage fallback + reconciliation.
6. **M5 — Telegram reminders:** notifier, DM + channel, friendly Ukrainian copy, "who hasn't predicted" scheduler.
7. **M6 — Extras + polish:** live-updating leaderboard/scores, achievements + matchday winner, bonus picks (champion/finalist/top scorer), final enterprise-grade UI polish.

Tests are written alongside each milestone by sub-agents (scoring engine, lock/backfill, reconciliation, and provider parsing are the priority units).

## 10. Open questions
1. Leaderboard tie-breakers (e.g. most exact scores, then head-to-head).
2. Private-pool gate: email allow-list vs. admin approval on first Google login (ADR-0005).
3. Sync cadence + ESPN-fallback trigger if the FIFA API schema shifts or Akamai throttles mid-tournament.
4. Optional `champion_bonus` (e.g. +5) — include a tournament-winner pick or keep strictly `3/1 (+1 KO)`?

**Resolved:** auth → Google OAuth + nickname (ADR-0005); scoring → exact 3 / outcome 1 / knockout winner +1, penalties count for the winner pick, no goal-diff (ADR-0008); time zone → Europe/Kyiv display, UTC storage; manual entry → outage fallback only (ADR-0006); backfill → scoring from match #1, one-time audited backfill window for the opening matchdays (ADR-0010); **no Excel — a polished, production/enterprise-grade web product**.
```
