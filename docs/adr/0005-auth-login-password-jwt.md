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

## Alternatives considered
- **Login + password + JWT** (earlier proposal) — more code (hashing, rate-limit, resets) and worse UX. Superseded.
- **Single shared access code + name pick** — weak isolation, anyone edits anyone's predictions. Rejected.
