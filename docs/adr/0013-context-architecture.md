# ADR-0013: Context architecture — static vs dynamic, with a budget

**Status:** Accepted · **Date:** 2026-07-01

## Context
Agentic development needs the right context in the model at the right time.
Putting everything in one always-loaded file bloats the prompt and goes stale;
putting nothing means each session re-derives the project from scratch. We need
an explicit boundary between context that is stable (load always) and context
that changes per session/slice (load on resume).

## Decision
Split context into **static** and **dynamic** with a clear home for each:

- **Static** (stable, committed, always-loaded — keep lean, target ≲4k tokens):
  `AGENTS.md` (rules, pinned stack, guardrails — the **single source of truth**)
  + `docs/` SDD (product brief, requirements, capability plan, specs, ADRs,
  architecture). Changes here are deliberate and reviewed.
- **Dynamic** (changes often, loaded to resume): `docs/memory/activeContext.md`
  + `docs/memory/progress.md`, `current-state.md` (one-screen fast handoff),
  `CHECKLIST.md` (loop driver), `.workflow-state.toon` (loop state).

**Cross-tool, no drift:** the rules live in **`AGENTS.md`** (which Codex and
other tools read directly). `CLAUDE.md` is a one-liner that imports it
(`@AGENTS.md`), so Claude Code gets the same content; `.codex/config.toml`
documents the same for Codex. We deliberately do **not** maintain two full
rulesets — earlier, exporting a Claude project to Codex produced an `AGENTS.md`
that was just a *copy* of `CLAUDE.md` and immediately drifted; the import fixes that.

## Consequences
- The always-loaded budget stays small and current; session resume is fast
  (read `current-state.md` + memory).
- One source of truth for rules (no `AGENTS.md`/`CLAUDE.md` drift).
- Requires discipline: per-slice state goes to dynamic files, not `CLAUDE.md`.

## Alternatives considered
- **Single monolithic context file** — bloats the prompt, goes stale, mixes
  stable rules with transient state. Rejected.
- **Full `AGENTS.md` + full `CLAUDE.md` (two copies)** — duplicate rulesets
  drift (observed after a Codex export copied CLAUDE.md → AGENTS.md). Rejected in
  favor of one SSOT + an import.
- **CLAUDE.md as SSOT, AGENTS.md imports it** — Claude Code resolves `@imports`,
  but Codex may not reliably resolve an `@CLAUDE.md` inside AGENTS.md, so Codex
  could miss the rules. Rejected: the SSOT must be the file Codex reads natively
  (`AGENTS.md`).
