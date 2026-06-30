# Architecture Decision Records

Each ADR captures one significant decision: its context, the choice, and the consequences. Format is a lightweight MADR. Status: `Proposed` → `Accepted` → `Superseded by ADR-XXXX`.

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-backend-go-gin-postgres.md) | Backend stack: Go + Gin + Postgres 17 | Accepted |
| [0002](0002-data-source-official-fifa-api.md) | Match data from the official FIFA API, providers pluggable | Accepted |
| [0003](0003-frontend-react-vite-typescript.md) | Frontend: React + Vite + TypeScript, production-grade | Accepted |
| [0004](0004-local-docker-compose-deployment.md) | Local-first deployment via docker-compose | Accepted |
| [0005](0005-auth-login-password-jwt.md) | Auth: Sign in with Google (OAuth) + profile nickname | Accepted |
| [0006](0006-manual-override-precedence.md) | API is source of truth; manual entry is an outage fallback | Accepted |
| [0007](0007-knockout-bracket-from-fifa.md) | Knockout qualifiers & bracket come from FIFA, not locally derived | Accepted |
| [0008](0008-scoring-deterministic-utc.md) | Scoring (exact 3 / outcome 1 / KO winner +1), penalties, UTC + Kyiv | Accepted |
| [0009](0009-public-audit-log.md) | Public, immutable audit log with timed disclosure | Accepted |
| [0010](0010-initial-backfill-window.md) | One-time audited backfill window for the opening matches | Accepted |
| [0011](0011-notifications-telegram.md) | Notifications via Telegram, friendly Ukrainian copy | Accepted |
| [0012](0012-demo-mode-access-tiers.md) | Demo mode with per-user access tiers (none/ro/rw) | Accepted |

New ADR: copy the structure of an existing one, take the next number, link it here.
