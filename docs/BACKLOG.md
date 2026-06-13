# WC-Tournament — Backlog & Status

Single source of truth for what's done, in flight, and planned. Updated as we go.
Legend: ✅ done & deployed · 🟡 coded, pending deploy/verify · ⬜ planned

## Foundation
- ✅ Docs: project brief, architecture, ADRs 0001–0011 (`docs/`)
- ✅ `.gitignore`/`.env.example` hygiene (Telegram token protected)
- ✅ M0 skeleton: docker-compose (db + Go health + React shell), hardened (named containers, healthchecks, security, limits)
- ✅ Go 1.26, Postgres 17

## Data & calendar (M1)
- ✅ Schema + migrations (teams, matches)
- ✅ Official FIFA API sync (104 real WC2026 matches; idempotent; manual-override safe)
- ✅ `GET /api/matches`, `/api/teams`, `/api/standings`
- ✅ Day-centric calendar (date strip + prev/today/next), responsive grid
- ✅ Groups index (A–L cards) + group detail (standings table + matches)
- ✅ Flags via FIFA-3→ISO-2 map (flag-icons)
- ✅ i18n EN/UK + language switcher (Kyiv time, locale-aware dates)
- ✅ Fixed: `/api/matches` contract, "Group Group A" double label, flag URL tokens

## Visual polish
- ✅ Trophy as a deliberate element (AppBar wordmark + Calendar hero) + faint watermark
- ✅ Host-country flag (🇺🇸/🇨🇦/🇲🇽) next to venue/stadium
- ✅ Ukrainian country/team names (Мексика, Південна Корея…) — switch with language
- ✅ Reduced-motion robustness (fixture cards render reliably; fixed blank lists)
- ✅ "Stars to watch" band on Calendar (Messi/Ronaldo/Mbappé/Vinícius/Bellingham — placeholder portraits + flag)
- ⬜ **Star photos on group pages** — each notable team's star (Argentina→Messi, Norway→Haaland), styled differently than the calendar band. Needs licensed image URLs (owner to provide; copyright — not scraping).
- ⬜ Stars in calendar AND groups, presented differently per context *(requested)*
- ✅ Public access via Cloudflare quick tunnel (temporary URL)

## Match detail (researched)
- ⬜ Match-detail page with FIFA stats. Calendar endpoint already gives officials/attendance/stadium/formations/penalties/weather; deep stats (possession/shots/cards/lineups/events) live in the FIFA `live/...` endpoint — one URL was sandbox-blocked, **verify its fields against a real match before coding structs**.

## Game mechanics (M2+, not started)
- ⬜ Google OAuth + profile (nickname, photo, favorite-team flag, Telegram link)
- ⬜ Predictions: regular-time score + knockout winner pick; server-side kickoff lock
- ⬜ One-time audited backfill window (scoring from match #1)
- ⬜ Scoring engine: exact 3 / outcome 1 / KO winner +1 (penalties count for the pick)
- ⬜ Public, immutable audit feed (predictions masked until kickoff + admin actions)
- ⬜ Telegram reminders (4 types, friendly Ukrainian copy) — token in `.env`
- ⬜ Live-updating leaderboard
- ⬜ Achievements + per-matchday winner

## Bonus scoring (design)
- ⬜ Optional bonus picks: champion / finalist / top scorer (integer points)
- ⬜ **Time-tiered champion bonus**: more points if the champion is picked NOW (before/early) vs after the group stage (later, safer = fewer points). Define the tiers and lock deadlines in `scoring_rules`/`bonus_rules`. *(requested)*
- ⬜ Leaderboard tie-breakers (most exact scores → head-to-head)
- ⬜ Private-pool gate (email allow-list vs admin approval)
