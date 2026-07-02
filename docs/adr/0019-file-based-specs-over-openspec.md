# ADR-0019 — File-based specs over the OpenSpec CLI (and when to choose OpenSpec)

Status: Accepted (2026-07-02)

## Context

This repo practices spec-driven development (SDD): behavior is specified before
it is implemented, and every requirement is traced to a proving test. Our
implementation is plain files + our own scripts:

- **Specs:** `docs/features/<cap>/spec.md` (Given/When/Then), one per capability.
- **Requirement ids:** `docs/requirements.md` (FR/NFR/TC/BC grammar, ADR-0014).
- **Trace chain:** `@trace FR-id` comments in tests → generated matrix
  (`scripts/gen-traceability.mjs` → `docs/qa/requirements-traceability-matrix.md`)
  → non-regressing ratchets (`quality/*-baseline.json`) → CI + pre-commit gates.
- **Decisions:** `docs/adr/` (this folder).

[OpenSpec](https://github.com/Fission-AI/OpenSpec) is a lightweight open-source
SDD framework for AI coding assistants that standardizes the same discipline as
a CLI + directory convention: `openspec/specs/` is the source of truth for
current behavior, `openspec/changes/` holds change proposals (each a package of
`proposal.md`, `design.md`, `tasks.md`, and spec deltas), and slash commands
drive the loop (`/opsx:propose` → `/opsx:apply` → `/opsx:archive`). The question
this ADR answers: should this repo migrate to OpenSpec, and when should a new
project pick it?

## Decision

**Keep the file-based approach in this repo.** Do not migrate.

- Everything OpenSpec would enforce is already built and *gated* here: specs
  exist per capability, and `gen-traceability --check` validates the thing
  OpenSpec cannot — that each requirement is **proven by a test** and that
  coverage only ratchets up. `openspec validate` checks spec/change structure;
  our gate checks spec→test truth. For a live product, the second is stronger.
- Migration would touch 6 specs, the traceability generator, hooks, CI and the
  doc-graph — real risk and effort on a production app, for zero new capability.
- It would add a third-party CLI dependency that is young and evolving (schema
  profiles appeared in 1.2; formats may still shift).

This is a **per-project choice, not a verdict on OpenSpec**. The tool implements
the same philosophy we arrived at by hand, and for a fresh project it is the
faster way to get this discipline.

## When OpenSpec IS the right choice

Pick OpenSpec (over hand-rolling what this repo does) when:

1. **Greenfield.** Nothing to migrate — `openspec init` gives you the structure,
   agent instructions, and the propose→apply→archive loop on day one, instead of
   growing `docs/features/`, a requirements grammar, and scripts organically.
2. **A team (or several AI tools) needs a standard change protocol.** OpenSpec's
   change packages (`proposal.md` + `design.md` + `tasks.md` + spec deltas) make
   "agree on the spec BEFORE code" an explicit, reviewable artifact. In this solo
   repo, ADRs + PR review cover that; in a team they usually don't scale as well.
3. **Multiple assistants must share one brain.** It generates instructions for
   25+ tools (Claude Code, Copilot, Cursor, …) from one source, so every
   assistant sees the same specs — the same problem our `AGENTS.md` +
   `@AGENTS.md` import solves for two tools, solved generally.
4. **Brownfield you want to discipline incrementally.** It adopts iteratively —
   spec only the areas you're changing; no big-bang restructuring required.

Stay file-based (this repo's model) when the project is small/solo, already has
working specs+gates, or must avoid third-party workflow dependencies.

## How to use OpenSpec on a new project (quickstart)

```bash
npm install -g @fission-ai/openspec@latest
cd <project> && openspec init        # creates openspec/{specs/,changes/,project.md}
                                     # + slash commands for your AI tools
```

The loop per change:

1. `/opsx:explore` — (optional) think through options before committing to one.
2. `/opsx:propose "<idea>"` — the agent drafts a change package: proposal,
   design decisions, task breakdown, and spec deltas (how requirements change).
3. **Human reviews/edits the proposal** — this is the point of the method: align
   on WHAT before any code.
4. `/opsx:apply` — the agent implements the tasks against the agreed spec.
5. `/opsx:archive` — specs in `openspec/specs/` are updated to the new truth;
   the change package moves to `openspec/changes/archive/<timestamped>/`.

Maintenance: `openspec list` (open changes), `openspec validate` (structure
check — wire it into CI), `openspec update` (refresh agent instructions after
upgrades), `openspec config profile` (default vs expanded command set).

**Carry over from this repo regardless of tool:** OpenSpec does not do
requirement→test traceability. Keep the `@trace FR-id` convention + a generated
matrix + ratchets (see ADR-0014 and `scripts/gen-traceability.mjs`) on top of
whichever spec store you use — specs tell you what should be true; only traced
tests tell you it *is* true.

## Consequences

- This repo stays dependency-free for SDD; its `docs/features/` + trace chain
  remain the reference pattern for projects that hand-roll.
- New projects referencing this repo should read this ADR first and default to
  OpenSpec for greenfield/team setups, adding the trace chain on top.
- Revisit if this project grows beyond solo maintenance or OpenSpec stabilizes
  at 2.x with a migration tool worth its cost.

## Sources

- [OpenSpec repo](https://github.com/Fission-AI/OpenSpec) · [openspec.dev](https://openspec.dev/)
- [OpenSpec concepts](https://github.com/Fission-AI/OpenSpec/blob/main/docs/concepts.md)
