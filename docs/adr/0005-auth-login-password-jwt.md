# ADR-0005: Auth — Google OAuth (Sign in with Google), nickname in profile

**Status:** Accepted · **Date:** 2026-06-13 · *(supersedes the earlier login+password proposal)*

## Context
A closed group of friends. We need per-player identity (predictions and scores isolated and attributable) with the least possible onboarding friction. The owner decided players sign in with **Google**, and each player sets their own **nickname** (the display name on the leaderboard/audit) in their profile.

## Decision
**Sign in with Google (OAuth 2.0 / OIDC).** On first sign-in a `users` row is created from the Google profile (`google_sub`, email, name, picture). The player can set a **`nickname`** in their profile, which is the name shown everywhere (leaderboard, fixtures, public audit); the Google `name`/`picture` are defaults/fallbacks. Session is a backend-issued **JWT** after the OAuth exchange.

- **Access control:** since it's a private pool, restrict who can join — either an allow-list of emails or an admin-approval step on first login (avoid the leaderboard filling with strangers). First/owner account is admin; admin can grant the `admin` role.
- **Secrets:** Google OAuth client ID/secret and the JWT signing secret live in `.env` (local) / Secret Manager (if GCP) — never committed.
- **Profile customization:** avatar defaults to the Google picture but the player can **upload a photo** to override it; the player also picks a **favorite-team flag** (ISO code) shown beside the nickname. Uploaded photos are size/type-validated and stored as a static asset (local volume now, object storage if on GCP).

## Consequences
- Lowest-friction onboarding; no password storage, reset flows, or brute-force surface.
- Requires a Google Cloud OAuth client (redirect URIs for local `http://localhost` + any deployed origin) — the available GCP project covers this.
- Need an allow-list / approval gate so a public Google login can't self-enroll into a private pool.
- Nickname is user-controlled → enforce uniqueness/length and keep the immutable `google_sub` as the real identity key.

## Implemented auth model (M2) — hardened
Three sign-in paths + a hardened security posture (from a security review + owner requirements):
- **Players — dev-login by nickname** (PoC, friends-trust): `POST /api/auth/dev-login {nickname}` creates/logs in a **player-only** account. It **refuses** to authenticate any nickname that maps to an `admin` role or a Google account (no impersonation of admins/Google users). Never issues an admin token.
- **Players — Google OAuth** (when `GOOGLE_OAUTH_CLIENT_ID/SECRET` set): the friendlier path; upserts by `google_sub`.
- **Admin — password**: `POST /api/auth/admin-login {password}` checked against `ADMIN_PASSWORD` (from `.env`) with a **constant-time compare**; disabled (503) if unset, 401 on mismatch (no enumeration). Admin is **not** derived from user count — `roleForNewUser` always returns `player`.
- **JWT — fail closed**: `JWT_SECRET` (≥32 bytes) is **required**; the server refuses to boot without it. No shipped default secret. HttpOnly cookie session.
- **Admin powers**: admin may set/edit a prediction **on behalf of any user** and edit after kickoff — every such action is written to the public audit (`admin_override` / on-behalf), actions-only.
- **Profile validation**: nickname (1–32, safe charset), `avatarUrl` (http/https + host, ≤512), `favoriteTeamCode` (must exist in `teams`).
- Secrets (`JWT_SECRET`, `ADMIN_PASSWORD`, `GOOGLE_OAUTH_*`, `TELEGRAM_*`) live in `.env` (gitignored), passed to the backend via `env_file`.

## Alternatives considered
- **Login + password + JWT for everyone** (earlier proposal) — more code (hashing, resets) and worse UX for players. Superseded; password kept only for admin.
- **Single shared access code + name pick** — weak isolation. Rejected.
- **First-user-becomes-admin by count** — privilege-escalation race; replaced by explicit admin password.
