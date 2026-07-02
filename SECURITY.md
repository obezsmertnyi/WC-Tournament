# Security Policy

WC-Tournament is a private, friends-only prediction pool. It is not a public
service, but it handles auth, sessions, and third-party tokens, so it is built
and shipped with a real security posture.

## Reporting a vulnerability

Please **do not** open a public issue for a security problem. Report privately
via GitHub **Security Advisories** ("Report a vulnerability" on the repo's
Security tab), or contact the maintainer through GitHub
([@obezsmertnyi](https://github.com/obezsmertnyi)).
Expect an acknowledgement within a few days.

## Supported versions

The latest released `v0.x` tag and `main` are supported. Older tags are not
patched — upgrade to the latest release.

## Posture & controls

**Authentication & sessions**
- Sign-in via Google OAuth (OIDC); admin via a password checked in
  **constant time**, disabled if `ADMIN_PASSWORD` is unset. No admin role is
  derived from user count.
- Sessions are backend-issued **JWTs** (HS256) in an **HttpOnly** cookie;
  `JWT_SECRET` (≥32 bytes) is **required** — the server refuses to boot
  without it (fail-closed).

**Authorization**
- Private-pool gate: **demo mode + per-user access tiers** (`none`/`ro`/`rw`)
  so a published Google sign-in can't self-enroll and see others' data or
  participate until an admin grants access (ADR-0012). Enforced by a single
  server-side middleware; admins always resolve to full access.
- Admin actions (result overrides, on-behalf predictions, access grants,
  demo toggle) are written to a public, actions-only **audit log**.

**Secrets**
- `JWT_SECRET`, `ADMIN_PASSWORD`, `GOOGLE_OAUTH_*`, `TELEGRAM_*` live only in
  `.env` (gitignored) / the host environment — never committed. `.env.example`
  documents the shape with placeholders.
- CI runs **gitleaks** on the full history; a pre-commit `detect-private-key`
  + `gitleaks` hook catches leaks before they land.

**Supply chain & runtime**
- CI gates every change: `go vet`, race tests, `govulncheck`, frontend build,
  **Trivy** image scans (fail on CRITICAL/HIGH), and container smoke tests.
- Releases publish **multi-arch** images to GHCR with an **SBOM** (SPDX) per
  image attached to the GitHub Release. **Dependabot** keeps Go, npm, Actions,
  and Docker base images current.
- Backend runs as a non-root **distroless** container, read-only rootfs, all
  capabilities dropped, `no-new-privileges`.

**Data**
- Match data comes from the read-only public FIFA API; manual admin overrides
  are audited and take precedence only as an outage fallback (ADR-0006).
