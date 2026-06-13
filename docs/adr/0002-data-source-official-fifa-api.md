# ADR-0002: Match data from the official FIFA API, providers pluggable

**Status:** Accepted · **Date:** 2026-06-13

## Context
We need fixtures (calendar), live scores, final results, venues, team flags, and knockout bracket data for WC2026. An early idea was to scrape sport.ua. Research verified that the official FIFA backend `api.fifa.com/api/v3` is live and serving real WC2026 data (`IdCompetition=17`, `IdSeason=285023`), unauthenticated, CORS-open, Akamai-fronted with a 5-minute edge cache. It also returns flag URLs (`PictureUrl`) and bracket placeholders (`PlaceHolderA/B`).

## Decision
Use **`api.fifa.com/api/v3` as the primary data source**, accessed through a pluggable `ResultsProvider` interface. Implement **ESPN's hidden API** (`site.api.espn.com/.../soccer/fifa.world`) and optionally **football-data.org** (`/v4/competitions/WC`, has clean standings) as fallbacks. Drop sport.ua scraping entirely.

Key endpoints: `/calendar/matches` (schedule + scores + status, paginate via `ContinuationToken`), `/stages` (knockout stage IDs), `/live/football/...` (live detail).

## Consequences
- Cleaner, structured JSON than HTML scraping; gives flags and pre-draw bracket placeholders for free.
- The API is **undocumented and unsupported** — schema can change without notice, and Akamai may challenge aggressive polling. Mitigations: reuse one `*http.Client` with a cookie jar, rate-limit with `golang.org/x/time/rate`, poll live matches ~30–60s and schedule sync ~daily, cache in Postgres, keep ESPN as a hot fallback.
- **ToS gray area:** acceptable for a private, non-commercial, low-frequency pool. **Legal review required** before any non-personal/commercial use.

## Alternatives considered
- **sport.ua scraping** — brittle HTML, no structured bracket/flags. Rejected.
- **football-data.org as primary** — clean and documented but free tier is 10 req/min and not the official source. Kept as a fallback.
- **Paid APIs (API-Football, SportMonks, Sportradar)** — overkill and cost money for a friends pool. Rejected.
