# ADR-0007: Knockout qualifiers & bracket come from FIFA, not locally derived

**Status:** Accepted · **Date:** 2026-06-13

## Context
The tournament's exact format (number of groups, how many advance, how third-placed teams are ranked) is **not a requirement we were given and must not be assumed** — earlier drafts wrongly hardcoded a specific bracket. Whatever the real format, official tie-breakers are non-trivial (and can end in drawing of lots). A locally computed ranking that diverges from the official one would produce a **wrong bracket** and unfair scoring.

## Decision
Take the **format, knockout qualifiers, and bracket structure directly from the data source** (FIFA API `/stages`, `/calendar/matches` with `idStage`, and `PlaceHolderA/B`) — never hardcode team/group/round counts. Locally **derived standings are display-only** (group tables shown in the UI) and are **never** used to decide who advances. The admin can still override any bracket slot (`source=manual`, ADR-0006).

## Consequences
- The bracket always matches the official tournament — no custom tie-break logic to get wrong.
- Before the draw, bracket cards render `placeholder_home/away` ("Winner Group A").
- Dependency on the FIFA API for bracket data; the ESPN fallback (ADR-0002) also exposes stage groupings.

## Alternatives considered
- **Derive qualifiers locally from results** — would require replicating FIFA's full tie-break ladder (incl. lot-drawing); high risk of divergence. Rejected as the source of truth.
