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
| [0013](0013-context-architecture.md) | Context architecture: static vs dynamic, with a budget | Accepted |
| [0014](0014-requirement-id-grammar.md) | Requirement-id grammar and the generated trace chain | Accepted |
| [0015](0015-mcp-read-only-tools.md) | Expose the pool to agents via a read-only MCP server | Accepted |
| [0016](0016-ai-recap-grounded-guardrail.md) | AI match recap — grounded generation behind a guardrail | Accepted |
| [0017](0017-ai-assistant-gemini-guardrail.md) | Football-only AI assistant — Gemini via keyless WIF, layered guardrail | Accepted |
| [0018](0018-ai-grounding-function-calling.md) | Ground the AI in live app data via function-calling (summary agent) | Accepted |
| [0019](0019-file-based-specs-over-openspec.md) | File-based specs over the OpenSpec CLI — and when to choose OpenSpec | Accepted |
| [0020](0020-regulation-score-from-fifa-periods.md) | Knockout regulation score from FIFA goal periods (ET-goal root fix) | Accepted |
| [0021](0021-design-first-ui-workflow.md) | Design-first UI workflow — design artifact before implementation | Accepted |
| [0022](0022-tournament-editions.md) | Multi-edition support — first-class `tournaments` entity, admin-managed | Accepted |

New ADR: copy the structure of an existing one, take the next number, link it here.
