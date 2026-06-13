# ADR-0003: Frontend — React + Vite + TypeScript, production-grade

**Status:** Accepted · **Date:** 2026-06-13

## Context
This explicitly **replaces the old Excel sheet with a polished, production/enterprise-grade web product** — not a spreadsheet clone. The frontend must be genuinely good-looking and production-ready (explicit, repeated requirement), football-themed, and show fixtures grouped by stage, a leaderboard/ratings table, a knockout bracket, player profiles, and a public audit feed. Country flags are **mandatory** on every team; player avatars appear on the leaderboard and profiles. Players use phones.

## Decision
**React + Vite + TypeScript**, styled with **Tailwind + shadcn/ui**, animated with **framer-motion**, built to a static SPA and served by **nginx** (which proxies `/api` to the backend).

**Design language — dark, premium, minimalist ("expensive"), mobile-first.** Reference: a near-black spotlight stage with one restrained accent.
- **Background:** `#0B0C0E → #15171B` charcoal gradient with a soft radial spotlight glow.
- **Text:** `#F5F6F7` primary, `#9AA0A6` muted.
- **Accent:** champagne gold `#C9A24B`, used sparingly (active tab, points↑, "2026", winner row). Color present but never garish.
- **Surfaces:** glass cards `rgba(255,255,255,.04)` with a 1px hairline border + backdrop blur; generous spacing; tabular-nums for scores.
- **Motion:** subtle only — fade/slide-in on mount, gentle hover lift, number transitions, a faint spotlight breathe. No flashy effects.
- **Responsive:** mobile-first; must open and work well on phones (the primary device for entering predictions).

- **Flags:** render from a local package (`flag-icons` keyed by ISO 3166-1 alpha-2) so they never break if the FIFA CDN is unavailable; the provider `PictureUrl` is a fallback, not the source of truth.
- **Avatars:** generated (e.g. DiceBear seeded by name) for MVP, with optional upload later.

## Consequences
- Strong typing across the app; component library accelerates a polished UI.
- Mobile-first layout required (prediction entry from a phone is the primary flow).
- Flags bundled locally → larger asset, but reliability over the live tournament is worth it.

## Alternatives considered
- **Server-rendered templates (Go html/template)** — faster to stand up but does not meet the "polished, prod-ready" bar.
- **Next.js** — SSR unneeded for a small private SPA; adds a Node runtime to the deployment.
