# WC-Tournament MCP server

A **read-only** Model Context Protocol server that lets an agent (Claude, or any
MCP client) query the public state of the pool. Spec:
[`docs/features/mcp/spec.md`](../docs/features/mcp/spec.md) · decision:
[ADR-0015](../docs/adr/0015-mcp-read-only-tools.md).

## Tools
| Tool | Args | Returns |
|------|------|---------|
| `list_fixtures` | `date?` (YYYY-MM-DD), `stage?` | fixtures (teams, kickoff, status, score) |
| `group_standings` | `group?` (A–L) | derived group table(s) |
| `leaderboard` | `limit?` (1–100) | players ranked by points (match+bonus) |
| `bracket` | — | knockout ties (R32→final) |
| `player_predictions` | `nickname`, `limit?` (1–50) | a player's picks, **kicked-off matches only** |

## Security posture (OWASP MCP)
- **Read-only** — only GETs; no tool mutates pool state.
- **No secrets** — talks to a configurable `WC_API_BASE`; never reads or forwards
  a JWT/admin/OAuth/Telegram secret (no token passthrough).
- **Schema-validated input** — every tool uses a strict zod schema; unknown
  params, wrong types, and out-of-range values are rejected *before* any fetch.
- **Reveal lock** — `player_predictions` only queries matches that have kicked
  off; the reveal endpoint returns a locked marker before kickoff.

## Run
```bash
cd mcp
npm install
npm run build           # → dist/server.js
WC_API_BASE=http://localhost:8080 node dist/server.js   # stdio MCP server
```
Then connect via the repo-root [`.mcp.json`](../.mcp.json) (Claude Code picks it
up automatically). `WC_API_BASE` must point at a read-reachable WC-Tournament API
(a local `docker compose up`); do **not** bake a session token into it.

## Verify
```bash
npm run typecheck       # tsc
npm run eval            # vitest golden-fixture evals (offline, no network/secrets)
```
Evals live in `evals/` and are traced to FR-070/FR-071 (see
`docs/qa/requirements-traceability-matrix.md`).
