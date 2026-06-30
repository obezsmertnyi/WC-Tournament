# ADR-0012: Demo mode with per-user access tiers (none / ro / rw)

**Status:** Accepted · **Date:** 2026-06-30

## Context
ADR-0005 left an open concern: *"Need an allow-list / approval gate so a public Google login can't self-enroll into a private pool."* Google sign-in is published (anyone with a Google account can authenticate), so any stranger could land in the app and immediately see other players' predictions, the leaderboard, the audit feed, and could submit predictions/bonus picks. We need a gate that (a) lets the owner show the product to anyone safely, and (b) keeps the live pool private until the owner deliberately admits someone — **without locking out the existing players** mid-tournament.

## Decision
A global **demo mode** toggle plus a **per-user access tier**. Tiers, ordered:

- **`none`** — browse the whole UI (calendar, groups, bracket, own empty profile/history); cannot see other players' data, cannot participate.
- **`ro`** — also see reveals, leaderboard, audit, top-scorers.
- **`rw`** — also participate: submit predictions and bonus picks.

Resolution rules (`auth.ComputeAccess`):
- **admins** always resolve to `rw`;
- when demo mode is **off**, *everyone* resolves to `rw` (pre-demo behaviour preserved);
- only when demo mode is **on** does a non-admin's stored `access_level` apply.

Initial level for a new account is derived in SQL from the current demo mode: a self-service **Google** sign-up gets `none` while demo mode is on (else `rw`); an **admin-provisioned** roster player is always `rw` (a deliberate act). Existing users were migrated to `rw`, so enabling demo mode never locks out current participants.

Enforcement is a single global middleware, **`auth.DemoGate`**, that maps `METHOD + matched-route` → required tier and returns `403 {error:"demo_locked", access, need}` when insufficient. Reads of *other* players' data require `ro`; participation requires `rw`; own/public data (calendar, standings, `/api/me`, own history) is open to any authenticated user. Centralising in one middleware means **zero handler-signature changes**.

- **Storage:** `users.access_level TEXT NOT NULL DEFAULT 'rw' CHECK (access_level IN ('none','ro','rw'))` (migration 0010); `app_state['demo_mode']` = `'true'|'false'`.
- **API:** `/api/me` exposes `demoMode` + effective `access`; admin endpoints `GET/PUT /api/admin/demo` (toggle) and `PUT /api/admin/users/:id/access` (grant), both audited (`admin_demo_mode`, `admin_set_access`). Changing an admin's access is refused (403).
- **UI:** a preview banner explains the restriction; gated panels render a "request access" card instead of fetching; the admin roster gains a demo switch and a per-player tier selector.

## Consequences
- Resolves the ADR-0005 private-pool gap: the owner can hand the URL to anyone; strangers see a harmless preview until granted access.
- Off by default and a no-op when off — safe to ship without behaviour change; the owner enables it explicitly.
- One DB read per guarded request (cheap, single-column) to resolve a non-admin's tier under demo mode.
- Manual per-user admission (no email allow-list). Acceptable for a small friends pool; an allow-list could layer on later if the roster grows.

## Alternatives considered
- **Email allow-list** — less flexible for ad-hoc "let them look around first" demos; still viable as a future addition.
- **Per-route guards threaded through every `Register*` function** — more code and test churn than one path-keyed middleware.
- **Reusing the unused `approved` boolean** — only two states; can't express browse-only vs. read-only vs. play.
