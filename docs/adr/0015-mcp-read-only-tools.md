# ADR-0015: Expose the pool to agents via a read-only MCP server

**Status:** Accepted · **Date:** 2026-07-01

## Context
We want an agent (Claude, or any MCP client) to be able to answer questions about
the pool — "who leads?", "what are today's fixtures?", "show X's predictions" —
without scraping the UI or being handed credentials. The app already exposes a
read API; an MCP server is the agent-native way to surface it.

## Decision
Ship a small **read-only** MCP server (`mcp/`, TypeScript) that wraps the public
read API as tools: `list_fixtures`, `group_standings`, `leaderboard`, `bracket`,
`player_predictions`. Security posture, built to the `mcp-secure-server-dev`
discipline:

- **Read-only:** no tool mutates pool state — only GETs against the public API.
- **No secrets:** the server uses the public, unauthenticated read base URL
  (`WC_API_BASE`); it never reads or transmits `JWT_SECRET`/admin/OAuth/Telegram.
- **Validate every input:** tool args are schema-validated (zod); unknown params,
  wrong types, and out-of-range values (e.g. an over-long nickname, a non-date
  `date`) are rejected before any fetch; result sizes are bounded.
- **Reveal lock honored:** `player_predictions` returns only kicked-off matches,
  mirroring the app's reveal rule (no pre-kickoff leak).
- Tested (unit + a tiny integration against a stub) and **eval'd**
  (`mcp/evals/`) with golden fixtures + a rubric, traced to FR-070/FR-071.

## Consequences
- An agent can reason over live pool data with zero credentials and zero write
  risk; the blast radius is a read of already-public data.
- Adds a small Node service + `.mcp.json`; kept out of the prod request path
  (it's a developer/agent tool, not a user-facing endpoint).

## Alternatives considered
- **Give the agent the app's session/API token** — credential sprawl, write
  risk. Rejected.
- **A read+write MCP (let the agent predict on a user's behalf)** — turns a
  convenience tool into an attack surface; out of scope for a read-only bonus.
  Deferred.
