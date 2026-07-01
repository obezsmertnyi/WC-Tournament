# Spec — Tools / MCP server (CAP-08, bonus)

A read-only Model Context Protocol server that lets an agent (e.g. Claude) query
the public state of the pool. Owns: FR-070, FR-071 (FR-072 Future). Built
security-first (the `mcp-secure-server-dev` discipline): **read-only, no secrets,
validate every input**.

Decision record: [ADR-0015](../../adr/0015-mcp-read-only-tools.md).

## Tools

### FR-070 — fixtures
- **Given** an agent connected to the server
- **When** it calls `list_fixtures` with an optional `{ date?, stage? }`
- **Then** it receives the fixtures (teams, kickoff in Kyiv, status, score) for that day/stage from the live API.

### FR-070 — standings
- **When** the agent calls `group_standings { group? }`
- **Then** it receives the derived group table(s) (P/W/D/L/GF/GA/GD/Pts).

### FR-070 — leaderboard
- **When** the agent calls `leaderboard`
- **Then** it receives players ranked by points (match + bonus split), admins excluded.

### FR-070 — bracket
- **When** the agent calls `bracket`
- **Then** it receives the knockout tree in tree order (R32→final), with resolved/placeholder slots.

### FR-070 — a player's revealed predictions
- **When** the agent calls `player_predictions { nickname }`
- **Then** it receives that player's predictions **only for matches that have kicked off** (mirrors the app's reveal lock).

## Safety (FR-071)

### read-only
- **Given** any tool call
- **Then** the server performs only GET/read operations; there is no tool that mutates pool state.

### no secrets
- **Given** the server's responses and config
- **Then** no JWT/admin/OAuth/Telegram secret is ever read or returned; the server talks to the public read API only.

### input validation
- **Given** a tool call with an unknown parameter, wrong type, or out-of-range value (e.g. a 400-char nickname, a non-date `date`)
- **Then** the call is rejected with a clear validation error before any fetch; ranges are clamped/limited.

### FR-072 (Future) — per-match reveal
- A `match_reveal { matchId }` tool returns "who predicted what" **only after kickoff**, returning a locked marker before it.
