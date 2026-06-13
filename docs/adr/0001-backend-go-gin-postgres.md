# ADR-0001: Backend stack — Go + Gin + Postgres 17

**Status:** Accepted · **Date:** 2026-06-13

## Context
We need a backend for a friends-only WC2026 prediction pool: a REST API, a scoring engine, a scheduled results sync, and persistence. The team operates a Go-heavy stack at Everstake (claimer, sre_ai_agent) and is fluent in it. The tournament is already live, so familiarity and speed-to-ship matter.

## Decision
Backend in **Go 1.26** (latest) with **Gin** (HTTP) and **Postgres 17** (state), using `pgx` for data access — mirroring the existing `claimer` service conventions (numbered packages, `internal/...` layout, `cmd/server` entrypoint with CLI subcommands).

## Consequences
- Reuse of known patterns and tooling → faster delivery, lower ramp.
- Single static binary, trivial to containerize.
- Go's HTTP client + `golang.org/x/time/rate` fit the FIFA API polling needs (see ADR-0002).
- Trade-off: more boilerplate than a Node/Python equivalent for CRUD; accepted for consistency.

## Alternatives considered
- **Node/TypeScript** — shared language with the frontend, but no in-house operational consistency.
- **Python/FastAPI** — fastest for scraping, but we dropped scraping (ADR-0002) and prefer Go for the long-lived service.
