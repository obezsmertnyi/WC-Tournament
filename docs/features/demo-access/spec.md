# Spec — Auth & demo-mode access (CAP-05)

How effective access is resolved and enforced. Owns: FR-030, FR-031, FR-032,
FR-033. Decision record: [ADR-0012](../../adr/0012-demo-mode-access-tiers.md).

## Boot / auth

### FR-030 — fail-closed secret
- **Given** the server is starting
- **When** `JWT_SECRET` is missing or shorter than 32 bytes
- **Then** the server refuses to boot (no insecure default).

### FR-030 — admin password
- **Given** an admin sign-in attempt
- **When** the password is compared
- **Then** the comparison is constant-time, and admin is never derived from user count.

## Effective access (`ComputeAccess`)

### FR-032 — demo off → everyone rw
- **Given** demo mode is **off**
- **When** any signed-in user's access is computed
- **Then** it resolves to `rw`.

### FR-032 — admin always rw
- **Given** demo mode is **on** and the user is an admin
- **When** access is computed
- **Then** it resolves to `rw`.

### FR-031 — demo on → stored tier applies
- **Given** demo mode is **on** and the user is a non-admin with tier `t`
- **When** access is computed
- **Then** it resolves to `t` (`none` / `ro` / `rw`).

## Gating (`DemoGate` middleware)

### FR-031 — read gating
- **Given** demo mode on and a `none` user
- **When** they call a "see others" route (leaderboard, audit, reveals, top-scorers)
- **Then** the request is rejected `403 demo_locked`; a `ro`/`rw` user passes.

### FR-031 — write gating
- **Given** demo mode on and a `ro` user
- **When** they call a participation route (submit prediction, bonus pick)
- **Then** the request is rejected `403 demo_locked`; only `rw` passes.

### FR-031 — anonymous falls through to 401
- **Given** a guarded route and no valid session
- **When** the request arrives
- **Then** the demo gate does not 403; `RequireUser` owns the `401`.

## Provisioning

### FR-033 — initial tier
- **Given** demo mode is **on**
- **When** a new user is created via Google self-service
- **Then** their initial tier is `none`; an admin-provisioned roster player is `rw`.
